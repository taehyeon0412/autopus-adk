# SPEC-ORCH-014 리서치

## 기존 코드 분석

### 1. Provider Config 구조체 (`pkg/orchestra/types.go:32-39`)

```go
type ProviderConfig struct {
    Name             string
    Binary           string
    Args             []string   // non-interactive mode args
    PaneArgs         []string   // pane mode args (overrides Args)
    PromptViaArgs    bool       // prompt as last arg
    InteractiveInput string     // "args" = via CLI arg at launch, "" = via sendkeys
}
```

`InteractiveInput`이 핵심 스위치. `"args"`이면 launch 시 prompt를 CLI 인자로 전달하고
round 1에서 SendLongText 스킵. 빈 문자열이면 SendLongText로 프롬프트 전달.

### 2. Interactive Launch (`pkg/orchestra/interactive_launch.go:16-43`)

`buildInteractiveLaunchCmd()` 함수:
- PaneArgs를 순회하며 `--print`, `-p`, `--quiet`, `-q` 플래그 스킵
- `InteractiveInput != "args"`이면 `run` 서브커맨드도 스킵
- `InteractiveInput == "args"` && prompt 있으면 → prompt를 shellQuote로 마지막 인자에 추가

opencode를 TUI 모드로 전환하면 (`InteractiveInput` 빈 문자열):
- PaneArgs에서 `run` 자동 스킵됨 (line 24)
- prompt는 launch 시 포함되지 않음
- 결과: `opencode -m openai/gpt-5.4`

### 3. executeRound에서의 args 스킵 (`pkg/orchestra/interactive_debate.go:201`)

```go
if pi.provider.InteractiveInput == "args" && round == 1 {
    continue  // SendLongText 스킵
}
```

opencode가 `InteractiveInput == ""`이 되면 이 조건이 false → round 1에서도 SendLongText 전달.

### 4. sendPrompts에서의 args 스킵 (`pkg/orchestra/interactive.go:187`)

```go
if pi.provider.InteractiveInput == "args" {
    continue
}
```

단일 라운드(non-debate) interactive에서도 동일한 스킵 로직. TUI 모드 전환 시 자동 해제.

### 5. Hook Signal 프로토콜 (`pkg/orchestra/hook_signal.go:26-30`)

```go
var defaultHookProviders = map[string]bool{
    "claude":   true,
    "gemini":   true,
    "opencode": true,  // 이미 등록됨
}
```

opencode는 이미 hook provider로 등록. `hook-opencode-complete.ts`가 `text.complete`
이벤트에서 round-scoped 시그널 파일 생성.

### 6. Completion Pattern (`pkg/orchestra/types.go:100`)

```go
{Provider: "opencode", Pattern: regexp.MustCompile(`(?m)^>\s*$`)},
```

opencode TUI 프롬프트 패턴 이미 등록. screen polling fallback에서 사용 가능.

### 7. Default Config (`pkg/config/defaults.go:68`)

```go
"opencode": {Binary: "opencode", Args: []string{"run", "-m", "openai/gpt-5.4"},
             PaneArgs: []string{"run", "-m", "openai/gpt-5.4"}, PromptViaArgs: true},
```

변경 필요: PaneArgs에서 `run` 제거, `InteractiveInput` 빈 문자열 유지 (기본값이므로 변경 불필요).

### 8. Migration Config (`pkg/config/migrate.go:10`)

```go
"opencode": {Binary: "opencode", Args: []string{"run", "-m", "openai/gpt-5.4"},
             PaneArgs: []string{"run", "-m", "openai/gpt-5.4"}, PromptViaArgs: true},
```

`defaultProviderEntries`에서도 PaneArgs 갱신 필요.

### 9. InjectOrchestraPlugin (`pkg/adapter/opencode/opencode.go:95-132`)

opencode.json에 `autopus-result` 플러그인 등록 함수 이미 구현. `text.complete` 이벤트에
`bun {scriptPath}` 실행 등록. orchestra 실행 전에 호출되어야 함.

### 10. autopus.yaml 현재 상태

```yaml
opencode:
    binary: opencode
    args: [run, -m, "openai/gpt-5.4"]
    prompt_via_args: true
```

`pane_args`와 `interactive_input` 미설정. 현재는 Args가 PaneArgs 없을 때 fallback으로 사용됨.

## 설계 결정

### 왜 InteractiveInput 제거인가

**결정**: opencode의 `InteractiveInput`을 `"args"`에서 빈 문자열(기본값)로 변경.

**근거**:
1. opencode TUI가 cmux pane에서 정상 실행됨 (검증 완료)
2. paste-buffer로 프롬프트 전달 성공 (검증 완료)
3. TUI 세션 유지로 멀티턴 가능 (검증 완료)
4. claude/gemini와 동일한 코드 경로 사용 → 유지보수 비용 감소

**대안 검토**:
- A) 새 InteractiveInput 모드 추가 (예: `"tui"`) — 불필요한 복잡성. 기존 빈 문자열이 정확히 원하는 동작
- B) opencode 전용 launch 로직 — 코드 분기 증가. 기존 인프라가 충분히 지원

### 왜 PaneArgs와 Args 분리인가

**결정**: Args는 `[run, -m, openai/gpt-5.4]` 유지, PaneArgs는 `[-m, openai/gpt-5.4]`로 변경.

**근거**:
- Args: 비인터랙티브 실행 (`opencode run -m ... 'prompt'`) → `run` 필요
- PaneArgs: TUI 인터랙티브 실행 (`opencode -m ...`) → `run` 불필요
- 이 분리가 PaneArgs 필드의 존재 이유
