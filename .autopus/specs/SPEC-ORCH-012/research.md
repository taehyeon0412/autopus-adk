# SPEC-ORCH-012 리서치

## 기존 코드 분석

### 버그 경로 1: cmux SendLongText

**파일**: `pkg/terminal/cmux.go:70-74`
```go
func (a *CmuxAdapter) SendLongText(ctx context.Context, paneID PaneID, text string) error {
    return a.SendCommand(ctx, paneID, text)
}
```

`SendCommand` (`cmux.go:59-67`)는 `exec.Command("cmux", "send", "--surface", ref, text)`를 실행한다. Go의 `exec.Command`는 shell을 bypass하므로 shell 메타문자 문제는 없지만, OS의 `execve()` argument 크기 제한(macOS: `ARG_MAX` ~262144, 하지만 단일 arg 제한은 더 낮을 수 있음)과 cmux 내부 처리 제한에 걸린다.

**비교 — tmux 어댑터**: `pkg/terminal/tmux.go:121-163`
- 500B 미만: `send-keys` (Enter 없이)
- 500B 이상: temp file → `load-buffer` → `paste-buffer` → cleanup
- cmux에는 이 분기가 전혀 없음

### 버그 경로 2: launch command의 SendCommand 전달

**파일**: `pkg/orchestra/interactive.go:134-153`
```go
func launchInteractiveSessions(ctx context.Context, cfg OrchestraConfig, panes []paneInfo) []FailedProvider {
    // ...
    cmd := buildInteractiveLaunchCmd(pi.provider, launchPrompt) + "\n"
    if err := cfg.Terminal.SendCommand(ctx, pi.paneID, cmd); err != nil {
```

**파일**: `pkg/orchestra/interactive_launch.go:36-38`
```go
if p.InteractiveInput == "args" && prompt != "" {
    cmd += " " + shellQuote(prompt)
}
```

opencode의 args 모드에서 프롬프트가 launch command에 포함되어 `SendCommand`로 전달된다. 긴 프롬프트일 경우 truncation 발생.

### cmux CLI 확인

- `cmux set-buffer --name <name> <text>`: 10K+ 문자 정상 처리 확인
- `cmux paste-buffer --name <name> --surface <ref>`: surface에 버퍼 내용 붙여넣기
- `cmux delete-buffer --name <name>`: 버퍼 정리 (best-effort)
- Go `exec.Command`는 shell bypass이므로 `set-buffer`의 text argument에 한글/특수문자 안전

### 호출 지점 분석

`SendLongText` 호출 지점:
1. `interactive.go:203` — `sendPrompts()`: non-args 프로바이더의 프롬프트 전달 (이미 SendLongText)
2. `interactive_debate.go:191` — debate round의 rebuttal 프롬프트 전달 (이미 SendLongText)

`SendCommand`로 긴 텍스트를 전달하는 위험 지점:
1. `interactive.go:144` — `launchInteractiveSessions()`: **이번 수정 대상**
2. `interactive_debate_helpers.go:57` — judge judgment 전달: judgment은 짧으므로 OOS

### Terminal 인터페이스 계약

**파일**: `pkg/terminal/terminal.go:38-41`
```go
SendLongText(ctx context.Context, paneID PaneID, text string) error
```
- "Callers must send Enter separately after this call" — 이 계약이 launch command 수정의 근거

### 테스트 인프라

**파일**: `pkg/terminal/cmux_test.go:45-62` — `newCmuxMockV2`: 모든 exec 호출 기록, output/error 설정 가능
**파일**: `pkg/orchestra/pane_mock_test.go:54-70` — `mockTerminal`: SendLongText → SendCommand 위임

mock에 호출 순서 추적 기능 추가 필요 (set-buffer → paste-buffer → delete-buffer 순서 검증).

## 설계 결정

### D1: set-buffer 직접 사용 vs temp file 경유

**선택**: `cmux set-buffer --name <name> <text>`를 직접 사용
**이유**: cmux는 `set-buffer`에 text를 직접 CLI arg로 받으며, Go `exec.Command`가 shell을 bypass하므로 특수문자 안전. tmux의 temp file 경로는 `load-buffer`가 파일 경로만 받기 때문이며, cmux는 이 제약이 없다.
**대안**: temp file + `set-buffer --file` — 불필요한 I/O 오버헤드

**리뷰 피드백 반영 (gemini critical)**:
gemini가 "CLI arg도 OS ARG_MAX 제한에 걸린다"고 지적. 이론적으로 맞지만 실측 결과:
- `cmux set-buffer` CLI arg 방식으로 500KB(500,000자) 정상 동작 확인
- `cmux set-buffer` CLI arg 방식으로 80KB(한글 5000회 반복) 정상 동작 확인
- macOS ARG_MAX = ~1,048,576 bytes
- Orchestra 프롬프트 최대 크기: rebuttal(3 providers × 평균 2KB) + topic isolation ≈ 10KB

**실제 truncation 원인**: `cmux send`가 PTY 입력 버퍼(~4KB)를 통해 전달하여 잘림 발생. `set-buffer`는 PTY를 bypass하여 cmux 내부 메모리에 직접 저장하므로 이 한계에 걸리지 않음.

**결론**: CLI arg 방식으로 충분. 만약 향후 100KB+ 프롬프트가 필요하면 stdin 경로(`cmux set-buffer --name <name> -`)를 추가할 수 있으나, 현재 Out of Scope.

### D2: 500B 임계값 유지

**선택**: tmux와 동일한 500B 임계값
**이유**: 일관성. 두 어댑터 모두 같은 기준으로 buffer 경로를 선택하면 동작 예측이 쉽다.
**대안**: cmux는 CLI arg 제한이 다르므로 다른 임계값 가능하지만, 안전 마진 확보를 위해 동일 값 유지.

### D3: launch command도 SendLongText로 통합

**선택**: `launchInteractiveSessions`에서 command 본문을 `SendLongText`로, Enter를 별도 `SendCommand`로 전달
**이유**: Terminal 인터페이스의 계약("Callers must send Enter separately")을 준수. args 프로바이더의 긴 프롬프트 포함 launch command가 안전하게 전달됨.
**대안**: launch command만 별도 처리 — 불필요한 분기, SendLongText가 이미 short/long 분기를 처리함.

### D4: unique 버퍼 이름 전략

**선택**: `autopus-<sanitized-paneID>-<unix-nano>` 형식
**이유**: paneID(surface:7)에서 콜론을 제거하고 nanosecond timestamp를 결합하면 병렬 실행에서도 충돌 없음.
**대안**: UUID — 외부 의존성(crypto/rand) 추가 불필요, timestamp로 충분.
