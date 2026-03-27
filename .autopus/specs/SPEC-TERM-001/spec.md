# SPEC-TERM-001: cmux-first Terminal Adapter Layer — 멀티에이전트 Visual Pipeline

**Status**: completed
**Created**: 2026-03-25
**Domain**: TERM

## 목적

현재 `auto go --multi` 파이프라인은 단일 터미널 세션에서 순차/병렬로 에이전트를 실행하며, 각 에이전트의 진행 상황을 시각적으로 구분할 수 없다. cmux(1순위) 또는 tmux(fallback) 터미널 멀티플렉서를 활용하여 각 에이전트를 독립 패인에서 실행하고, 파이프라인 Phase 전환에 따라 레이아웃을 동적으로 변경하는 Visual Pipeline 기능을 제공한다.

이를 통해 사용자는 planner, executor(×N), tester, reviewer 등 각 에이전트의 실시간 출력을 동시에 모니터링할 수 있다.

## 요구사항

### R1: Terminal Interface (P0)
THE SYSTEM SHALL define a `Terminal` interface in `pkg/terminal/terminal.go` with the following methods:
- `Name() string` — adapter name ("cmux", "tmux", "plain")
- `CreateWorkspace(name string) error` — create a named workspace/session
- `SplitPane(direction Direction) (PaneID, error)` — split current pane horizontally or vertically
- `SendCommand(pane PaneID, command string) error` — send a shell command to a specific pane
- `Notify(message string) error` — display a notification to the user
- `Close() error` — clean up all resources (workspace, panes)

### R2: Terminal Auto-Detection (P0)
WHEN the `--multi` flag is set, THE SYSTEM SHALL detect the available terminal multiplexer using `DetectTerminal()` in `pkg/terminal/detect.go` with priority: cmux > tmux > plain.

### R3: cmux Adapter (P0)
WHEN cmux is detected as the active terminal, THE SYSTEM SHALL use the cmux Socket API via `cmux workspace create`, `cmux pane split`, and `cmux notify` commands to manage workspaces and panes in `pkg/terminal/cmux.go`.

### R4: tmux Adapter (P0)
WHEN tmux is detected and cmux is not available, THE SYSTEM SHALL use tmux CLI commands (`new-session`, `split-window`, `send-keys`, `display-message`) to manage sessions and panes in `pkg/terminal/tmux.go`.

### R5: Graceful Degradation (P0)
WHEN neither cmux nor tmux is available, THE SYSTEM SHALL fall back to `plain` mode (no-op adapter) in `pkg/terminal/plain.go` and log a warning message indicating that visual pipeline is unavailable.

### R6: `auto agent run` Subcommand (P0)
THE SYSTEM SHALL provide an `auto agent run <task-id>` subcommand in `cmd/auto/agent_run.go` that executes a single pipeline task independently, reading task context from `.autopus/runs/<task-id>/` and writing results back to the same directory.

### R7: Pipeline Integration (P0)
WHEN the `--multi` flag is active and a terminal adapter (cmux or tmux) is available, THE SYSTEM SHALL create a workspace, split panes per agent, and execute `auto agent run <task-id>` in each pane during pipeline execution.

### R8: Phase Layout Transition (P1)
WHEN a pipeline transitions between phases, THE SYSTEM SHALL dynamically adjust the pane layout:
- Phase 1 (Plan): 1 pane (planner)
- Phase 2 (Execute): N panes (executor × N)
- Phase 3 (Test): 1 pane (tester)
- Phase 4 (Review): 2 panes (reviewer + auditor)

### R9: Dashboard Pane (P1)
WHILE the pipeline is running, THE SYSTEM SHALL maintain a persistent dashboard pane showing pipeline progress, phase status, and per-agent completion percentage.

## 생성 파일 상세

### `pkg/terminal/terminal.go`
Terminal interface 정의, Direction/PaneID 타입, 공통 에러 타입.

### `pkg/terminal/detect.go`
`DetectTerminal() Terminal` — cmux > tmux > plain 우선순위로 사용 가능한 어댑터를 반환. `pkg/detect.IsInstalled()` 재활용.

### `pkg/terminal/cmux.go`
cmux Socket API 어댑터. `cmux workspace create`, `cmux pane split --direction h|v`, `cmux notify` 등의 CLI 래핑.

### `pkg/terminal/tmux.go`
tmux CLI 어댑터. `tmux new-session`, `tmux split-window`, `tmux send-keys`, `tmux display-message` 래핑.

### `pkg/terminal/plain.go`
No-op 어댑터. 모든 메서드가 no-op으로 동작하며 경고 로그만 출력.

### `internal/cli/agent_run.go`
`auto agent run <task-id>` 서브커맨드. `.autopus/runs/<task-id>/context.yaml`에서 태스크 정보를 읽고 결과를 `.autopus/runs/<task-id>/result.yaml`에 기록.

### 기존 파일 수정
- `internal/cli/root.go` — `newAgentCmd()`에 `run` 서브커맨드 등록 (agent_create.go의 `newAgentCmd` 수정)
- `internal/cli/agent_create.go` — `newAgentCmd()`에서 `newAgentRunSubCmd()` 추가
- Pipeline 실행 코드 — `--multi` 시 Terminal adapter 연동
