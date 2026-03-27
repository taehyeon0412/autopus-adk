# Review: SPEC-ORCH-003

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-03-26 12:40:49

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | Logical contradiction between REQ-001 and REQ-008/009. REQ-001 mandates that the system return a Job ID within 2 seconds (detaching), while REQ-008/009 require the system to wait for completion, merge results, print them, and then clean up. A process that exits in 2 seconds cannot perform these subsequent blocking operations unless a separate collection command or background daemon is specified, which is missing from the SPEC. |
| gemini | major | Violation of Backward Compatibility (REQ-010). The current `RunPaneOrchestra` behavior is blocking and prints results to stdout. REQ-001 states the system "SHALL automatically switch to detach mode." If this becomes the default behavior, existing scripts and users expecting immediate output will instead receive a Job ID, breaking compatibility. |
| gemini | major | Missing Retrieval Mechanism. The SPEC defines how to start a detached job but fails to specify the CLI command or mechanism for a user to later retrieve the status or results of that job (e.g., `auto orchestra status <id>` or `auto orchestra collect <id>`). |
| gemini | minor | Ambiguity in REQ-003. While it mandates forcing blocking execution, it does not define the user interface for this override (e.g., a `--no-detach` flag or an environment variable). |
| gemini | minor | Feasibility of REQ-012. TTL-based cleanup is mentioned, but the mechanism for triggering this cleanup (e.g., a background worker, or a "garbage collection" phase during other CLI commands) is not defined, leading to potential resource leaks if not implemented carefully. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.1. VERDICT: REVISE

2. FINDINGS:
- FINDING: [critical] Logical contradiction between REQ-001 and REQ-008/009. REQ-001 mandates that the system return a Job ID within 2 seconds (detaching), while REQ-008/009 require the system to wait for completion, merge results, print them, and then clean up. A process that exits in 2 seconds cannot perform these subsequent blocking operations unless a separate collection command or background daemon is specified, which is missing from the SPEC.
- FINDING: [major] Violation of Backward Compatibility (REQ-010). The current `RunPaneOrchestra` behavior is blocking and prints results to stdout. REQ-001 states the system "SHALL automatically switch to detach mode." If this becomes the default behavior, existing scripts and users expecting immediate output will instead receive a Job ID, breaking compatibility.
- FINDING: [major] Missing Retrieval Mechanism. The SPEC defines how to start a detached job but fails to specify the CLI command or mechanism for a user to later retrieve the status or results of that job (e.g., `auto orchestra status <id>` or `auto orchestra collect <id>`).
- FINDING: [minor] Ambiguity in REQ-003. While it mandates forcing blocking execution, it does not define the user interface for this override (e.g., a `--no-detach` flag or an environment variable).
- FINDING: [minor] Feasibility of REQ-012. TTL-based cleanup is mentioned, but the mechanism for triggering this cleanup (e.g., a background worker, or a "garbage collection" phase during other CLI commands) is not defined, leading to potential resource leaks if not implemented carefully.

3. REASONING:
The specification introduces a valuable "Detach Mode" to improve the responsiveness of the orchestration engine, but it contains a fundamental logical flaw regarding the lifecycle of the command. It is impossible for a single command invocation to both return control to the user in 2 seconds (detach) and subsequently print the merged results of a multi-minute process (block). The SPEC needs to clearly distinguish between the "Job Submission" phase and the "Result Collection" phase, likely by introducing new subcommands or defining a "wait/follow" flag. Additionally, "automatically switching" to a non-blocking mode by default would break the promised 100% backward compatibility with current blocking workflows.


