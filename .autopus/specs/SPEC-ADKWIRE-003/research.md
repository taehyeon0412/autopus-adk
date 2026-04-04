# Research — SPEC-ADKWIRE-003

## Findings

### Dead Code Inventory

All 5 packages have complete implementations but zero call sites in production code:

| Package | Key Types | Start() Exists | Called | Test Coverage |
|---------|-----------|---------------|--------|---------------|
| `pkg/worker/audit` | RotatingWriter | Yes | No | Has tests |
| `pkg/worker/scheduler` | Dispatcher | Yes | No | Has tests |
| `pkg/worker/parallel` | TaskSemaphore, WorktreeManager | Yes | Only in tests | Has tests |
| `pkg/worker/knowledge` | KnowledgeSearcher | Yes | No | Has tests |
| `pkg/worker/workspace` | MultiWorkspace | Yes | No (0 references) | Has tests |

### Loop Architecture

- `pkg/worker/loop.go` — main Loop struct, currently only starts A2A server
- `pkg/worker/loop_exec.go` — task execution, no semaphore integration
- `pkg/worker/loop_lifecycle.go` — needs to be created or extended

### Learn CLI Status

- `pkg/learn/` — complete JSONL store with query, record, prune, summary functions
- `internal/cli/learn.go` — **does not exist** (Cobra commands not registered)
- Root command in `internal/cli/root.go` — needs learn subcommand registration

### TUI Control Keys

- `pkg/worker/tui/model.go` — has key handler switch/case
- 'a' (approve), 'd' (deny), 's' (skip), 'v' (view diff) — implemented
- 'p' (pause), 'c' (cancel) — referenced in help text but handler bodies are empty/stub

### Pipeline Dashboard

- `internal/cli/pipeline_dashboard.go` — renders phase status
- Currently hardcodes `PhasePending` for all phases
- `.autopus-checkpoint.yaml` format documented in `pkg/pipeline/checkpoint.go`

### Agent Run

- `internal/cli/agent_run.go` — currently creates a dummy result record
- Should dispatch via A2A path using `pkg/worker/a2a/` client
- Line ~32: `// TODO: replace with actual execution` (intentional placeholder)

### Consolidation Note

This SPEC supersedes SPEC-ADKSTUB-001, SPEC-ADKWIRE-001, and SPEC-ADKWIRE-002 by combining all worker wiring work into a single implementation. Those 3 draft SPECs should be marked as `superseded` after this SPEC is approved.
