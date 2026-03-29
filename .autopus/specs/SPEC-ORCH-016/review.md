# Review: SPEC-ORCH-016

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-03-29 08:50:37

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | Missing `AUTOPUS_ROUND` environment setup in recreated panes** |
| gemini | major | Relaunching `args` providers uses the stale Round 1 prompt** |
| gemini | major | Missing environment setup during `SendLongText` fallback** |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.### VERDICT: REVISE

### Findings

**FINDING: [critical] Missing `AUTOPUS_ROUND` environment setup in recreated panes**
The plan proposes `recreatePane(ctx, cfg, pi)` without a `round` parameter. When a pane is recreated via `Terminal.SplitPane`, it starts a fresh shell. For TUI providers like `opencode`, `executeRound` explicitly skips `SendRoundEnvToPane` (to prevent pasting `export` commands into the chat UI). Because the new shell is fresh, it will lack the `AUTOPUS_ROUND` environment variable. Consequently, hook scripts (e.g., `hook-opencode-complete.ts`) will write to the wrong filename (missing the `-roundN` suffix), causing `waitAndCollectResults` / hook polling to time out and stall the debate. 
*Recommendation*: Update the `recreatePane` signature to include `round int`, and explicitly call `SendRoundEnvToPane` on the newly created pane *before* `launchInteractiveSessions`.

**FINDING: [major] Relaunching `args` providers uses the stale Round 1 prompt**
The plan reuses `launchInteractiveSessions` inside `recreatePane`. For providers configured with `InteractiveInput == "args"` (such as `gemini`), `launchInteractiveSessions` is hardcoded to append `cfg.Prompt` (the original Round 1 prompt) to the launch command. If the pane is recreated in Round 2+, the provider CLI will be launched and forced to process the Round 1 prompt *again*. Immediately after, `executeRound` will send the Round 2 rebuttal prompt via `SendLongText`, leading to duplicate processing, prompt collisions, and potential shell command injection if the provider exits unexpectedly.
*Recommendation*: Inside `recreatePane`, explicitly clear the prompt (e.g., `cfg.Prompt = ""`) before calling `launchInteractiveSessions`. This ensures the CLI starts cleanly in REPL mode and only processes the intended Round 2 prompt delivered subsequently via `SendLongText`.

**FINDING: [major] Missing environment setup during `SendLongText` fallback**
Requirement R6 specifies attempting pane recreation once if `SendLongText` fails on a previously validated surface. In the `executeRound` flow, `SendLongText` occurs *after* `SendRoundEnvToPane` has already been executed for `args` providers. If the pane is recreated at this later stage as a fallback, the new pane will completely miss the environment setup, leading to the same hook/collection failures described in Finding 1. 
*Recommendation*: By implementing the recommendation in Finding 1 (having `recreatePane` handle `SendRoundEnvToPane` internally), this issue is automatically resolved regardless of where in `executeRound` the recreation is triggered.

### Reasoning

The SPEC accurately diagnoses the problem with provider CLIs terminating and leaving stale cmux surfaces, and correctly identifies pane recreation as the optimal recovery path. Using `ReadScreen` as a lightweight validation proxy is also an elegant solution that prevents modifying core terminal interfaces.

However, the proposed `recreatePane` strategy fails to account for the context drift between a fresh Round 1 launch and a Round 2+ recreation. The existing `launchInteractiveSessions` function is tightly coupled to Round 1 semantics (specifically, passing `cfg.Prompt` as an argument). Furthermore, a fresh pane starts a new shell that inherently loses the `AUTOPUS_ROUND` environment variable. Without explicitly restoring this state *before* launching the CLI, hook-based providers will write invalid results, and `args`-based providers will incorrectly repeat Round 1. Addressing these state initialization gaps within `recreatePane` is required before implementation can proceed.


