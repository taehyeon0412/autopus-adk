# Review: SPEC-TEAMPANE-001

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-03-26 14:56:20

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | SPEC R4와 다이어그램의 분할 방향(Horizontal vs Vertical)이 `pkg/terminal`의 정의와 상충합니다. `terminal.Horizontal`은 좌우(side-by-side) 분할을 의미하지만, SPEC은 "하단에 새 패널 생성" 및 수직 스택 다이어그램을 제시하고 있습니다. 수직 스택을 원한다면 `Vertical`을 사용해야 합니다. |
| gemini | major | Plan T1에서 shell-escape 유틸리티를 자체 재구현(duplicate)하도록 설계되어 있습니다. 이는 `pkg/orchestra/pane_shell.go`의 기존 검증된 로직과 중복되므로, 이를 `pkg/terminal` 등 공통 패키지로 이동하여 공유하는 것이 바람직합니다. |
| gemini | major | `PipelineMonitor` 인터페이스의 정의 위치가 부적절합니다. Plan T1은 `team_pane.go`에 정의하도록 제안하지만, 기존 `MonitorSession`이 이를 구현하고 파이프라인에서 공통으로 사용하려면 `pkg/pipeline/monitor.go` 등 중앙 위치에 정의해야 순환 참조를 방지할 수 있습니다. |
| gemini | minor | `TmuxAdapter.Close()`가 `kill-session`을 사용하므로, 만약 모니터가 기존 세션 내의 새 윈도우에서 실행 중일 경우 세션 전체를 종료시킬 위험이 있습니다. 윈도우 기반 실행 시 `kill-window`를 고려해야 합니다. |
| gemini | suggestion | `UpdateTeammate(role, status)` 메서드 대신, `MonitorSession`과 `TeamMonitorSession`이 공통으로 만족할 수 있는 더 일반적인 `UpdateAgent(agentName, role, status)` 형태의 인터페이스를 정의하는 것이 확장성 면에서 유리합니다. |
| gemini | minor | R4의 "순차적 분할" 전략에서 매번 새로 생성된 패널이 '활성' 상태가 된다는 가정이 명시되어야 합니다. `SplitPane`이 항상 새로운 패널 ID를 반환하고 해당 패널을 활성화한다고 가정할 때만 제시된 중첩 레이아웃이 유효합니다. |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.I will search for the content of `SPEC-TEAMPANE-001` in the `.autopus/specs/` directory to review it.
I will search for any files containing `SPEC-TEAMPANE-001` or "Agent Teams 터미널 패널 시각화" in the codebase to find the spec content.
I will search for "Agent Teams" or "터미널 패널 시각화" to find the SPEC document content.
I will use `glob` to find any SPEC documents in the `.autopus/specs/` directory.
I will search for "SPEC-TEAMPANE" in the entire project to find the SPEC content.
I will list the `.autopus/specs/` directory again, ensuring that I see all files, including those that might be ignored.
I will list the content of the `.autopus/specs/SPEC-TEAMPANE-001` directory to find the SPEC document.
I will list the content of the `.autopus/specs/SPEC-TEAMPANE-001` directory, making sure to show all files.
I will read the `spec.md` file in the `.autopus/specs/SPEC-TEAMPANE-001` directory to review the SPEC.
I will check the `.gitignore` file to see what it ignores.
I will read the content of the `spec.md` file using `cat` to bypass the ignore patterns.
I will use `grep_search` with `no_ignore: true` to read the content of the `spec.md` file.
I will read the `acceptance.md` and `plan.md` files in the `.autopus/specs/SPEC-TEAMPANE-001` directory to fully understand the SPEC and its implementation plan.
I will read the `plan.md` file in the `.autopus/specs/SPEC-TEAMPANE-001` directory to review the implementation plan.
I will search for the `DetectTerminal` function and the `Terminal` interface in the codebase to verify their existence and signatures.
I will find the `Terminal` interface definition to check its methods.
I will search for `pane_runner.go` and check its implementation of `SplitPane`.
I will read `pkg/orchestra/pane_runner.go` to understand how it uses `SplitPane`.
I will search for the `cmux_adapter.go` file to see how it implements the `SplitPane` method.
I will list the files in the `pkg/terminal/` directory to find the terminal adapters.
I will read `pkg/terminal/cmux.go` to see how `SplitPane` is implemented for `cmux`.
I will read `pkg/terminal/tmux.go` to see how `SplitPane` is implemented for `tmux`.
I will search for the `dashboard.go` file to see how `RenderDashboard` is implemented.
I will read `pkg/pipeline/dashboard.go` to understand how `RenderDashboard` works.
I will search for `EventAgentSpawn` and `EventAgentDone` in the codebase to verify if they are existing event types.
I will search for the `MonitorSession` struct to find its location and current implementation.
I will read the rest of the `pkg/pipeline/monitor.go` file to see the `MonitorSession` methods.
I will search for the shell-escape utility to see where it is currently defined and used.
1. VERDICT: REVISE

