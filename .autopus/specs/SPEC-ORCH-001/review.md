# Review: SPEC-ORCH-001

**Verdict**: PASS
**Revision**: 0
**Date**: 2026-03-25 22:57:15

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | major | Context Mismatch.** The "Existing Code Context" provided belongs to the Autopus backend server (`cmd/server/main.go`), whereas the SPEC describes logic for a CLI orchestrator and terminal multiplexer (`orchestra`, `cmux`, `pkg/terminal`). There is no code provided for the actual `orchestra` implementation or the `terminal` package mentioned in **REQ-001**. |
| gemini | minor | stdout Redirection vs. Interactive Visibility.** **REQ-008** states stdout will be redirected to a temporary file, while **REQ-002/003** specifies an "interactive mode" with panes. If standard redirection (`> file`) is used, the user will see a blank screen in the `cmux` pane, defeating the purpose of the split-pane visualization. The spec should clarify if `tee` or a similar mechanism is used to maintain visibility. |
| gemini | minor | Sentinel Reliability.** **REQ-008** relies on a "sentinel" (completion marker) to detect when a provider is finished. Terminal-based sentinel detection is notoriously brittle (e.g., if the provider's output contains the sentinel string, or if an error prevents the sentinel from being written). A more robust mechanism (like monitoring the PID or using `cmux` state) should be considered. |
| gemini | suggestion | ANSI Escape Code Handling.** Captured output from a terminal (especially in interactive mode) often contains ANSI escape codes for colors and cursor movement. **REQ-005** should specify that these will be stripped before being passed to the `merge/judge` logic to prevent parsing errors. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.1. **VERDICT**: **REVISE**

2. **FINDINGS**:
   - **FINDING: [major] Context Mismatch.** The "Existing Code Context" provided belongs to the Autopus backend server (`cmd/server/main.go`), whereas the SPEC describes logic for a CLI orchestrator and terminal multiplexer (`orchestra`, `cmux`, `pkg/terminal`). There is no code provided for the actual `orchestra` implementation or the `terminal` package mentioned in **REQ-001**.
   - **FINDING: [minor] stdout Redirection vs. Interactive Visibility.** **REQ-008** states stdout will be redirected to a temporary file, while **REQ-002/003** specifies an "interactive mode" with panes. If standard redirection (`> file`) is used, the user will see a blank screen in the `cmux` pane, defeating the purpose of the split-pane visualization. The spec should clarify if `tee` or a similar mechanism is used to maintain visibility.
   - **FINDING: [minor] Sentinel Reliability.** **REQ-008** relies on a "sentinel" (completion marker) to detect when a provider is finished. Terminal-based sentinel detection is notoriously brittle (e.g., if the provider's output contains the sentinel string, or if an error prevents the sentinel from being written). A more robust mechanism (like monitoring the PID or using `cmux` state) should be considered.
   - **FINDING: [suggestion] ANSI Escape Code Handling.** Captured output from a terminal (especially in interactive mode) often contains ANSI escape codes for colors and cursor movement. **REQ-005** should specify that these will be stripped before being passed to the `merge/judge` logic to prevent parsing errors.

3. **REASONING**:
   The specification is well-structured and covers the basic workflow (Detection -> Execution -> Capture -> Cleanup). However, the primary reason for the **REVISE** verdict is the significant gap between the provided backend server context and the CLI-focused requirements. Without seeing the existing `orchestra` or `terminal` package structures, it is difficult to verify the feasibility of "reusing existing merge/judge logic" (**REQ-005**) or how the `send-keys` mechanism (**REQ-004**) will interact with the system's process management. Additionally, the conflict between "interactive panes" and "stdout redirection" needs a technical resolution (e.g., using `script` or `tee`) to ensure the feature provides the intended UX.


### Response 2




