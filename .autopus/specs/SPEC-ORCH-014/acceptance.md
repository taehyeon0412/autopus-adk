# SPEC-ORCH-014 수락 기준

## 시나리오

### S1: opencode TUI 모드 Launch

- Given: autopus.yaml에서 opencode provider의 PaneArgs가 `[-m, openai/gpt-5.4]`이고 InteractiveInput이 빈 문자열일 때
- When: `buildInteractiveLaunchCmd`가 opencode provider의 launch 커맨드를 생성하면
- Then: 생성된 커맨드는 `opencode -m openai/gpt-5.4`이어야 하며 `run` 서브커맨드를 포함하지 않아야 한다

### S2: Debate Round 1 프롬프트 전달

- Given: opencode가 cmux pane에서 TUI 모드로 실행되어 `> ` 프롬프트가 표시된 상태일 때
- When: interactive debate round 1에서 프롬프트가 전송되면
- Then: SendLongText를 통해 프롬프트가 opencode TUI 입력란에 전달되어야 한다 (CLI args 스킵 안 함)

### S3: Debate Round 2+ 세션 유지

- Given: opencode TUI 세션이 round 1을 완료하고 프롬프트(`> `)가 다시 표시된 상태일 때
- When: round 2 rebuttal 프롬프트가 전송되면
- Then: 기존 TUI 세션 내에서 SendLongText로 전달되고, 새 프로세스가 생성되지 않아야 한다

### S4: Hook 기반 완료 감지

- Given: opencode.json에 `autopus-result` 플러그인이 등록되고 AUTOPUS_SESSION_ID가 설정된 상태일 때
- When: opencode가 응답 생성을 완료하면
- Then: `/tmp/autopus/{session-id}/opencode-round{N}-done` 파일과 `opencode-round{N}-result.json` 파일이 생성되어야 한다

### S5: Screen Polling Fallback 완료 감지

- Given: hook mode가 비활성화되었고 opencode TUI가 응답을 완료한 상태일 때
- When: `waitForCompletion`이 opencode pane을 폴링하면
- Then: `^>\s*$` 패턴으로 프롬프트 재출현을 감지하고 2-phase consecutive match로 완료를 확인해야 한다

### S6: Config 마이그레이션

- Given: 기존 autopus.yaml에 opencode provider가 `interactive_input: "args"`, PaneArgs: `[run, -m, openai/gpt-5.4]`로 설정되어 있을 때
- When: `MigrateOrchestraConfig`가 실행되면
- Then: `interactive_input`이 빈 문자열로 변경되고 PaneArgs에서 `run`이 제거되어야 한다

### S7: 비인터랙티브 모드 유지

- Given: opencode provider의 Args가 `[run, -m, openai/gpt-5.4]`로 설정되어 있을 때
- When: 비인터랙티브(non-pane) 모드에서 opencode가 실행되면
- Then: `opencode run -m openai/gpt-5.4 'prompt'` 형태로 one-shot 실행이 유지되어야 한다
