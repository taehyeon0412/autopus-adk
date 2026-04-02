# Plan: SPEC-ADKWQ-001

## Task Assignment Table

| Task ID | Description | Agent | Mode | File Ownership | Complexity |
|---------|-------------|-------|------|----------------|------------|
| T1 | Security: URL injection fix + shell quote + selective hook removal + explicit chmod | executor | parallel | `pkg/worker/mcpserver/tools.go`, `pkg/worker/security/hook_template.go`, `pkg/worker/security/policy_cache.go` | MEDIUM |
| T2 | Subprocess: stderr capture + task_notification + cancel context + input-required trigger | executor | parallel | `pkg/worker/loop_exec.go`, `pkg/worker/loop.go`, `pkg/worker/a2a/server.go` | MEDIUM |
| T3 | MCP types + Setup auth refresh + workspace selection | executor | parallel | `pkg/worker/mcp/types.go`, `pkg/worker/setup/auth.go`, `pkg/worker/setup/workspace.go` | MEDIUM |
| T4 | Quality: io.Copy error, scheduler prune, audit nil guard, launchd XML escape, watcher excluder | executor | parallel | `pkg/worker/scheduler/dispatcher.go`, `pkg/worker/audit/rotation.go`, `pkg/worker/daemon/launchd.go`, `pkg/worker/knowledge/watcher.go` | MEDIUM |
| T5 | Tests: workspace_test, provider_auth_test, approval_test, monitor_test expansion, auth_test expansion | executor | parallel | `pkg/worker/setup/*_test.go`, `pkg/worker/tui/approval_test.go`, `pkg/worker/net/monitor_test.go` | MEDIUM |

## Dependency

All tasks are independent (no file ownership overlap) → parallel execution.

## Notes

- T1-T4: Each task modifies source + updates corresponding test files
- T5: Test-only additions, no source file changes
- All tasks target `autopus-adk/pkg/worker/` (single module)
