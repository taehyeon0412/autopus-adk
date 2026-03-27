# Review: SPEC-ORCH-002

**Verdict**: PASS
**Revision**: 0
**Date**: 2026-03-26 00:04:36

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | major | ** Agent Logging Locality (REQ-002/REQ-005): The spec assumes agents have local filesystem access to `/tmp/autopus-pipeline-{spec-id}.log`. If agents run in remote containers or different environments (as suggested by the `backend` and `workerbus` context), they cannot write to a local file on the CLI user's machine. |
| gemini | major | ** Missing Lifecycle Management: REQ-006 specifies closing panes on completion, but the spec does not address "unclean" exits (crashes, SIGINT). Without a cleanup mechanism for orphaned `cmux` sessions or stale log files in `/tmp`, the system will leak resources. |
| gemini | minor | ** Dashboard Command Scope (REQ-008): It is unclear if `auto pipeline dashboard` is an internal hidden command for `cmux` to call or a user-facing command. If user-facing, it needs to handle the case where the log file is missing or the pipeline has already finished. |
| gemini | suggestion | ** Log Format Standardization: REQ-002 mentions "structured logs" but doesn't define the schema. To ensure the dashboard (REQ-004/REQ-008) can parse them reliably, a shared JSON schema should be defined in `autopus-agent-protocol`. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.I have completed the review of **SPEC-ORCH-002 — cmux 모니터링 대시보드 — Agent Pipeline 실행 상태 시각화**.

1. **VERDICT**: **REVISE**

2. **FINDINGS**:
   - **FINDING: [major]** Agent Logging Locality (REQ-002/REQ-005): The spec assumes agents have local filesystem access to `/tmp/autopus-pipeline-{spec-id}.log`. If agents run in remote containers or different environments (as suggested by the `backend` and `workerbus` context), they cannot write to a local file on the CLI user's machine.
   - **FINDING: [major]** Missing Lifecycle Management: REQ-006 specifies closing panes on completion, but the spec does not address "unclean" exits (crashes, SIGINT). Without a cleanup mechanism for orphaned `cmux` sessions or stale log files in `/tmp`, the system will leak resources.
   - **FINDING: [minor]** Dashboard Command Scope (REQ-008): It is unclear if `auto pipeline dashboard` is an internal hidden command for `cmux` to call or a user-facing command. If user-facing, it needs to handle the case where the log file is missing or the pipeline has already finished.
   - **FINDING: [suggestion]** Log Format Standardization: REQ-002 mentions "structured logs" but doesn't define the schema. To ensure the dashboard (REQ-004/REQ-008) can parse them reliably, a shared JSON schema should be defined in `autopus-agent-protocol`.

3. **REASONING**:
   The specification successfully outlines the visual goal of providing a multiplexed terminal view for pipelines. However, it treats the entire system as a local-only process. Given that the provided context includes a complex `backend` with `workerbus` and `eventbus`, the assumption that agents can directly write to a `/tmp` file on the user's terminal is a significant architectural mismatch. The data flow should likely involve the `backend` streaming events to the CLI, which then writes to the local log for `tail -f` and the dashboard. Additionally, the lack of robust cleanup logic for terminal panes is a common source of "terminal pollution" in CLI tools.


