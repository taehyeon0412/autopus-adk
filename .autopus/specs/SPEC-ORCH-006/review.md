# Review: SPEC-ORCH-006

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-03-26 17:52:36

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | REQ-005 is incomplete.** The requirement ends abruptly with "THE SYSTEM SHALL 다음 순서로 인터랙티브 모드를 실행한다:" without providing the execution sequence. |
| gemini | critical | REQ-007 is incomplete.** The requirement ends with "THE SYSTEM SHALL 다음 전략으로 완료를 감지한다:" but lacks the actual strategy for completion detection. In an interactive terminal session, detecting when a sub-agent (like Claude) has finished its task is non-trivial and essential for the SPEC. |
| gemini | major | REQ-010 (Output cleaning) is underspecified.** "CLI 프롬프트 장식 제거" (removing CLI prompt decorations) is highly dependent on the provider (Claude, Gemini, etc.). The SPEC should define how these prompts are identified (e.g., regex per provider) to ensure `ReadScreen` returns clean results. |
| gemini | major | Ambiguity between `ReadScreen` and `PipePane`.** The SPEC adds both to the `Terminal` interface (REQ-001) but does not clarify their relationship. For instance, in `TmuxAdapter`, `ReadScreen` would likely use `capture-pane`, while `PipePane` uses `pipe-pane`. The SPEC should specify which one is the primary source for the final result. |
| gemini | minor | REQ-004 (PlainAdapter no-op) fallback logic.** While the no-op implementation is defined, the SPEC should explicitly state that if `ReadScreen` returns an empty string, the system must trigger the fallback to the non-interactive mode (REQ-008). |
| gemini | suggestion | Granular Timeout handling.** REQ-009 mentions a timeout, but interactive modes may hang at different stages (e.g., initialization vs. final response). A more detailed timeout strategy (e.g., idle timeout vs. total timeout) would improve robustness. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.1. **VERDICT: REVISE**

2. **FINDINGS:**
   - **FINDING: [critical] REQ-005 is incomplete.** The requirement ends abruptly with "THE SYSTEM SHALL 다음 순서로 인터랙티브 모드를 실행한다:" without providing the execution sequence.
   - **FINDING: [critical] REQ-007 is incomplete.** The requirement ends with "THE SYSTEM SHALL 다음 전략으로 완료를 감지한다:" but lacks the actual strategy for completion detection. In an interactive terminal session, detecting when a sub-agent (like Claude) has finished its task is non-trivial and essential for the SPEC.
   - **FINDING: [major] REQ-010 (Output cleaning) is underspecified.** "CLI 프롬프트 장식 제거" (removing CLI prompt decorations) is highly dependent on the provider (Claude, Gemini, etc.). The SPEC should define how these prompts are identified (e.g., regex per provider) to ensure `ReadScreen` returns clean results.
   - **FINDING: [major] Ambiguity between `ReadScreen` and `PipePane`.** The SPEC adds both to the `Terminal` interface (REQ-001) but does not clarify their relationship. For instance, in `TmuxAdapter`, `ReadScreen` would likely use `capture-pane`, while `PipePane` uses `pipe-pane`. The SPEC should specify which one is the primary source for the final result.
   - **FINDING: [minor] REQ-004 (PlainAdapter no-op) fallback logic.** While the no-op implementation is defined, the SPEC should explicitly state that if `ReadScreen` returns an empty string, the system must trigger the fallback to the non-interactive mode (REQ-008).
   - **FINDING: [suggestion] Granular Timeout handling.** REQ-009 mentions a timeout, but interactive modes may hang at different stages (e.g., initialization vs. final response). A more detailed timeout strategy (e.g., idle timeout vs. total timeout) would improve robustness.

3. **REASONING:**
   The SPEC aims to introduce a significant architectural shift from "sentinel-based file polling" to a truly "interactive screen-reading" mode for terminal panes. However, the most critical parts of this transition—the execution sequence and the completion detection strategy—are missing text in the current draft. Without defining how the system knows an interactive agent has finished processing, the implementation cannot be completed or verified. Additionally, the logic for cleaning "CLI prompts" needs more technical detail to be feasible across different providers.


