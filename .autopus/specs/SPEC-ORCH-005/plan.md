# SPEC-ORCH-005 구현 계획

## 태스크 목록

- [ ] T1: `relay_pane.go` 생성 — 순차 pane relay 실행 엔진
- [ ] T2: `runner.go` 수정 — relay pane fallback 제거 및 pane 라우팅 통합
- [ ] T3: `pane_runner.go` 수정 — relay 전략을 pane orchestration에 통합
- [ ] T4: relay pane 전용 명령 빌더 구현 (인터랙티브 모드 인수)
- [ ] T5: `relay_pane_test.go` 생성 — 유닛 테스트
- [ ] T6: 통합 테스트 및 기존 전략 회귀 확인

## 구현 전략

### 접근 방법: 순차 Pane Runner 분리

기존 `pane_runner.go`는 병렬 실행 패턴(모든 pane 동시 생성 → 동시 대기)으로 설계되어 있다. relay의 순차 패턴은 근본적으로 다르므로 `relay_pane.go`로 분리하여 구현한다.

### T1: `relay_pane.go` 핵심 로직 (신규 파일)

**파일**: `pkg/orchestra/relay_pane.go`
**예상 크기**: 150-200줄

핵심 함수:
- `runRelayPane(ctx, cfg) ([]ProviderResponse, error)` — 순차 pane relay 진입점
  - jobID 생성 및 relay temp 디렉토리 생성
  - for loop으로 프로바이더 순차 실행:
    1. `SplitPane` → pane 생성
    2. `buildRelayPaneCommand` → 인터랙티브 명령 구성 (프롬프트 + 이전 결과 주입)
    3. `SendCommand` → pane에 명령 전송
    4. `waitForSentinel` → 완료 대기 (기존 함수 재사용)
    5. `readOutputFile` → 결과 수집 (기존 함수 재사용)
    6. 결과를 relay temp 파일에 저장
    7. `previous` 슬라이스에 결과 누적
  - 에러 핸들링: 실패 시 skip-continue (SPEC-ORCH-004 REQ-3a 패턴)
  - defer로 pane cleanup

- `buildRelayPaneCommand(provider, prompt, outputFile) string` — 인터랙티브 pane 명령 빌드
  - `-p` 플래그 제거
  - heredoc 또는 stdin으로 프롬프트 전달
  - output tee + sentinel 마커 기록
  - 기존 shellEscapeArg 보안 함수 활용

### T2: `runner.go` 수정

**변경 범위**: 3줄 삭제 (L28-31 fallback 블록)

현재 코드:
```go
if cfg.Strategy == StrategyRelay {
    fmt.Fprintf(os.Stderr, "relay pane mode not yet supported — using standard execution\n")
} else {
    return RunPaneOrchestra(ctx, cfg)
}
```

변경 후:
```go
return RunPaneOrchestra(ctx, cfg)
```

relay도 다른 전략과 동일하게 `RunPaneOrchestra`로 라우팅된다.

### T3: `pane_runner.go` 수정

**변경 범위**: ~10줄

`RunPaneOrchestra` 함수에 relay 전략 분기 추가:
```go
if cfg.Strategy == StrategyRelay {
    return runRelayPaneOrchestra(ctx, cfg)
}
```

relay는 병렬이 아닌 순차이므로, 기존 splitProviderPanes → sendPaneCommands → collectPaneResults 파이프라인을 타지 않고 `relay_pane.go`의 전용 함수로 분기한다.

`mergeByStrategy` 함수는 이미 relay 케이스를 처리하지 않으므로 (`StrategyRelay`가 switch에 없어 default consensus로 빠짐) — relay pane은 자체적으로 병합을 처리하므로 수정 불필요.

### T4: 인터랙티브 모드 명령 빌더

`buildRelayPaneCommand` 구현 상세:

**claude 인터랙티브**:
```bash
claude <<'PROMPT_EOF'
{injected prompt with previous results}
PROMPT_EOF
 | tee /tmp/output.txt; echo __AUTOPUS_DONE__ >> /tmp/output.txt
```

**codex 인터랙티브**:
```bash
codex <<'PROMPT_EOF'
{injected prompt with previous results}
PROMPT_EOF
 | tee /tmp/output.txt; echo __AUTOPUS_DONE__ >> /tmp/output.txt
```

**gemini 인터랙티브**:
```bash
gemini <<'PROMPT_EOF'
{injected prompt with previous results}
PROMPT_EOF
 | tee /tmp/output.txt; echo __AUTOPUS_DONE__ >> /tmp/output.txt
```

핵심: `-p` 플래그 없이 실행하여 CLI가 전체 TUI를 렌더링할 수 있게 한다. 프롬프트는 heredoc으로 전달.

### T5: 테스트

- relay pane 순차 실행 순서 검증
- 맥락 주입 내용 검증 (이전 결과가 다음 프롬프트에 포함)
- 프로바이더 실패 시 skip-continue 동작 검증
- pane cleanup 검증
- plain 터미널 fallback 검증 (standard relay로 실행)

### 기존 코드 활용

| 재사용 대상 | 위치 | 용도 |
|------------|------|------|
| `waitForSentinel` | `pane_runner.go:195` | sentinel 완료 감지 |
| `readOutputFile` | `pane_runner.go:229` | 출력 파일 읽기 + sentinel 제거 |
| `buildRelayPrompt` | `relay.go:88` | 이전 결과를 프롬프트에 주입 |
| `relayStageResult` | `relay.go:12` | 단계별 결과 구조체 |
| `cleanupRelayDir` | `relay.go:128` | relay temp 디렉토리 정리 |
| `shellEscapeArg` | `pane_shell.go:32` | 쉘 인젝션 방지 |
| `uniqueHeredocDelimiter` | `pane_shell.go:48` | heredoc 구분자 충돌 방지 |
| `sanitizeProviderName` | `pane_shell.go:13` | 파일명 안전 처리 |
| `randomHex` | `pane_runner.go:273` | jobID 생성 |

### 변경 범위 요약

| 파일 | 변경 유형 | 예상 줄 수 |
|------|----------|-----------|
| `relay_pane.go` | 신규 | 150-200 |
| `relay_pane_test.go` | 신규 | 150-200 |
| `runner.go` | 수정 | -3줄 |
| `pane_runner.go` | 수정 | +5줄 |
