# Review: SPEC-ORCH-005

**Verdict**: PASS
**Revision**: 0
**Date**: 2026-03-26 15:17:54

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | major | ** REQ-005 is incomplete; the list of provider-specific pane relay arguments is missing after "프로바이더별 pane relay 인수:". Without these, the default interactive behavior for each provider cannot be validated. |
| gemini | major | ** REQ-009 is incomplete; the rules for keeping or closing previous panes during sequential execution are not defined after the colon. This is critical for managing terminal UI clutter and resource cleanup. |
| gemini | minor | ** Feasibility of sequential execution: The SPEC relies on the existing sentinel mechanism (`echo sentinel >> output` in `buildPaneCommand`), which only executes after the provider process exits. If "interactive mode" implies the process stays open (common for some CLIs), the sequential relay will hang indefinitely. The SPEC should clarify that providers must be invoked in a way that ensures exit after prompt completion (e.g., via heredoc EOF as currently implemented in `pane_runner.go`). |
| gemini | suggestion | ** To maintain consistency and avoid logic duplication, `relay_pane.go` (REQ-011) should explicitly reuse `buildRelayPrompt` and `agenticArgs` from `relay.go`. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.I will start by exploring the codebase to understand the current implementation of relay mode and pane execution, specifically focusing on `internal/cli/runner.go` and `internal/cli/pane_runner.go`.
I will search for `runner.go` and `pane_runner.go` to locate them within the project.
I will search for `pane_runner.go` to confirm its location.
I will read `pkg/orchestra/runner.go` to examine the current relay pane fallback implementation.
I will read `pkg/orchestra/pane_runner.go` to understand how `RunPaneOrchestra` is implemented.
I will check for the existence of `pkg/orchestra/provider.go` or other relevant files to find the default interactive mode arguments for providers.
I will read `pkg/orchestra/types.go` and `pkg/orchestra/relay.go` to find the default interactive mode arguments and other relevant details.
I will read `pkg/orchestra/relay.go` to understand the current implementation of relay mode.
I will search for "REQ-005" in the repository to find the full text of the specification and confirm the provider-specific pane relay arguments.
I will check `pkg/detect` to see if there are any files defining default provider arguments.
I will check `internal/cli/orchestra_config.go` for provider configuration details.
I will check the `.autopus/specs/` directory to see if there are any existing specification files that follow a similar format.
I will list the contents of the `.autopus` directory to check for any hidden or ignored specification files.
I will use `ls -R .autopus` to see if there are any files within the `.autopus` directory.
I will use `glob` to search for any Markdown files within the `.autopus/specs/` directory.
I will search for the text of REQ-005 in the codebase to see if it was copied from an existing document.
I will check for the `pkg/terminal/` directory to see the implementations of the `Terminal` interface.
I will read `pkg/terminal/tmux.go` to understand the `tmux` implementation of the `Terminal` interface.
I will read `pkg/orchestra/relay_test.go` to see if there are any existing tests for pane relay and to understand the testing patterns used.
I will check the line count of `pkg/orchestra/pane_runner.go` to see if it is close to the 300-line limit.
I will use `read_file` to determine the line count of `pkg/orchestra/pane_runner.go`.
I will search for the definition of `sanitizeProviderName` in the codebase.
1. **VERDICT**: REVISE

2. **FINDINGS**:
   - **FINDING: [major]** REQ-005 is incomplete; the list of provider-specific pane relay arguments is missing after "프로바이더별 pane relay 인수:". Without these, the default interactive behavior for each provider cannot be validated.
   - **FINDING: [major]** REQ-009 is incomplete; the rules for keeping or closing previous panes during sequential execution are not defined after the colon. This is critical for managing terminal UI clutter and resource cleanup.
   - **FINDING: [minor]** Feasibility of sequential execution: The SPEC relies on the existing sentinel mechanism (`echo sentinel >> output` in `buildPaneCommand`), which only executes after the provider process exits. If "interactive mode" implies the process stays open (common for some CLIs), the sequential relay will hang indefinitely. The SPEC should clarify that providers must be invoked in a way that ensures exit after prompt completion (e.g., via heredoc EOF as currently implemented in `pane_runner.go`).
   - **FINDING: [suggestion]** To maintain consistency and avoid logic duplication, `relay_pane.go` (REQ-011) should explicitly reuse `buildRelayPrompt` and `agenticArgs` from `relay.go`.

3. **REASONING**:
   The specification is fundamentally sound in its architectural direction—routing `StrategyRelay` through `RunPaneOrchestra` and separating the logic into `relay_pane.go` is a clean approach that preserves the existing parallel execution path for other strategies. However, the document provided is physically incomplete (truncated at REQ-005 and REQ-009), which prevents a full review of the provider configurations and the pane lifecycle management. Additionally, while the sentinel mechanism is reused, its dependency on process termination may conflict with the "interactive" nature of pane mode if not explicitly managed (e.g., ensuring the shell command string forces an exit). Once the truncated sections are restored and the exit behavior is clarified, the SPEC will be ready for implementation.


