# Acceptance Criteria: SPEC-ADKWA-001

## A2A Types
- [ ] `a2a/types.go`: `MethodApproval` and `MethodApprovalResponse` constants defined
- [ ] `a2a/types.go`: `ApprovalRequestParams` struct with TaskID, Action, RiskLevel, Context
- [ ] `a2a/types.go`: `ApprovalResponseParams` struct with TaskID, Decision

## A2A Server
- [ ] `a2a/server.go`: `handleMessage` routes `tasks/approval` to `handleApproval`
- [ ] `a2a/server.go`: `handleApproval` parses params, updates status to `input-required`, calls callback
- [ ] `a2a/server.go`: `SendApprovalResponse` sends `tasks/approvalResponse` and restores `working` status
- [ ] `a2a/server.go`: `ApprovalCallback` field in `ServerConfig`

## TUI
- [ ] `tui/model.go`: `OnApprovalDecision` callback field on `WorkerModel`
- [ ] `tui/model.go`: approval keypress (a/d/s) invokes `OnApprovalDecision(taskID, decision)`
- [ ] `tui/model.go`: `v` keypress invokes `OnViewDiff` callback (if set)

## Integration
- [ ] `loop.go`: `WorkerLoop` wires A2A approval callback to TUI message channel

## Tests
- [ ] `a2a/server_test.go`: test `handleApproval` updates status and calls callback
- [ ] `a2a/server_test.go`: test `SendApprovalResponse` sends correct JSON-RPC
- [ ] `tui/model_test.go`: test approval keys invoke callback with correct decision
- [ ] `go test -race ./pkg/worker/...` passes
