# Review: SPEC-ORCHCFG-002

**Verdict**: PASS
**Revision**: 0
**Date**: 2026-03-30 10:02:48

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | minor | The SPEC explicitly mandates removing the `MigrateOpencodeToTUI` step (REQ-003) but omits mentioning the existing `MigrateCodexToOpencode` step (Migration 1.5 in `pkg/config/migrate.go`). The SPEC should explicitly state to remove or replace `MigrateCodexToOpencode` to prevent conflicting migrations from running sequentially and potentially corrupting the user's provider settings. |
| gemini | suggestion | For REQ-007 ("preserving the user's custom model settings"), it would be beneficial to clarify the handling of the provider prefix. `opencode` conventionally used the `openai/` prefix (e.g., `openai/gpt-5.4`), while REQ-001 specifies the `codex` default without the prefix (`gpt-5.4`). The SPEC should specify whether the migration logic needs to strip the `openai/` prefix from user-customized models during conversion. |
| gemini | suggestion | For REQ-006, along with retaining the `Ask anything` completion pattern, the SPEC should explicitly mention retaining `content/hooks/hook-opencode-complete.ts`. This ensures bidirectional IPC remains intact for users who opt to manually configure and use `opencode` in TUI mode. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.VERDICT: PASS

The SPEC is structurally sound, highly feasible, and correctly maps to the existing configuration and orchestra system. The proposed migration appropriately reverses the previous `opencode` default and returns to `codex` while gracefully handling TUI characteristics. 

Here are a few findings to refine the implementation details:

FINDING: [minor] The SPEC explicitly mandates removing the `MigrateOpencodeToTUI` step (REQ-003) but omits mentioning the existing `MigrateCodexToOpencode` step (Migration 1.5 in `pkg/config/migrate.go`). The SPEC should explicitly state to remove or replace `MigrateCodexToOpencode` to prevent conflicting migrations from running sequentially and potentially corrupting the user's provider settings.

FINDING: [suggestion] For REQ-007 ("preserving the user's custom model settings"), it would be beneficial to clarify the handling of the provider prefix. `opencode` conventionally used the `openai/` prefix (e.g., `openai/gpt-5.4`), while REQ-001 specifies the `codex` default without the prefix (`gpt-5.4`). The SPEC should specify whether the migration logic needs to strip the `openai/` prefix from user-customized models during conversion.

FINDING: [suggestion] For REQ-006, along with retaining the `Ask anything` completion pattern, the SPEC should explicitly mention retaining `content/hooks/hook-opencode-complete.ts`. This ensures bidirectional IPC remains intact for users who opt to manually configure and use `opencode` in TUI mode.


