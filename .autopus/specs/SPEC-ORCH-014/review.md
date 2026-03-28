# Review: SPEC-ORCH-014

**Verdict**: PASS
**Revision**: 0
**Date**: 2026-03-28 18:58:44

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | Auto-derivation logic conflict.** The requirement to set `interactive_input` to an empty string (REQ-001, REQ-009) critically conflicts with existing configuration parsing logic. In `internal/cli/orchestra_config.go` (lines 46-48), if `InteractiveInput` is empty `""` and `PromptViaArgs` is `true`, the system automatically overrides `InteractiveInput` to `"args"`. Because opencode still requires `PromptViaArgs: true` for non-interactive pipeline operations, setting `interactive_input` to `""` will simply result in the system forcing it back to `"args"`. Consequently, this will re-enable the skip logic (violating REQ-005) and prevent TUI launch. |
| gemini | major | TUI launch failure due to subcommand retention.** REQ-002 and REQ-003 state that the `run` subcommand must be removed for TUI mode execution (`opencode -m ...`). While `pkg/orchestra/interactive_launch.go` successfully strips the `run` subcommand when `InteractiveInput != "args"`, this mechanism will fail because of the auto-derivation issue described above. `InteractiveInput` will be evaluated as `"args"`, causing opencode to retain the `run` subcommand and mistakenly execute in non-interactive mode. |
| gemini | suggestion | Clarify `PaneArgs` modification.** The SPEC should explicitly clarify whether the `run` subcommand should be removed from the hardcoded `PaneArgs` in `pkg/config/migrate.go` (`defaultProviderEntries`) or if the implementation relies entirely on the dynamic stripping logic inside `buildInteractiveLaunchCmd`. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.**VERDICT:** REVISE

**FINDINGS:**

*   **FINDING: [critical] Auto-derivation logic conflict.** The requirement to set `interactive_input` to an empty string (REQ-001, REQ-009) critically conflicts with existing configuration parsing logic. In `internal/cli/orchestra_config.go` (lines 46-48), if `InteractiveInput` is empty `""` and `PromptViaArgs` is `true`, the system automatically overrides `InteractiveInput` to `"args"`. Because opencode still requires `PromptViaArgs: true` for non-interactive pipeline operations, setting `interactive_input` to `""` will simply result in the system forcing it back to `"args"`. Consequently, this will re-enable the skip logic (violating REQ-005) and prevent TUI launch.
*   **FINDING: [major] TUI launch failure due to subcommand retention.** REQ-002 and REQ-003 state that the `run` subcommand must be removed for TUI mode execution (`opencode -m ...`). While `pkg/orchestra/interactive_launch.go` successfully strips the `run` subcommand when `InteractiveInput != "args"`, this mechanism will fail because of the auto-derivation issue described above. `InteractiveInput` will be evaluated as `"args"`, causing opencode to retain the `run` subcommand and mistakenly execute in non-interactive mode.
*   **FINDING: [suggestion] Clarify `PaneArgs` modification.** The SPEC should explicitly clarify whether the `run` subcommand should be removed from the hardcoded `PaneArgs` in `pkg/config/migrate.go` (`defaultProviderEntries`) or if the implementation relies entirely on the dynamic stripping logic inside `buildInteractiveLaunchCmd`.

**REASONING:**

While the hook integration (REQ-007, REQ-008) and prompt pattern detection (`> `) align perfectly with the current codebase (`DefaultCompletionPatterns` in `pkg/orchestra/types.go`), the core mechanism for enabling TUI mode is technically flawed. 

Relying on an empty string for `interactive_input` to disable `"args"` mode is incompatible with the existing fallback logic that prioritizes `PromptViaArgs`. To make this SPEC feasible, it must either propose changing the auto-derivation logic in `internal/cli/orchestra_config.go`, or mandate using an explicit non-empty value (such as `"stdin"` or `"tui"`) for `interactive_input` to represent the TUI mode state and bypass the `"args"` fallback. Until this architectural conflict is resolved, the requirements cannot be successfully implemented.


