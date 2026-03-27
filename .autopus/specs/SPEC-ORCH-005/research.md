# SPEC-ORCH-005 리서치

## 기존 코드 분석

### 1. relay 전략 현재 구현 (`pkg/orchestra/relay.go`)

- `runRelay(ctx, cfg)` — standard execution 진입점 (L22-85)
  - `randomHex()` 2번 호출로 jobID 생성
  - `/tmp/autopus-relay-{jobID}/` 디렉토리 생성
  - for loop으로 프로바이더 순차 실행 (L35-70)
  - `agenticArgs(providerName)` — 프로바이더별 agentic 플래그 추가 (L105-115)
  - `buildRelayPrompt(original, previousResults)` — 이전 결과를 프롬프트에 주입 (L88-101)
  - `runProvider(ctx, provider, prompt)` — 단일 프로바이더 실행 (`runner.go:234`)
  - 실패 시 skip-continue 패턴 (L43-57)

핵심 재사용: `buildRelayPrompt`, `relayStageResult`, `cleanupRelayDir`, `FormatRelay`는 pane 모드에서도 동일하게 사용 가능.

### 2. 병렬 pane runner (`pkg/orchestra/pane_runner.go`)

- `RunPaneOrchestra(ctx, cfg)` — pane 실행 진입점 (L30-75)
  - plain 터미널이면 `RunOrchestra`로 fallback (L33-34)
  - `splitProviderPanes` — 모든 pane 동시 생성 (L79-99)
  - `sendPaneCommands` — 모든 pane에 동시 명령 전송 (L103-116)
  - `collectPaneResults` — goroutine으로 병렬 sentinel 대기 (L120-163)
  - `mergeByStrategy` — 전략별 결과 병합 (L239-253)

핵심 차이: 병렬 pane runner는 모든 pane을 동시에 열고 goroutine으로 대기하지만, relay pane은 순차적으로 1개씩 열고 대기해야 한다.

### 3. Pane 명령 빌더 (`pkg/orchestra/pane_runner.go:176-192`)

- `buildPaneCommand(provider, prompt, outputFile)` — 쉘 명령 구성
  - `paneArgs(provider)` — PaneArgs 우선, 없으면 Args (L167-172)
  - SEC-001: `shellEscapeArg`로 프롬프트 이스케이프
  - SEC-004: `shellEscapeArgs`로 인수 이스케이프
  - SEC-006: 바이너리 경로 이스케이프
  - heredoc 방식 vs args 방식 분기 (`PromptViaArgs`)

relay pane에서 재사용할 보안 함수: `shellEscapeArg`, `shellEscapeArgs`, `uniqueHeredocDelimiter` (모두 `pane_shell.go`)

### 4. runner.go의 relay pane fallback (`pkg/orchestra/runner.go:27-34`)

```go
if cfg.Terminal != nil && cfg.Terminal.Name() != "plain" {
    if cfg.Strategy == StrategyRelay {
        fmt.Fprintf(os.Stderr, "relay pane mode not yet supported — using standard execution\n")
    } else {
        return RunPaneOrchestra(ctx, cfg)
    }
}
```

이 분기를 제거하면 relay도 `RunPaneOrchestra`로 진입한다. `RunPaneOrchestra` 내부에서 relay 전략을 감지하여 순차 실행으로 분기해야 한다.

### 5. Terminal 인터페이스 (`pkg/terminal/terminal.go`)

- `SplitPane(ctx, direction) (PaneID, error)` — pane 생성
- `SendCommand(ctx, paneID, cmd) error` — 명령 전송
- `Close(ctx, name) error` — pane 닫기

순차 relay에서 사용하는 Terminal 메서드는 병렬과 동일하지만, 호출 패턴이 다르다 (동시가 아닌 순차).

### 6. ProviderConfig의 PaneArgs (`pkg/orchestra/types.go:35`)

```go
PaneArgs []string // args for pane mode (overrides Args when set)
```

현재 hardcoded 기본값 (`internal/cli/orchestra.go:237-241`):
- claude: `PaneArgs: []string{"-p"}` — relay pane에서는 `-p` 없이 실행해야 함
- codex: `PaneArgs: []string{"-q"}` — relay pane에서는 `-q` 없이 실행해야 함
- gemini: `PaneArgs: []string{"-p"}` — relay pane에서는 `-p` 없이 실행해야 함

relay pane 전용으로 `PaneArgs`를 비워야 하거나, 별도 필드 `RelayPaneArgs`를 추가해야 한다.

## 설계 결정

### D1: 별도 파일 분리 vs pane_runner.go 확장

**결정**: `relay_pane.go`로 분리

**이유**:
- `pane_runner.go`는 현재 280줄. relay 로직 추가 시 300줄 초과 (file-size-limit 위반)
- 병렬과 순차는 제어 흐름이 근본적으로 다름. 같은 파일에 두면 복잡도가 급증
- 독립 파일로 유지하면 테스트도 분리 가능

**대안 검토**:
- pane_runner.go 내부에 분기 추가 — file size 초과, 복잡도 증가로 기각

### D2: 인터랙티브 인수 결정 방식

**결정**: relay pane에서는 `buildRelayPaneCommand` 전용 함수에서 `-p`/`-q` 플래그를 제외하고 명령을 구성

**이유**:
- 기존 `PaneArgs`는 non-interactive pane용 (`-p` 포함)
- relay pane은 인터랙티브 실행이 목적이므로 `-p` 제거 필요
- `ProviderConfig`에 `RelayPaneArgs` 필드를 추가하면 타입이 비대해짐
- 대신 `buildRelayPaneCommand` 내부에서 프로바이더별 분기 처리

**대안 검토**:
- `RelayPaneArgs` 필드 추가 — 타입 비대화, 설정 복잡도 증가로 기각
- `PaneArgs` 재활용 — 기존 pane 전략과 인수가 달라 충돌, 기각

### D3: 이전 Pane 유지 vs 즉시 닫기

**결정**: 완료된 이전 pane을 유지 (사용자가 결과를 볼 수 있도록)

**이유**:
- 사용자가 각 프로바이더의 실행 결과를 pane에서 확인할 수 있음
- 순차 실행이므로 pane 수가 최대 프로바이더 수(보통 3개)로 제한됨
- 전체 실행 완료 후 defer로 일괄 정리

**대안 검토**:
- 즉시 닫기 — 사용자 관찰 불가, 기각
- 영구 유지 — cleanup 누락 리스크, 기각

### D4: 맥락 주입 — heredoc vs 파일 참조

**결정**: heredoc으로 전체 프롬프트(이전 결과 포함)를 주입

**이유**:
- 기존 `buildPaneCommand`가 heredoc 방식을 사용 (보안 처리 완비)
- 파일 참조 방식은 프로바이더 CLI마다 지원 여부가 다름
- heredoc은 모든 프로바이더에서 동일하게 작동

**대안 검토**:
- `--context-file` 플래그 활용 — 프로바이더별 지원이 불균일, 기각

### D5: RunPaneOrchestra 분기점

**결정**: `RunPaneOrchestra` 함수 초반에 `if cfg.Strategy == StrategyRelay` 분기를 추가하여 `runRelayPaneOrchestra`로 라우팅

**이유**:
- 기존 병렬 파이프라인(split → send → collect)을 타지 않고 완전히 별도 흐름
- `RunPaneOrchestra` 진입점 하나로 모든 pane 전략을 통합

**대안 검토**:
- `runner.go`에서 직접 `runRelayPaneOrchestra` 호출 — 기존 pane/non-pane 라우팅 패턴 파괴, 기각
