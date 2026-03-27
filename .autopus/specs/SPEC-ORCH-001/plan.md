# SPEC-ORCH-001 구현 계획

## 태스크 목록

- [x] T1: `OrchestraConfig`에 `Terminal` 필드 추가 (`pkg/orchestra/types.go`)
- [x] T2: `pane_runner.go` 신규 생성 — pane 기반 병렬 실행 엔진 (`pkg/orchestra/pane_runner.go`)
- [x] T3: `runner.go`에 cmux 모드 분기 로직 추가 — `runParallelWithPanes()` 위임 (`pkg/orchestra/runner.go`)
- [x] T4: CLI 레이어에서 `DetectTerminal()` 호출 및 config 주입 (`internal/cli/orchestra.go`)
- [x] T5: `pane_runner_test.go` 단위 테스트 작성 (`pkg/orchestra/pane_runner_test.go`)
- [x] T6: 기존 테스트 회귀 검증 — Terminal=nil 시 기존 동작 유지 확인

## 구현 전략

### 접근 방법: Terminal 인터페이스 주입 + 조건부 분기

기존 `pkg/terminal/` 패키지의 `Terminal` 인터페이스와 `DetectTerminal()`을 그대로 재활용한다. `OrchestraConfig`에 optional `Terminal` 필드를 추가하고, 이 필드가 nil이 아니고 `Name() != "plain"`이면 pane 모드로 전환한다.

### pane 기반 실행 흐름

```
1. CLI에서 DetectTerminal() 호출 → OrchestraConfig.Terminal에 주입
2. RunOrchestra()에서 Terminal 확인
3. cmux/tmux 사용 가능 → runParallelWithPanes() 호출
4. 각 프로바이더마다:
   a. SplitPane(Horizontal) → PaneID 획득
   b. 임시 출력 파일 생성 (/tmp/autopus-orch-{provider}-{uuid}.out)
   c. SendCommand(paneID, "{binary} {interactive_args} | tee {output_file}; echo __AUTOPUS_DONE__ >> {output_file}")
   d. 완료 대기 (sentinel 파일 감시 또는 polling)
5. 모든 프로바이더 완료 → 출력 파일에서 결과 읽기
6. 기존 merge 로직에 ProviderResponse 전달
7. pane 정리 (Close)
```

### 기존 코드 활용

- `pkg/terminal/terminal.go` — Terminal 인터페이스 (SplitPane, SendCommand, Close)
- `pkg/terminal/cmux.go` — CmuxAdapter (이미 구현됨)
- `pkg/terminal/detect.go` — DetectTerminal() (이미 구현됨)
- `pkg/orchestra/runner.go` — runProvider() 결과 포맷을 pane 결과에서 재구성
- `pkg/orchestra/debate.go`, `merger.go` — merge 로직 그대로 재활용

### 변경 범위 최소화

- `types.go`: Terminal 필드 1줄 추가
- `runner.go`: 분기 조건 5-10줄 추가
- `orchestra.go` (CLI): DetectTerminal() 호출 3-5줄 추가
- `pane_runner.go`: 신규 파일 (~150줄)
- `pane_runner_test.go`: 신규 파일 (~200줄)

### 인터랙티브 프로바이더 설정

pane 모드에서는 기존 ProviderConfig의 Args에서 `-p`, `-q` 등 비인터랙티브 플래그를 제거하고, 인터랙티브 모드로 실행한다. 이 변환은 pane_runner 내부에서 수행한다.

| Provider | 기존 (non-interactive) | pane 모드 (interactive) |
|----------|----------------------|------------------------|
| claude   | `claude -p`          | `claude`               |
| codex    | `codex -q`           | `codex`                |
| gemini   | `gemini -p "prompt"` | `gemini`               |
