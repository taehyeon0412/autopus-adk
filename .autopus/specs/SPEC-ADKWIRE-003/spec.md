# SPEC-ADKWIRE-003: ADK Worker 전체 패키지 Wiring 통합

**Status**: completed
**Created**: 2026-04-05
**Target Module**: autopus-adk
**Priority**: P2

## Overview

SPEC-ADKSTUB-001, SPEC-ADKWIRE-001, SPEC-ADKWIRE-002에서 식별된 ADK Worker의 모든 dead code 및 stub을 실제로 연결한다. 이 SPEC은 3개 기존 draft SPEC의 작업을 통합하여 한 번의 파이프라인으로 완료한다.

## Background

현재 ADK Worker의 구현 상태:
- **Core loop** (`pkg/worker/loop.go`): A2A 서버 시작, 태스크 수신/실행은 동작
- **Dead code 5개 패키지**: audit, scheduler, parallel, knowledge, workspace — 코드 존재하나 Start() 호출 없음
- **Stub CLI 3개**: `auto learn` 미등록, `auto agent run` 더미 결과, pipeline dashboard 상태 미로딩
- **TUI 제어**: pause/cancel 키 미동작

## Requirements

### P0 — Must Have (Worker Lifecycle)

| ID | Requirement |
|----|-------------|
| REQ-01 | WHEN the Worker starts via `auto agent run`, THE SYSTEM SHALL call `Start()` on auth (token refresher), poll (REST poller), and net (network monitor) packages in sequence |
| REQ-02 | WHEN the Worker stops (SIGINT/SIGTERM), THE SYSTEM SHALL call `Stop()` on all started packages in reverse order with a 10-second grace period |
| REQ-03 | WHEN `auto agent run <task-id>` is called with a task ID, THE SYSTEM SHALL execute the task via the A2A dispatch path and return the actual result instead of a dummy record |

### P0 — Must Have (Audit & Parallel)

| ID | Requirement |
|----|-------------|
| REQ-04 | WHEN the Worker loop initializes, THE SYSTEM SHALL instantiate `audit.RotatingWriter` with configured max size and age, and wire it as the audit logger for all task executions |
| REQ-05 | WHEN the Worker receives a task for execution, THE SYSTEM SHALL acquire a slot from `parallel.TaskSemaphore` before spawning the subprocess, and release it on completion |
| REQ-06 | WHEN the parallel semaphore has no available slots, THE SYSTEM SHALL queue the task and execute it when a slot becomes available |

### P0 — Must Have (Learn CLI)

| ID | Requirement |
|----|-------------|
| REQ-07 | WHEN `auto learn query` is called, THE SYSTEM SHALL query the JSONL store with the provided filters (--files, --pattern, --limit) and return matching entries |
| REQ-08 | WHEN `auto learn record` is called, THE SYSTEM SHALL append a new entry to the JSONL store |
| REQ-09 | WHEN `auto learn prune` is called, THE SYSTEM SHALL remove entries older than the specified --max-age |
| REQ-10 | WHEN `auto learn summary` is called, THE SYSTEM SHALL display a summary of learning entries since last sync |

### P1 — Should Have (Scheduler & Knowledge)

| ID | Requirement |
|----|-------------|
| REQ-11 | WHEN the Worker starts, THE SYSTEM SHALL call `scheduler.Dispatcher.Start()` to begin evaluating cron schedules |
| REQ-12 | WHEN a scheduled task fires, THE SYSTEM SHALL submit it to the Worker's task execution pipeline |
| REQ-13 | WHEN the Worker starts with knowledge sync enabled, THE SYSTEM SHALL start `knowledge.KnowledgeSearcher` and populate search results from the Knowledge Hub API |
| REQ-14 | WHEN a knowledge file changes (fsnotify), THE SYSTEM SHALL re-sync the changed file to the Knowledge Hub |

### P1 — Should Have (TUI & Dashboard)

| ID | Requirement |
|----|-------------|
| REQ-15 | WHEN the user presses 'p' in the Worker TUI, THE SYSTEM SHALL pause the current task execution (suspend subprocess) and display "PAUSED" status |
| REQ-16 | WHEN the user presses 'c' in the Worker TUI, THE SYSTEM SHALL cancel the current task (kill subprocess with SIGTERM, then SIGKILL after 5s) and report cancellation to the backend |
| REQ-17 | WHEN `auto pipeline dashboard` is called, THE SYSTEM SHALL load checkpoint state from `.autopus-checkpoint.yaml` and display actual phase statuses instead of all PhasePending |

### P2 — Could Have (Multi-Workspace)

| ID | Requirement |
|----|-------------|
| REQ-18 | WHEN the Worker is configured with multiple workspaces, THE SYSTEM SHALL maintain independent A2A connections per workspace via `workspace.MultiWorkspace` |
| REQ-19 | WHEN a workspace connection drops, THE SYSTEM SHALL reconnect only that workspace without affecting others |

## Implementation Details

### Worker Loop Lifecycle (`pkg/worker/loop_lifecycle.go`)

This file needs to be created or extended to manage the startup/shutdown sequence:

