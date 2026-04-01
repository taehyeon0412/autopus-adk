# Review: SPEC-MULTIPLATFORM-001

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-04-01 14:10:47

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | REQ-029 (Go file length limit will break CI)** |
| gemini | major | REQ-022 & REQ-023 (Incorrect source directory paths)** |
| gemini | minor | REQ-013 (Incorrect GEMINI.md import syntax)** |
| gemini | minor | REQ-002 (Codex TOML agent format)** |
| gemini | suggestion | REQ-018 (Accurate mapping for Gemini Subagents)** |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.**1. VERDICT: REVISE**

**2. FINDINGS:**

*   **FINDING: [critical] REQ-029 (Go file length limit will break CI)**
    The requirement mandates that CI verify *all* Go files are 300 lines or less ("모든 Go 파일이 300줄 이하임을 CI에서 검증한다"). However, a quick scan of the existing codebase reveals that several existing files already significantly exceed this limit (e.g., `pkg/orchestra/session.go` has 425 lines, `pkg/orchestra/surface_helpers_test.go` has 399 lines, `pkg/adapter/claude/claude_router_test.go` has 382 lines). Enforcing this in CI will immediately break the pipeline upon implementation unless all existing oversized files are refactored in the same PR.
    *Recommendation*: Revise the requirement to apply only to *new or modified* files, increase the threshold to accommodate existing files, or explicitly include a refactoring task in the SPEC.

*   **FINDING: [major] REQ-022 & REQ-023 (Incorrect source directory paths)**
    The requirements state to convert shared rule templates and agent metadata from `templates/shared/rules/` and `templates/shared/agents/`. However, those directories do not exist in the current project structure. The `templates/shared/` directory only contains `.tmpl` files at its root. The actual Markdown source files for agents and rules are located in `content/agents/` and `content/rules/` (as evidenced by `content/embed.go` and the directory tree).
    *Recommendation*: Update the paths in the SPEC to reference the correct source directories (`content/rules/` and `content/agents/`).

*   **FINDING: [minor] REQ-013 (Incorrect GEMINI.md import syntax)**
    The requirement states that rule files should be generated in `.gemini/rules/autopus/` and referenced in `GEMINI.md` using `@import` ("GEMINI.md에서 `@import`로 참조한다"). Gemini CLI's Memory Import Processor does not use a literal `@import` directive; rather, it uses the `@path/to/file.md` syntax (e.g., `@./rules/autopus/file.md`).
    *Recommendation*: Clarify the phrasing to specify the use of the Gemini CLI native `@<path>` syntax.

*   **FINDING: [minor] REQ-002 (Codex TOML agent format)**
    The requirement states to create TOML files for core agents in `.codex/agents/`. Unlike Gemini or Claude which have established CLI configuration schemas for agents/tools, standard GitHub Copilot/Codex does not have a widely recognized native `.codex/agents/` TOML architecture. 
    *Recommendation*: Ensure this specific structure is intentional and aligns with the custom abstract layer or specific Codex-compatible runner being targeted by this project.

*   **FINDING: [suggestion] REQ-018 (Accurate mapping for Gemini Subagents)**
    The requirement correctly maps Claude's `Agent()` to Gemini's `@agent` tool pattern. This is an excellent, native fit for Gemini CLI's subagent architecture.

**3. REASONING:**
The SPEC introduces an excellent, feasible multi-platform generation architecture that correctly leverages the specific native schemas of Gemini CLI (e.g., `settings.json` hooks for `BeforeTool`/`AfterTool`, `.gemini/commands/` TOML namespacing, and `mcpServers` configurations). However, the inclusion of a strict, retroactive CI line-limit rule (REQ-029) and referencing non-existent template directories (REQ-022, REQ-023) makes the current specification technically un-implementable without immediate test failures or missing file errors. Revising these operational flaws will yield a robust, pass-ready SPEC.


