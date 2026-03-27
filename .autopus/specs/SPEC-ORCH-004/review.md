# Review: SPEC-ORCH-004

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-03-26 14:17:34

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | REQ-002 specifies adding "tool access flags" for agentic one-shot mode but does not define these flags for each provider (e.g., `claude`, `gemini`, `codex`). Since these are external CLIs with different flag conventions, the SPEC must explicitly map the required flags for each supported provider. |
| gemini | major | The SPEC does not define error handling behavior for the sequential relay. If a provider in the middle of the relay chain fails (non-zero exit or timeout), it is unclear whether the system should stop and return a partial result or continue with the next provider. |
| gemini | major | REQ-003 and REQ-004 describe a "relay" logic that is functionally redundant with the existing `pipeline` strategy (which already supports sequential execution and output injection). The SPEC should clarify the unique value of `relay` (e.g., the specific agentic flags) or consider enhancing the `pipeline` strategy instead. |
| gemini | minor | REQ-003 and REQ-007 rely on a `{jobID}` for directory naming, but the existing architecture only generates a `jobID` for "detached" (pane-mode) runs. The SPEC needs to define how `jobID` is generated for standard (non-detached) relay executions to ensure consistent storage paths. |
| gemini | suggestion | The SPEC focuses on standard execution but ignores the project's established "pane mode" (tmux/cmux orchestration). To satisfy REQ-001 ("equivalent level to existing strategies"), the relay strategy should also be implemented in `pkg/orchestra/pane_runner.go` to support interactive multi-terminal relays. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.I will investigate the existing orchestra implementation and strategy handling in `pkg/orchestra/` and `internal/cli/orchestra.go` to ensure the SPEC's feasibility and alignment with the current architecture.

I will read `pkg/orchestra/types.go` and `pkg/orchestra/strategy.go` to understand how to register the new `StrategyRelay` and ensure the SPEC's requirements are technically sound within the existing codebase.

I will read `pkg/orchestra/runner.go` to determine if the sequential execution and prompt injection required for the "relay" strategy are compatible with the current orchestration flow.

I will read `pkg/orchestra/pane_runner.go` to evaluate how the "relay" strategy should be implemented for pane-based execution and ensure it remains consistent with the existing multi-terminal orchestration patterns.

I will read `internal/cli/orchestra.go` and `internal/cli/orchestra_job.go` to determine how Job IDs are generated and whether they can be properly passed to the "relay" strategy for output storage and cleanup.

I will read `pkg/orchestra/job.go` to understand how Job IDs are generated and managed, which will help clarify how the "relay" strategy should handle its stage outputs and cleanup.

1. VERDICT: REVISE

2. FINDINGS:
   - FINDING: [critical] REQ-002 specifies adding "tool access flags" for agentic one-shot mode but does not define these flags for each provider (e.g., `claude`, `gemini`, `codex`). Since these are external CLIs with different flag conventions, the SPEC must explicitly map the required flags for each supported provider.
   - FINDING: [major] The SPEC does not define error handling behavior for the sequential relay. If a provider in the middle of the relay chain fails (non-zero exit or timeout), it is unclear whether the system should stop and return a partial result or continue with the next provider.
   - FINDING: [major] REQ-003 and REQ-004 describe a "relay" logic that is functionally redundant with the existing `pipeline` strategy (which already supports sequential execution and output injection). The SPEC should clarify the unique value of `relay` (e.g., the specific agentic flags) or consider enhancing the `pipeline` strategy instead.
   - FINDING: [minor] REQ-003 and REQ-007 rely on a `{jobID}` for directory naming, but the existing architecture only generates a `jobID` for "detached" (pane-mode) runs. The SPEC needs to define how `jobID` is generated for standard (non-detached) relay executions to ensure consistent storage paths.
   - FINDING: [suggestion] The SPEC focuses on standard execution but ignores the project's established "pane mode" (tmux/cmux orchestration). To satisfy REQ-001 ("equivalent level to existing strategies"), the relay strategy should also be implemented in `pkg/orchestra/pane_runner.go` to support interactive multi-terminal relays.

3. REASONING:
   The "Orchestra Agentic Relay Mode" is a valuable addition for multi-model reasoning chains, but the SPEC is currently too vague regarding the actual CLI integration (the "agentic" flags) and error recovery. Without defining the specific flags for REQ-002, the implementation will be blocked by ambiguity. Additionally, the overlap with the existing `pipeline` strategy suggests that `relay` might be better framed as an "agentic-pipeline" variant rather than a completely separate strategy, or it must be more clearly differentiated in its prompt injection and result merging logic. Consistency with the project's dual-execution model (standard vs. pane) is also required to maintain architectural integrity.


