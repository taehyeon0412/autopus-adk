# Review: SPEC-TERM-001

**Verdict**: PASS
**Revision**: 0
**Date**: 2026-03-25 18:20:05

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | Incomplete Requirements (Truncated Text).** `REQ-001` and `REQ-008` end with colons but do not list the mandatory methods or layout rules. Without the definition of the `Terminal` interface methods and the logic for dynamic adjustment, the SPEC cannot be implemented or verified. |
| gemini | major | Missing Lifecycle/Cleanup Requirements.** The SPEC does not define a requirement for cleaning up terminal resources (closing tmux sessions, deleting cmux workspaces) upon pipeline completion (success or failure) or process interruption (SIGINT/SIGTERM). This will result in orphaned terminal sessions and resource leakage. |
| gemini | major | Signal Propagation and Process Management.** There is no requirement for how the master `auto` process should propagate signals (like `Ctrl+C`) to the `auto agent run` processes living inside multiplexer panes. |
| gemini | minor | Pipeline State Initialization.** `REQ-006` and `REQ-007` rely on context stored in `.autopus/runs/<task-id>/`, but the SPEC lacks a requirement for the orchestrator to prepare/hydrate these directories before launching the agent panes. |
| gemini | minor | Dashboard Synchronization.** `REQ-009` requires a persistent dashboard but does not specify the communication mechanism between the independent `agent run` processes and the dashboard pane (e.g., file-based polling vs. IPC/Socket). |
| gemini | suggestion | Interface Ambiguity.** `REQ-003` mentions using the "cmux Socket API via commands." It should clarify if the implementation should use a Go client library (Socket) or simply wrap the `cmux` CLI binary via `os/exec`. |

## Provider Responses

### Response 1




### Response 2

MCP issues detected. Run /mcp list for status.### 1. VERDICT: **REVISE**

---

### 2. FINDINGS

- **FINDING: [critical] Incomplete Requirements (Truncated Text).** `REQ-001` and `REQ-008` end with colons but do not list the mandatory methods or layout rules. Without the definition of the `Terminal` interface methods and the logic for dynamic adjustment, the SPEC cannot be implemented or verified.
- **FINDING: [major] Missing Lifecycle/Cleanup Requirements.** The SPEC does not define a requirement for cleaning up terminal resources (closing tmux sessions, deleting cmux workspaces) upon pipeline completion (success or failure) or process interruption (SIGINT/SIGTERM). This will result in orphaned terminal sessions and resource leakage.
- **FINDING: [major] Signal Propagation and Process Management.** There is no requirement for how the master `auto` process should propagate signals (like `Ctrl+C`) to the `auto agent run` processes living inside multiplexer panes.
- **FINDING: [minor] Pipeline State Initialization.** `REQ-006` and `REQ-007` rely on context stored in `.autopus/runs/<task-id>/`, but the SPEC lacks a requirement for the orchestrator to prepare/hydrate these directories before launching the agent panes.
- **FINDING: [minor] Dashboard Synchronization.** `REQ-009` requires a persistent dashboard but does not specify the communication mechanism between the independent `agent run` processes and the dashboard pane (e.g., file-based polling vs. IPC/Socket).
- **FINDING: [suggestion] Interface Ambiguity.** `REQ-003` mentions using the "cmux Socket API via commands." It should clarify if the implementation should use a Go client library (Socket) or simply wrap the `cmux` CLI binary via `os/exec`.

---

### 3. REASONING

The SPEC provides a solid architectural direction for a "visual pipeline" using terminal multiplexers, which is highly beneficial for debugging multi-agent workflows. The use of a standalone `auto agent run` subcommand (`REQ-006`) is an excellent design choice for process isolation and scalability.

However, the document is currently **incomplete**. The truncation of `REQ-001` (Terminal interface methods) and `REQ-008` (Layout rules) leaves the core implementation details undefined. Furthermore, from an operational perspective, the lack of a cleanup mechanism for tmux/cmux is a significant flaw that would negatively impact the user experience by leaving cluttered sessions behind. 

To move to a **PASS**, the SPEC needs to:
1. Complete the truncated requirements.
2. Define a `Teardown()` or `Close()` method in the `Terminal` interface.
3. Explicitly state how state is prepared before agents are launched.
4. Specify the signal handling strategy for the sub-processes.


### Response 3




