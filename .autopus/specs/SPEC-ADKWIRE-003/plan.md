# Plan — SPEC-ADKWIRE-003

## Tasks

### Phase A: Core Lifecycle (P0)

| Task ID | Description | Agent | Mode | File Ownership | Priority |
|---------|-------------|-------|------|----------------|----------|
| T1 | Create/extend LoopLifecycle with Start/Stop for auth, poll, net | executor | sequential | autopus-adk/pkg/worker/loop_lifecycle.go | P0 |
| T2 | Wire audit.RotatingWriter instantiation in loop init | executor | sequential | autopus-adk/pkg/worker/loop.go | P0 |
| T3 | Wire parallel.TaskSemaphore acquire/release in task execution | executor | sequential | autopus-adk/pkg/worker/loop_exec.go | P0 |
| T4 | Wire real task execution in `auto agent run <task-id>` | executor | sequential | autopus-adk/internal/cli/agent_run.go | P0 |

### Phase B: Learn CLI (P0)

| Task ID | Description | Agent | Mode | File Ownership | Priority |
|---------|-------------|-------|------|----------------|----------|
| T5 | Create learn.go with 4 Cobra subcommands (query/record/prune/summary) | executor | parallel | autopus-adk/internal/cli/learn.go | P0 |
| T6 | Register learn command in root command | executor | parallel | autopus-adk/internal/cli/root.go | P0 |

### Phase C: Scheduler & Knowledge (P1)

| Task ID | Description | Agent | Mode | File Ownership | Priority |
|---------|-------------|-------|------|----------------|----------|
| T7 | Wire scheduler.Dispatcher.Start() in loop lifecycle | executor | sequential | autopus-adk/pkg/worker/loop_lifecycle.go | P1 |
| T8 | Wire knowledge.KnowledgeSearcher.Start() in loop lifecycle | executor | sequential | autopus-adk/pkg/worker/loop_lifecycle.go | P1 |

### Phase D: TUI & Dashboard (P1)

| Task ID | Description | Agent | Mode | File Ownership | Priority |
|---------|-------------|-------|------|----------------|----------|
| T9 | Implement TUI pause ('p') and cancel ('c') key handlers | executor | parallel | autopus-adk/pkg/worker/tui/model.go | P1 |
| T10 | Load checkpoint state in pipeline dashboard | executor | parallel | autopus-adk/internal/cli/pipeline_dashboard.go | P1 |

### Phase E: Multi-Workspace (P2)

| Task ID | Description | Agent | Mode | File Ownership | Priority |
|---------|-------------|-------|------|----------------|----------|
| T11 | Wire MultiWorkspace in loop with independent A2A connections | executor | sequential | autopus-adk/pkg/worker/loop.go, autopus-adk/pkg/worker/workspace/*.go | P2 |

### Phase F: Tests

| Task ID | Description | Agent | Mode | File Ownership | Priority |
|---------|-------------|-------|------|----------------|----------|
| T12 | Tests for lifecycle Start/Stop, semaphore, audit wiring | tester | sequential | autopus-adk/pkg/worker/*_test.go | P0 |
| T13 | Tests for learn CLI commands | tester | parallel | autopus-adk/internal/cli/learn_test.go | P0 |

## Execution Order

Phase A (T1→T2→T3→T4) → Phase B (T5∥T6) → Phase C (T7→T8) → Phase D (T9∥T10) → Phase E (T11) → Phase F (T12, T13)

## Estimated Complexity

HIGH — 11 files modified/created, lifecycle management, CLI registration, TUI integration, multi-workspace wiring.
