# Review: SPEC-ORCH-017

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-03-29 18:16:03

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | REQ-003 is completely missing its implementation details. The text abruptly ends at "hook 스크립트를 확장하여:". It must specify how the script is expanded (e.g., how it loops/blocks waiting for the input file, reads the content, and passes it back to the agent). |
| gemini | critical | REQ-004 is completely missing its implementation steps. The text abruptly ends at "다음 순서를 실행한다:". It needs to define the specific sequence `executeRound` performs (e.g., wait for the `ready` signal, atomically write to `{provider}-round{N}-input.json`, and define timeout behavior). |
| gemini | major | REQ-005 fallback logic can cause an agent deadlock. If Orchestra falls back to `SendLongText` because of an "input 파일 쓰기 실패" (input file write failure) *after* the hook script has already created the `ready` signal, the agent will be stuck infinitely waiting for the `input.json` file. The SPEC must include a timeout mechanism inside the hook script, or a way for Orchestra to send an abort/cancel file to unblock the agent before attempting `SendLongText`. |
| gemini | minor | REQ-002 does not clarify the relationship between the existing `done` signal and the new `ready` signal. It should specify if `ready` is created concurrently with `done` or if it's a distinct phase indicating the hook has successfully entered a blocking read state for the next round. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.1. VERDICT: REVISE

2. For each issue found, write:
FINDING: [critical] REQ-003 is completely missing its implementation details. The text abruptly ends at "hook 스크립트를 확장하여:". It must specify how the script is expanded (e.g., how it loops/blocks waiting for the input file, reads the content, and passes it back to the agent).

FINDING: [critical] REQ-004 is completely missing its implementation steps. The text abruptly ends at "다음 순서를 실행한다:". It needs to define the specific sequence `executeRound` performs (e.g., wait for the `ready` signal, atomically write to `{provider}-round{N}-input.json`, and define timeout behavior).

FINDING: [major] REQ-005 fallback logic can cause an agent deadlock. If Orchestra falls back to `SendLongText` because of an "input 파일 쓰기 실패" (input file write failure) *after* the hook script has already created the `ready` signal, the agent will be stuck infinitely waiting for the `input.json` file. The SPEC must include a timeout mechanism inside the hook script, or a way for Orchestra to send an abort/cancel file to unblock the agent before attempting `SendLongText`.

FINDING: [minor] REQ-002 does not clarify the relationship between the existing `done` signal and the new `ready` signal. It should specify if `ready` is created concurrently with `done` or if it's a distinct phase indicating the hook has successfully entered a blocking read state for the next round.

3. Provide reasoning for your verdict:
The SPEC proposes a solid approach to fixing unreliable terminal input injections via bidirectional file-based IPC. However, the document is structurally incomplete—the core algorithmic steps for both the hook scripts (REQ-003) and the orchestra coordinator (REQ-004) are missing. Additionally, the fallback mechanism (REQ-005) needs state safety considerations to prevent the agent's REPL from hanging if Orchestra abandons the file IPC route halfway through. The SPEC must be revised to fill in these missing steps and address the potential deadlock.


