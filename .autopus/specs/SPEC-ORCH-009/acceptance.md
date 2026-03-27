# SPEC-ORCH-009 수락 기준

## 시나리오

### S1: Brainstorm debate 기본 라운드 활성화

- Given: brainstorm 커맨드가 `--strategy debate`로 실행됨
- When: `--rounds` 플래그가 명시되지 않음
- Then: `resolveRounds("debate", 0)`이 호출되어 rounds=2가 반환됨
- And: `runOrchestraCommand`에 rounds=2가 전달됨
- And: `cfg.DebateRounds`가 2로 설정됨

### S2: Debate 라우팅 활성화 확인

- Given: brainstorm 경로에서 cfg.DebateRounds=2가 설정됨
- When: `RunInteractivePaneOrchestra`가 호출됨
- Then: `interactive.go:26`의 `cfg.Strategy == StrategyDebate && cfg.DebateRounds >= 2` 조건이 true
- And: `runInteractiveDebate()`가 호출됨

### S3: 멀티턴 Rebuttal 발생 확인

- Given: debate 2라운드가 설정됨 (DebateRounds=2)
- When: Round 1이 완료되고 Round 2가 시작됨
- Then: `executeRound()`에서 `buildRebuttalPrompt()`를 호출하여 교차 응답 주입
- And: 각 프로바이더가 다른 프로바이더의 Round 1 응답을 포함한 프롬프트를 수신

### S4: ANSI 이스케이프 시퀀스 제거

- Given: ReadScreen 결과에 `\x1b[31mError\x1b[0m` 같은 ANSI 색상 코드가 포함됨
- When: `SanitizeScreenOutput()`이 호출됨
- Then: `Error`만 남고 이스케이프 시퀀스는 제거됨

### S5: OSC 시퀀스 제거

- Given: ReadScreen 결과에 `\x1b]0;window title\x07` 같은 OSC 시퀀스가 포함됨
- When: `SanitizeScreenOutput()`이 호출됨
- Then: OSC 시퀀스가 완전히 제거됨

### S6: 상태바 라인 제거

- Given: ReadScreen 결과에 tmux 상태바 패턴이 포함됨
- When: `SanitizeScreenOutput()`이 호출됨
- Then: 상태바 라인이 제거되고 응답 본문만 유지됨

### S7: Markdown 인용문 보존

- Given: 프로바이더 응답에 `> This is a quote` 같은 markdown 인용문이 포함됨
- When: `SanitizeScreenOutput()`이 호출됨
- Then: markdown 인용문은 그대로 보존됨 (프롬프트 패턴으로 오인 제거되지 않음)

### S8: 연속 빈 행 정리

- Given: ReadScreen 결과에 3개 이상의 연속 빈 행이 포함됨
- When: `SanitizeScreenOutput()`이 호출됨
- Then: 연속 빈 행이 최대 1개로 축소됨

### S9: --rounds 플래그 동작

- Given: brainstorm 커맨드에 `--rounds 3 --strategy debate`가 전달됨
- When: 커맨드가 실행됨
- Then: `resolveRounds("debate", 3)`이 호출되어 rounds=3이 반환됨
- And: 3라운드 debate가 실행됨

### S10: --rounds 없이 non-debate 전략

- Given: brainstorm 커맨드가 `--strategy consensus`로 실행됨 (rounds 미지정)
- When: `resolveRounds("consensus", 0)`이 호출됨
- Then: rounds=0이 반환되어 단일 실행 모드 유지

### S11: 기존 테스트 호환성

- Given: SPEC-ORCH-009 변경이 적용됨
- When: 기존 테스트 스위트가 실행됨
- Then: `orchestra_brainstorm_test.go`, `interactive_debate_test.go`, `round_signal_test.go` 모두 통과
