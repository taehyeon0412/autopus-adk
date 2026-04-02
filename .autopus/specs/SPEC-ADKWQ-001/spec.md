# SPEC: ADK Worker Quality — Security Hardening + P0 Gap Resolution

**SPEC-ID**: SPEC-ADKWQ-001
**Status**: completed
**Created**: 2026-04-02
**Source**: SPEC-ADKW-001 post-implementation audit
**Parent**: SPEC-ADKW-001

---

## Domain: Security Hardening

### REQ-SEC-H01 [P0]
WHEN `handleSearchKnowledge` builds a query URL,
THE SYSTEM SHALL use `url.QueryEscape` on the query parameter to prevent URL injection.

### REQ-SEC-H02 [P0]
WHEN `WriteHookConfig` generates the hook command,
THE SYSTEM SHALL shell-quote the `policyPath` argument to prevent shell injection.

### REQ-SEC-H03 [P0]
WHEN `RemoveHookConfig` removes worker hooks,
THE SYSTEM SHALL only remove the worker-specific hook key (`PreToolUse`), preserving other user-defined hooks.

### REQ-SEC-H04 [P0]
WHEN `CacheSecurityPolicy` writes a policy file,
THE SYSTEM SHALL explicitly call `os.Chmod(0600)` on the temp file before rename, not relying on umask.

## Domain: Subprocess Executor

### REQ-SUB-H01 [P0]
WHEN spawning a subprocess,
THE SYSTEM SHALL capture stderr via `cmd.Stderr = &stderrBuf` and include stderr content in error reports on non-zero exit.

### REQ-SUB-H02 [P0]
WHEN `parseStream` encounters a `system.task_notification` event,
THE SYSTEM SHALL log the event details and forward to the A2A status reporter (not silently drop).

### REQ-SUB-H03 [P1]
WHEN `parseStream` detects a context overflow error in the stream output,
THE SYSTEM SHALL automatically fallback to `PipelineExecutor.Execute()` for phase-split execution.

## Domain: A2A Integration

### REQ-A2A-H01 [P0]
WHEN a subprocess emits an MCP `request_approval` event,
THE SYSTEM SHALL update the A2A task status to `input-required` and surface the approval to the TUI.

### REQ-A2A-H02 [P0]
WHEN `dispatchTask` spawns a goroutine,
THE SYSTEM SHALL use a per-task cancellable context (not the server's root context), enabling `handleCancelTask` to cancel the running subprocess.

## Domain: MCP Client Types

### REQ-MCP-H01 [P0]
THE SYSTEM SHALL update `ProgressReport` struct fields to match SPEC-ADKW-001 REQ-MCP-02: `Phase string`, `Status string`, `Details map[string]any` (replacing `Progress float64` and `Message string`).

## Domain: Setup

### REQ-SETUP-H01 [P0]
THE SYSTEM SHALL implement `RefreshToken(ctx, backendURL, refreshToken) (*TokenResponse, error)` in `setup/auth.go`, calling the backend token endpoint with `grant_type=refresh_token`.

### REQ-SETUP-H02 [P0]
WHEN multiple workspaces are available in `SelectWorkspace`,
THE SYSTEM SHALL present an interactive selection (charmbracelet/huh) instead of returning the first workspace.

## Domain: Code Quality

### REQ-QUAL-01 [P1]
WHEN `io.Copy` writes the prompt to subprocess stdin in `loop_exec.go`,
THE SYSTEM SHALL check and log the error return value.

### REQ-QUAL-02 [P1]
WHEN a schedule is removed from the backend but still exists in the `lastTrigger` map,
THE SYSTEM SHALL prune stale schedule entries on each fetch cycle.

### REQ-QUAL-03 [P1]
WHEN `RotatingWriter.Close()` is called,
THE SYSTEM SHALL set `w.file = nil` and guard `Write()` against nil file.

### REQ-QUAL-04 [P1]
WHEN `GeneratePlist` renders the launchd plist template,
THE SYSTEM SHALL use `html/template` or manual XML escaping for argument values.

### REQ-QUAL-05 [P1]
WHEN `FileWatcher` detects file changes,
THE SYSTEM SHALL integrate with `Excluder` to skip excluded paths, and prune deleted files from the `mtimes` map.

## Domain: Test Coverage

### REQ-TEST-01 [P1]
THE SYSTEM SHALL add test files for: `setup/workspace_test.go`, `setup/provider_auth_test.go`, `tui/approval_test.go`.

### REQ-TEST-02 [P1]
THE SYSTEM SHALL expand `net/monitor_test.go` to test the polling loop, change detection, and callback invocation.

### REQ-TEST-03 [P1]
THE SYSTEM SHALL expand `setup/auth_test.go` to test `RefreshToken`, `RequestDeviceCode`, `PollForToken`, and `SaveCredentials`.