2. FINDINGS:
- FINDING: [critical] SPEC R4와 다이어그램의 분할 방향(Horizontal vs Vertical)이 `pkg/terminal`의 정의와 상충합니다. `terminal.Horizontal`은 좌우(side-by-side) 분할을 의미하지만, SPEC은 "하단에 새 패널 생성" 및 수직 스택 다이어그램을 제시하고 있습니다. 수직 스택을 원한다면 `Vertical`을 사용해야 합니다.
- FINDING: [major] Plan T1에서 shell-escape 유틸리티를 자체 재구현(duplicate)하도록 설계되어 있습니다. 이는 `pkg/orchestra/pane_shell.go`의 기존 검증된 로직과 중복되므로, 이를 `pkg/terminal` 등 공통 패키지로 이동하여 공유하는 것이 바람직합니다.
- FINDING: [major] `PipelineMonitor` 인터페이스의 정의 위치가 부적절합니다. Plan T1은 `team_pane.go`에 정의하도록 제안하지만, 기존 `MonitorSession`이 이를 구현하고 파이프라인에서 공통으로 사용하려면 `pkg/pipeline/monitor.go` 등 중앙 위치에 정의해야 순환 참조를 방지할 수 있습니다.
- FINDING: [minor] `TmuxAdapter.Close()`가 `kill-session`을 사용하므로, 만약 모니터가 기존 세션 내의 새 윈도우에서 실행 중일 경우 세션 전체를 종료시킬 위험이 있습니다. 윈도우 기반 실행 시 `kill-window`를 고려해야 합니다.
- FINDING: [suggestion] `UpdateTeammate(role, status)` 메서드 대신, `MonitorSession`과 `TeamMonitorSession`이 공통으로 만족할 수 있는 더 일반적인 `UpdateAgent(agentName, role, status)` 형태의 인터페이스를 정의하는 것이 확장성 면에서 유리합니다.
- FINDING: [minor] R4의 "순차적 분할" 전략에서 매번 새로 생성된 패널이 '활성' 상태가 된다는 가정이 명시되어야 합니다. `SplitPane`이 항상 새로운 패널 ID를 반환하고 해당 패널을 활성화한다고 가정할 때만 제시된 중첩 레이아웃이 유효합니다.

3. REASONING:
- SPEC-TEAMPANE-001은 에이전트 팀의 가시성을 높이기 위한 중요한 개선 사항이며, 순차적 분할(sequential split) 전략을 통한 기술적 제약 극복 방안은 타당합니다.
- 그러나 가장 핵심적인 레이아웃 엔진의 분할 방향(Horizontal vs Vertical)이 코드베이스의 상수 정의와 정반대로 기술되어 있어, 이대로 구현할 경우 의도하지 않은(매우 좁은 폭의 6개 좌우 패널) 결과가 발생할 위험이 큽니다.
- 또한, 인터페이스 정의 위치와 유틸리티 중복 문제는 유지보수성 및 코드 품질 관점에서 반드시 수정이 필요합니다. 특히 `PipelineMonitor`를 통한 다형성 구현이 핵심 전략인 만큼, 패키지 구조 설계를 명확히 해야 합니다.
- 이러한 설계상 오류와 중복을 수정한 후 구현 단계로 넘어가는 것을 권장합니다.


