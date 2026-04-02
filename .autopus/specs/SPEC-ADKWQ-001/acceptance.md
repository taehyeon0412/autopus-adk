# Acceptance Criteria: SPEC-ADKWQ-001

## Security

- [x] `mcpserver/tools.go`: query parameter is URL-encoded via `url.QueryEscape`
- [x] `security/hook_template.go`: policyPath is shell-quoted in hook command
- [x] `security/hook_template.go`: RemoveHookConfig only removes `PreToolUse` key
- [x] `security/policy_cache.go`: explicit `os.Chmod(0600)` on temp file

## Subprocess

- [x] `loop_exec.go`: `cmd.Stderr` set to capture buffer
- [x] `loop_exec.go`: stderr content included in error on non-zero exit
- [x] `loop_exec.go`: `system.task_notification` event handled in parseStream switch
- [x] `a2a/server.go`: per-task cancellable context in dispatchTask
- [ ] `a2a/server.go` or `loop.go`: input-required status triggered on approval event (deferred — requires MCP SSE integration)

## MCP Types

- [x] `mcp/types.go`: ProgressReport has Phase, Status, Details fields

## Setup

- [x] `setup/auth.go`: RefreshToken function implemented
- [x] `setup/workspace.go`: interactive multi-workspace selection (not first-return stub)

## Code Quality

- [x] `loop_exec.go`: io.Copy error checked and logged
- [x] `scheduler/dispatcher.go`: stale lastTrigger entries pruned on fetch
- [x] `audit/rotation.go`: Close() nils file, Write() guards against nil
- [x] `daemon/launchd.go`: XML escaping for plist arguments
- [x] `knowledge/watcher.go`: Excluder integrated, deleted files pruned

## Tests

- [x] `setup/workspace_test.go` exists with real tests
- [x] `setup/provider_auth_test.go` exists with real tests
- [x] `tui/approval_test.go` exists with render tests
- [x] `net/monitor_test.go` tests polling loop and callbacks
- [x] `setup/auth_test.go` tests RefreshToken + other functions
- [x] `go test -race ./pkg/worker/...` passes (19/19 packages)
