# Acceptance Criteria — SPEC-ADKWIRE-003

## AC-001: Worker lifecycle start

```gherkin
Given the Worker is started with `auto agent run`
When the process initializes
Then auth refresher, poller, and network monitor are all started
And audit writer is instantiated with configured size/age limits
```

**Priority**: P0

## AC-002: Graceful shutdown

```gherkin
Given the Worker is running with active services
When SIGINT is received
Then all packages are stopped in reverse order
And shutdown completes within 10 seconds
```

**Priority**: P0

## AC-003: Parallel execution with semaphore

```gherkin
Given the semaphore limit is configured to 3
And 5 tasks are submitted simultaneously
When tasks are dispatched
Then 3 tasks execute immediately
And 2 tasks are queued until slots become available
```

**Priority**: P0

## AC-004: Learn CLI query

```gherkin
Given pipeline.jsonl contains 10 learning entries
When `auto learn query --limit 3` is executed
Then 3 matching entries are returned in JSON format
And the output is valid JSONL
```

**Priority**: P0

## AC-005: Learn CLI record

```gherkin
Given the learning store exists
When `auto learn record --type fix_pattern --files pkg/foo.go --pattern "nil check"` is executed
Then a new entry is appended to pipeline.jsonl
And the entry has a unique ID and timestamp
```

**Priority**: P0

## AC-006: TUI pause

```gherkin
Given a task is executing in the Worker TUI
When the user presses 'p'
Then the subprocess is suspended (SIGSTOP)
And the TUI displays "PAUSED" status
And pressing 'p' again resumes (SIGCONT)
```

**Priority**: P1

## AC-007: TUI cancel

```gherkin
Given a task is executing in the Worker TUI
When the user presses 'c'
Then the subprocess receives SIGTERM
And after 5 seconds receives SIGKILL if still running
And the backend is notified of cancellation
```

**Priority**: P1

## AC-008: Pipeline dashboard loads checkpoint

```gherkin
Given .autopus-checkpoint.yaml exists with Phase 2 status "completed"
When `auto pipeline dashboard` is executed
Then Phase 1 and Phase 2 show as completed
And Phase 3+ shows as pending
```

**Priority**: P1

## AC-009: Scheduler integration

```gherkin
Given a cron schedule "*/5 * * * *" is configured for a task
When the Worker is running and the schedule fires
Then the task is submitted to the execution pipeline
And the scheduler log shows the trigger
```

**Priority**: P1

## AC-010: Knowledge sync

```gherkin
Given knowledge sync is enabled and a local file is modified
When fsnotify detects the change
Then the file is synced to the Knowledge Hub API
And the sync log shows the uploaded file path
```

**Priority**: P1

## AC-011: Multi-workspace

```gherkin
Given the Worker is connected to workspace A and workspace B
When workspace A's connection drops
Then workspace B's connection remains active
And workspace A reconnects independently within 60 seconds
```

**Priority**: P2

## AC-012: Real task execution via agent run

```gherkin
Given a valid task-id exists on the backend
When `auto agent run <task-id>` is executed
Then the task is dispatched via A2A
And the actual result is returned (not a dummy record)
```

**Priority**: P0