```go
type LoopLifecycle struct {
    auth      *auth.TokenRefresher
    poller    *poll.AdaptivePoller
    netMon    *net.InterfaceMonitor
    audit     *audit.RotatingWriter
    scheduler *scheduler.Dispatcher
    knowledge *knowledge.KnowledgeSearcher
    semaphore *parallel.TaskSemaphore
}

func (l *LoopLifecycle) Start(ctx context.Context) error {
    // Start in dependency order
    if err := l.auth.Start(ctx); err != nil { return err }
    if err := l.poller.Start(ctx); err != nil { return err }
    l.netMon.Start(ctx)
    l.audit.Start()
    if l.scheduler != nil { l.scheduler.Start(ctx) }
    if l.knowledge != nil { l.knowledge.Start(ctx) }
    return nil
}

func (l *LoopLifecycle) Stop(ctx context.Context) {
    // Stop in reverse order with 10s grace
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    if l.knowledge != nil { l.knowledge.Stop() }
    if l.scheduler != nil { l.scheduler.Stop() }
    l.audit.Close()
    l.netMon.Stop()
    l.poller.Stop()
    l.auth.Stop()
}
```

### Learn CLI Registration (`internal/cli/learn.go`)

Create a new file with 4 Cobra subcommands:
- `auto learn query --files <paths> --pattern <pattern> --limit <n> --format <prompt|json>`
- `auto learn record --type <type> --files <paths> --pattern <desc> --spec <id>`
- `auto learn prune --max-age <duration>`
- `auto learn summary --since-last-sync`

Each command delegates to `pkg/learn/` package which already has the JSONL store implementation.

### Task Execution (`pkg/worker/loop_exec.go`)

Add semaphore integration:
```go
func (l *Loop) executeTask(ctx context.Context, task *a2a.Task) {
    // Acquire semaphore slot
    if err := l.semaphore.Acquire(ctx); err != nil {
        // Queue for later
        return
    }
    defer l.semaphore.Release()

    // Existing execution logic...
}
```

### TUI Control Keys (`pkg/worker/tui/model.go`)

Wire 'p' and 'c' key handlers:
```go
case "p":
    if m.currentTask != nil {
        m.currentTask.Pause() // SIGSTOP to subprocess
        m.status = "PAUSED"
    }
case "c":
    if m.currentTask != nil {
        m.currentTask.Cancel() // SIGTERM, then SIGKILL after 5s
        m.status = "CANCELLED"
    }
```

## Acceptance Criteria

```gherkin
Feature: ADK Worker Full Wiring

  Scenario: AC-001 Worker lifecycle
    Given the Worker is started with `auto agent run`
    When the process initializes
    Then auth refresher, poller, and network monitor are all started
    And audit writer is instantiated

  Scenario: AC-002 Graceful shutdown
    Given the Worker is running with active tasks
    When SIGINT is received
    Then all packages are stopped in reverse order within 10 seconds

  Scenario: AC-003 Parallel execution
    Given the semaphore limit is 3
    And 5 tasks are submitted simultaneously
    Then 3 tasks execute immediately
    And 2 tasks are queued until slots free

  Scenario: AC-004 Learn CLI query
    Given pipeline.jsonl contains 10 entries
    When `auto learn query --limit 3` is run
    Then 3 matching entries are returned

  Scenario: AC-005 TUI pause
    Given a task is executing in the Worker TUI
    When the user presses 'p'
    Then the subprocess is suspended
    And the TUI shows "PAUSED" status

  Scenario: AC-006 TUI cancel
    Given a task is executing in the Worker TUI
    When the user presses 'c'
    Then the subprocess receives SIGTERM
    And after 5 seconds SIGKILL if still running
    And the backend is notified of cancellation

  Scenario: AC-007 Pipeline dashboard loads state
    Given a .autopus-checkpoint.yaml exists with Phase 2 completed
    When `auto pipeline dashboard` is run
    Then Phase 1 and 2 show as completed, not PhasePending

  Scenario: AC-008 Scheduler fires task
    Given a cron schedule "*/5 * * * *" is configured
    When the schedule fires
    Then the task is submitted to the Worker execution pipeline

  Scenario: AC-009 Knowledge sync on file change
    Given knowledge sync is enabled
    When a local file is modified
    Then the file is synced to the Knowledge Hub API

  Scenario: AC-010 Multi-workspace independent connections
    Given the Worker is connected to workspace A and B
    When workspace A's connection drops
    Then workspace B's connection remains active
    And workspace A reconnects independently
```

## Dependencies

- `pkg/worker/auth` — TokenRefresher (already implemented, needs Start() call)
- `pkg/worker/poll` — AdaptivePoller (already implemented, needs Start() call)
- `pkg/worker/net` — InterfaceMonitor (already implemented, needs Start() call)
- `pkg/worker/audit` — RotatingWriter (already implemented, needs instantiation)
- `pkg/worker/scheduler` — Dispatcher (already implemented, needs Start() call)
- `pkg/worker/parallel` — TaskSemaphore (already implemented, needs Acquire/Release calls)
- `pkg/worker/knowledge` — KnowledgeSearcher (already implemented, needs Start() call)
- `pkg/worker/workspace` — MultiWorkspace (already implemented, needs reference in loop)
- `pkg/learn` — JSONL store (already implemented, needs CLI commands)

## Out of Scope

- New Worker features not in existing packages
- Worker daemon mode improvements (launchd/systemd — already working)
- A2A server protocol changes
- Orchestra engine changes
