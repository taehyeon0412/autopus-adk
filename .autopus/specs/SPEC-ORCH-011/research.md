# SPEC-ORCH-011 리서치

## 기존 코드 분석

### 근본 원인 1: opencode 프롬프트 미전달

**파일**: `pkg/orchestra/interactive_launch.go:17-33` (`buildInteractiveLaunchCmd`)

```go
// line 21: "run" 플래그가 필터링됨
if arg == "--print" || arg == "-p" || arg == "--quiet" || arg == "-q" || arg == "run" {
    continue
}
```

opencode의 PaneArgs는 `["run", "-m", "openai/gpt-5.4"]`이지만, `buildInteractiveLaunchCmd`에서 `"run"` 플래그를 제거한다. 결과적으로 launch 명령은 `opencode -m openai/gpt-5.4`가 되며, 이는 opencode의 TUI 인터랙티브 모드를 실행한다.

TUI 모드에서는 자체 입력 위젯이 키 입력을 처리하므로, tmux `send-keys`로 전달된 텍스트가 opencode의 입력 버퍼에 도달하지 않거나 부분적으로만 도달한다. 결과적으로 opencode는 프롬프트 없이 시작하여 "안녕하세요. 무엇을 도와드릴까요?" 같은 기본 인사만 표시한다.

**해결 방향**:
- opencode는 interactive TUI 대신 `opencode run -m <model> "<prompt>"` 형태로 non-interactive 실행
- `buildInteractiveLaunchCmd`에서 opencode를 특별 처리하거나, `InteractiveInput: "args"` 설정으로 분기
- 또는 launch 시 `run` 플래그를 유지하되, 프롬프트를 마지막 인자로 포함

### 근본 원인 2: gemini 프롬프트 truncation

**파일**: `pkg/terminal/tmux.go:59-68` (`TmuxAdapter.SendCommand`)

```go
func (a *TmuxAdapter) SendCommand(_ context.Context, paneID PaneID, command string) error {
    target := a.session + ":" + string(paneID)
    cmd := execCommand("tmux", "send-keys", "-t", target, command, "Enter")
    ...
}
```

`tmux send-keys`는 커맨드라인 인자로 텍스트를 전달한다. OS의 ARG_MAX 제한(macOS: ~262144바이트)보다 훨씬 아래이지만, tmux 내부적으로 send-keys는 텍스트를 키 이벤트로 변환하여 전달한다. 긴 텍스트는:
1. 터미널 라인 에디터 버퍼 오버플로우
2. tmux의 내부 키 이벤트 큐 제한
3. 대상 프로세스(gemini)의 입력 버퍼 제한

으로 인해 잘릴 수 있다.

추가로, `sendPrompts()`에서 프롬프트 텍스트와 Enter를 별도로 보내는데 (`interactive.go:195-206`), 500ms 딜레이 사이에 터미널이 입력을 분할 처리할 수 있다.

**파일**: `pkg/terminal/cmux.go:59-68` (`CmuxAdapter.SendCommand`)

```go
func (a *CmuxAdapter) SendCommand(_ context.Context, paneID PaneID, command string) error {
    cmd := execCommand("cmux", "send", "--surface", string(paneID), command)
    ...
}
```

cmux는 커맨드라인 인자로 전달하므로 동일한 제한이 적용될 수 있다.

**해결 방향**:
- tmux: `load-buffer` + `paste-buffer` 조합 사용 (tmux 공식 권장 방식 for long text)
  ```
  tmux load-buffer /tmp/prompt.txt
  tmux paste-buffer -t session:pane
  ```
- cmux: `cmux send --surface <ref> --file <path>` 옵션이 있는지 확인, 없으면 stdin 파이프 사용

### sendPrompts() 흐름 분석

**파일**: `pkg/orchestra/interactive.go:188-215`

