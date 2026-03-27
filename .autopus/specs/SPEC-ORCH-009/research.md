# SPEC-ORCH-009 리서치

## 기존 코드 분석

### Root Cause 1: rounds=0 하드코딩

**파일**: `autopus-adk/internal/cli/orchestra_brainstorm.go:31`

```go
return runOrchestraCommand(cmd.Context(), "brainstorm", flagStrategy, flagProviders, timeout, judge, prompt, 0, noDetach, keepRelay)
//                                                                                                         ^ 하드코딩된 0
```

`resolveRounds()`는 `orchestra_helpers.go:71-79`에 이미 구현되어 있다:

```go
func resolveRounds(strategy string, rounds int) int {
    if rounds > 0 { return rounds }
    if strategy == "debate" { return 2 }
    return 0
}
```

`plan` 커맨드는 `orchestra.go:90`에서 올바르게 호출한다:

```go
resolvedRounds := resolveRounds(flagStrategy, rounds)
```

brainstorm만 누락. 수정은 plan 패턴을 그대로 복제하면 된다.

### Root Cause 2: Debate 라우팅 비활성화

**파일**: `autopus-adk/pkg/orchestra/interactive.go:26`

```go
if cfg.Strategy == StrategyDebate && cfg.DebateRounds >= 2 {
    return runInteractiveDebate(ctx, cfg)
}
```

rounds=0이 전달되면 `DebateRounds=0`이므로 이 조건이 항상 false. Root Cause 1이 수정되면 자동 해결.

### Root Cause 3: ReadScreen 출력 오염

**파일**: `autopus-adk/pkg/orchestra/interactive_detect.go:87-90`

```go
func cleanScreenOutput(raw string) string {
    cleaned := stripANSI(raw)
    return filterPromptLines(cleaned)
}
```

현재 `stripANSI()`는 기본 CSI 시퀀스(`\x1b[...m`)만 처리한다:

```go
var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
```

처리하지 못하는 패턴:
- OSC 시퀀스: `\x1b]...\x07` (윈도우 타이틀 등)
- DCS 시퀀스: `\x1bP...\x1b\\`
- 커서 위치 저장/복원: `\x1b7`, `\x1b8`
- tmux 상태바 라인

### 활성화될 기존 인프라 (변경 불필요)

| 파일 | 함수 | 역할 |
|------|------|------|
| `interactive_debate.go:14` | `runInteractiveDebate()` | 멀티턴 debate 진입점 |
| `interactive_debate.go:74` | `runPaneDebate()` | 터미널 pane 기반 debate 루프 |
| `interactive_debate.go:159` | `executeRound()` | 단일 라운드 실행 + 응답 수집 |
| `debate.go:95` | `buildRebuttalPrompt()` | 교차 응답 포함 rebuttal 프롬프트 생성 |
| `debate.go:108` | `buildJudgmentPrompt()` | judge 판정 프롬프트 생성 |
| `round_signal.go:14` | `RoundSignalName()` | 라운드별 시그널 파일명 생성 |
| `round_signal.go:21` | `CleanRoundSignals()` | 이전 라운드 시그널 정리 |
| `round_signal.go:33` | `SetRoundEnv()` | 라운드 환경변수 설정 |
| `interactive_debate_helpers.go:11` | `collectRoundHookResults()` | Hook 기반 라운드 결과 수집 |
| `interactive_debate_helpers.go:48` | `runJudgeRound()` | Judge 판정 라운드 실행 |
| `interactive_debate_helpers.go:86` | `consensusReached()` | 조기 합의 탐지 |

### 기존 테스트

| 파일 | 테스트 | 관련성 |
|------|--------|--------|
| `interactive_debate_test.go` | `TestRunInteractiveDebate_MultiRound_NoTerminal` | debate 루프 검증 |
| `runner_extra_test.go` | `TestRunDebate_WithRebuttalRound` | rebuttal 프롬프트 검증 |
| `round_signal_test.go` | 라운드 시그널 테스트 | 시그널 파일 생명주기 |
| `orchestra_brainstorm_test.go` | brainstorm 플래그 테스트 | --rounds 플래그 추가 시 영향 |

## 설계 결정

### D1: screen_sanitizer.go를 별도 파일로 분리

**결정**: 출력 정제 로직을 `interactive_detect.go`에 추가하지 않고 `screen_sanitizer.go`로 분리한다.

**이유**:
- `interactive_detect.go`는 이미 91줄이며, 정제 로직 추가 시 150줄+ 예상
- 정제 함수는 순수 함수이므로 독립적 테스트와 재사용이 용이
- `cleanScreenOutput()`은 기존 함수명을 유지하되 내부에서 `SanitizeScreenOutput()` 호출

**대안 검토**:
- `interactive_detect.go`에 인라인 추가 — 파일 크기 증가로 기각
- `pkg/sanitizer/` 별도 패키지 — 과도한 분리, orchestra 패키지 내부 용도이므로 기각

### D2: 기존 stripANSI() 유지

**결정**: 기존 `stripANSI()`와 `filterPromptLines()`는 그대로 유지하고, `SanitizeScreenOutput()`을 새로 추가한다.

**이유**:
- `isPromptLine()`, `isPromptVisible()` 등이 기존 함수에 의존
- 하위 호환성 보장
- `cleanScreenOutput()`만 내부 구현을 교체

### D3: Brainstorm --rounds 플래그는 plan 패턴 복제

**결정**: `plan` 커맨드의 `--rounds` 플래그 구현 패턴을 brainstorm에 그대로 복제한다.

**이유**:
- 코드 일관성 (같은 프로젝트 내 동일 패턴)
- `resolveRounds()`가 이미 debate/non-debate 분기를 처리
- validation 로직은 `runOrchestraCommand`에서 이미 처리 (`orchestra.go:194-199`)

### D4: 프로바이더 어댑터는 문서화만 (P2)

**결정**: OpenCode 프롬프트 전달 문제는 본 SPEC에서 분석/문서화만 하고, 실제 구현은 별도 SPEC으로 분리한다.

**이유**:
- OpenCode의 `run` 모드는 non-interactive 전용이며, interactive 모드에서는 `buildInteractiveLaunchCmd()`가 `run` 플래그를 올바르게 스킵(`interactive.go:283`)
- 프로바이더별 어댑터 인터페이스는 설계 범위가 크므로 별도 SPEC 필요
- 현재 `PromptViaArgs` 플래그로 기본 분기는 가능

### 프로바이더별 프롬프트 전달 분석

| 프로바이더 | Interactive 모드 | 프롬프트 전달 | 알려진 이슈 |
|-----------|-----------------|-------------|------------|
| claude | TUI 세션 시작 후 SendCommand | `SendCommand` → Enter | 정상 동작 |
| codex | TUI 세션 시작 후 SendCommand | `SendCommand` → Enter | `-q` 플래그 스킵 필요 |
| gemini | TUI 세션 시작 후 SendCommand | `SendCommand` → Enter | 정상 동작 |
| opencode | TUI 세션 시작 (`run` 스킵) | `SendCommand` → Enter | `run` 모드 미사용 시 정상, 프롬프트 길이 제한 가능성 |

`buildInteractiveLaunchCmd()` (`interactive.go:279-289`) 분석:
- `run` 플래그를 스킵하므로 opencode는 interactive TUI로 실행됨
- `-p`, `-q` (print/quiet) 플래그도 스킵 — interactive 모드에서는 불필요
- 현재 구현은 올바르나, 프로바이더별 세밀한 제어가 필요할 경우 어댑터 패턴 도입 검토
