# SPEC-ORCH-013 수락 기준

## 시나리오

### S1: Judge가 debate timeout 소진 후에도 정상 실행

- Given: debate 3라운드가 cfg.TimeoutSeconds(120s)의 대부분을 소진한 상태
- When: judge round가 시작될 때
- Then: judge는 `context.Background()` 기반의 독립 context로 실행되어 "판정: 없음" 없이 정상 verdict를 반환한다

### S2: perRoundTimeout이 judge budget을 침범하지 않음

- Given: cfg.TimeoutSeconds=120, DebateRounds=3
- When: perRoundTimeout(120, 3)이 계산될 때
- Then: judge budget(최소 60초)을 먼저 차감하고 나머지(60초)를 3라운드에 분배. per-round=20초이나 45초 floor 적용으로 각 round에 45초 할당. judge timeout은 별도로 최소 60초가 보장된다

### S3: Claude pane 이전 라운드 prompt로 인한 false-positive 방지

- Given: claude pane에 이전 라운드의 `> ` prompt가 이미 화면에 보이는 상태
- When: 새 prompt가 전송되고 waitForCompletion이 시작될 때
- Then: screen content가 baseline과 변경될 때까지 prompt match를 하지 않아 empty output 수집이 발생하지 않는다

### S4: Screen content 변화 후 정상 완료 감지

- Given: baseline screen이 저장된 상태에서 AI가 응답을 출력 중
- When: AI 응답이 완료되고 prompt `> `가 재등장할 때
- Then: screen이 baseline과 다르고 + 2-phase consecutive match가 통과하여 정상 완료로 판정한다

### S5: Diff 섹션에서 MCP noise가 제거됨

- Given: gemini provider output에 "MCP issues detected. Run /mcp list for status." 문자열이 포함
- When: FormatDebate가 "주요 차이점" 섹션을 생성할 때
- Then: diff 결과에 MCP noise 문자열이 포함되지 않는다

### S6: Diff 섹션에서 ANSI escape가 제거됨

- Given: provider output에 ANSI color escape sequence가 포함
- When: findDifferences가 줄 비교를 수행할 때
- Then: ANSI escape가 제거된 clean text 기준으로 비교되어 escape 시퀀스가 diff에 노출되지 않는다

### S7: Shell login banner 필터링

- Given: opencode pane output에 "Last login: Fri Mar 28 10:30:00 on ttys001" 라인이 포함
- When: cleanScreenOutput이 적용될 때
- Then: 해당 라인이 필터링되어 최종 output에 포함되지 않는다

### S8: User@host prompt 필터링

- Given: opencode pane output에 "bitgapnam@Mac ~ % " 라인이 포함
- When: isPromptLine이 해당 라인을 검사할 때
- Then: cliNoisePatterns 매칭으로 해당 라인이 noise로 분류된다

### S9: 기존 테스트 회귀 없음

- Given: R1-R4 변경이 모두 적용된 상태
- When: `go test ./pkg/orchestra/...`를 실행할 때
- Then: 모든 기존 테스트가 통과한다