현재 흐름:
1. `SendCommand(ctx, paneID, cfg.Prompt)` — 프롬프트 텍스트 전송
2. `time.Sleep(500ms)` — CLI가 paste를 등록할 시간
3. `SendCommand(ctx, paneID, "\n")` — Enter 전송으로 제출

문제점:
- 프롬프트가 길면 step 1에서 이미 잘림
- opencode TUI는 step 1의 send-keys를 자체 입력 위젯에서 처리 못함

### executeRound() 흐름 분석

**파일**: `pkg/orchestra/interactive_debate.go:159-203`

현재 흐름:
1. round > 1이면 이전 라운드 신호 정리 + 프롬프트 대기
2. `cfg.Terminal.SendCommand(ctx, pi.paneID, prompt)` — 프롬프트 전송
3. `time.Sleep(500ms)`
4. `cfg.Terminal.SendCommand(ctx, pi.paneID, "\n")` — Enter 전송

debate 라운드의 rebuttal 프롬프트는 `topicIsolationInstruction + buildRebuttalPrompt(...)` 형태로, 이전 라운드의 전체 응답이 포함되어 매우 길어진다 (3000-5000자+). truncation 문제가 더 심각하게 나타난다.

### Terminal 인터페이스 분석

**파일**: `pkg/terminal/terminal.go` (확인 필요)

현재 `Terminal` 인터페이스:
- `SendCommand(ctx, paneID, command) error` — 텍스트 전송
- `ReadScreen(ctx, paneID, opts) (string, error)` — 화면 읽기
- 기타: `CreateWorkspace`, `SplitPane`, `Close`, `PipePaneStart`, `PipePaneStop`, `Notify`

`SendLongText` 같은 긴 텍스트 전용 메서드가 없다.

## 설계 결정

### D1: 긴 텍스트 전달 — load-buffer/paste-buffer vs 청크 분할

**선택**: load-buffer/paste-buffer (tmux), 파일 경유 (cmux)

**이유**:
- 청크 분할은 CLI가 각 청크를 별도 입력으로 해석할 위험이 있음
- `load-buffer`/`paste-buffer`는 tmux 공식 메커니즘으로, 텍스트 크기 제한이 사실상 없음
- 한 번에 전체 텍스트를 paste하므로 CLI 입장에서 단일 입력으로 처리됨

**대안**:
1. ~~청크 분할 (500B 단위)~~: CLI가 각 청크를 별도 명령으로 해석할 수 있어 위험
2. ~~파이프 기반 stdin redirect~~: interactive pane에서는 이미 CLI가 실행 중이므로 stdin redirect 불가

### D2: opencode 처리 — TUI 모드 vs run 모드

**선택**: opencode 전용 분기 — interactive pane에서도 `run` 서브커맨드 유지

**이유**:
- opencode TUI는 send-keys 호환성이 불확실
- `opencode run -m <model> "prompt"`는 non-interactive 모드로 프롬프트를 직접 인자에 포함
- pane에서 출력을 캡처하면서도 프롬프트 전달 문제를 근본적으로 해결

**대안**:
1. ~~opencode TUI에 send-keys로 전달~~: TUI 입력 위젯 호환성 문제로 불안정
2. ~~opencode를 interactive 모드에서 제외~~: 멀티프로바이더 목적에 부합하지 않음

### D3: 인터페이스 확장 — SendLongText vs SendCommand 수정

**선택**: 새 메서드 `SendLongText` 추가

**이유**:
- `SendCommand`는 짧은 명령어(binary launch, Enter 등)에 적합
- 긴 텍스트 전달은 구현이 다르므로(임시 파일 경유) 별도 메서드가 명확
- 기존 `SendCommand` 호출자에 영향 없음 (하위 호환)

**대안**:
1. ~~SendCommand에 길이 분기 추가~~: 인터페이스 계약이 암시적으로 변경되어 혼란
2. ~~별도 SendFile 메서드~~: 텍스트를 파일로 보내는 것은 구현 세부사항이지 인터페이스가 아님
