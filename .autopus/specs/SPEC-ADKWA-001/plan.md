# Plan: SPEC-ADKWA-001

## Task Assignment Table

| Task ID | Description | Agent | Mode | File Ownership | Complexity |
|---------|-------------|-------|------|----------------|------------|
| T1 | A2A types + Server approval handling + SendApprovalResponse | executor | sequential | `pkg/worker/a2a/types.go`, `pkg/worker/a2a/server.go`, `pkg/worker/a2a/server_test.go` | MEDIUM |
| T2 | TUI approval callback wiring (OnApprovalDecision, OnViewDiff) | executor | sequential | `pkg/worker/tui/model.go`, `pkg/worker/tui/model_test.go` | LOW |
| T3 | Integration: WorkerLoop wiring (A2A callback → TUI program) | executor | sequential | `pkg/worker/loop.go` | LOW |

## Dependency

T1 → T2 → T3 (sequential: T2 depends on T1 types, T3 depends on both)

## Flow Diagram

```
Backend (A2A WS)
  │ tasks/approval {task_id, action, risk_level, context}
  ▼
Server.handleApproval()          ← T1
  │ UpdateTaskStatus(input-required)
  │ call ApprovalCallback(ApprovalRequestParams)
  ▼
WorkerLoop bridges to TUI        ← T3
  │ send ApprovalRequestMsg to tea.Program
  ▼
TUI.handleKey() a/d/s            ← T2
  │ call OnApprovalDecision(taskID, decision)
  ▼
Server.SendApprovalResponse()    ← T1
  │ tasks/approvalResponse {task_id, decision}
  │ UpdateTaskStatus(working)
  ▼
Backend → MCP response → Claude Code resumes
```
