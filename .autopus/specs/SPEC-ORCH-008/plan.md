# SPEC-ORCH-008 구현 계획

## 실행 전략

3개 Phase로 분리하여 순차 실행. Phase 1 내부의 독립 태스크는 병렬 가능.
총 신규 코드 약 500줄, 수정 약 90줄. 300줄 파일 크기 제한 준수.

---

## Phase 1: 코어 인프라 (P0 R1-R7)

### T1: 라운드 스코프 시그널 관리 — `round_signal.go` 신규

**소유 파일**: `pkg/orchestra/round_signal.go`
**에이전트**: executor-A (독립 실행 가능)

- [ ] `RoundSignalName(provider, round)` — 라운드별 done/result 파일명 생성
  - round=0 또는 round=1 && totalRounds=1 → 기존 `{provider}-done` 형식 (하위 호환)
  - round >= 1 && totalRounds >= 2 → `{provider}-round{N}-done`, `{provider}-round{N}-result.json`
- [ ] `CleanRoundSignals(sessionDir, providers, round)` — 해당 라운드의 done 시그널 삭제, result 보존
- [ ] `SetRoundEnv(round)` — `AUTOPUS_ROUND` 환경변수 설정
- [ ] `SendRoundEnvToPane(ctx, term, paneID, round)` — pane에 `export AUTOPUS_ROUND=N` 전송
- [ ] 줄 수 목표: < 100줄

### T2: hook_signal.go 라운드 파라미터 확장

**소유 파일**: `pkg/orchestra/hook_signal.go`
**에이전트**: executor-A (T1과 동일 파일 도메인, 순차 실행)

- [ ] `WaitForDone()` 시그니처에 `round int` 옵션 파라미터 추가 (variadic 또는 opts struct)
- [ ] round가 지정되면 `RoundSignalName()` 활용하여 done 파일 경로 결정
- [ ] `ReadResult()`에도 동일하게 round 파라미터 추가
- [ ] 기존 호출부(hook_watcher.go)는 round=0으로 동작하여 하위 호환 유지

### T3: Interactive Debate 루프 — `interactive_debate.go` 신규

**소유 파일**: `pkg/orchestra/interactive_debate.go`
**에이전트**: executor-B (T1/T2 완료 후 실행)
**의존**: T1, T2

- [ ] `runInteractiveDebate(ctx, cfg, panes, hookSession)` 메인 함수
  - Returns: `([]ProviderResponse, [][]ProviderResponse, error)` — 최종 응답 + 라운드 히스토리
- [ ] Round 1: 원본 프롬프트 전송 → hook 결과 수집 (라운드 스코프 시그널 사용)
- [ ] Round 2..N 루프:
  1. `CleanRoundSignals(prev round)`
  2. `SendRoundEnvToPane(ctx, term, pane, round)`
  3. `pollUntilPrompt()` — pane 입력 대기 확인
  4. 각 프로바이더별 `buildRebuttalPrompt(original, otherResponses)` 생성
  5. pane에 rebuttal 프롬프트 전송 (sendPrompts 패턴 재사용)
  6. hook 결과 수집 (라운드 스코프)
  7. 라운드 히스토리에 저장
- [ ] 라운드별 타임아웃: 전체 타임아웃 / 라운드 수
- [ ] 프로바이더 1개 실패 시 해당 프로바이더 skip, 나머지 계속
- [ ] 줄 수 목표: < 200줄

### T4: interactive.go에 debate 분기 추가

**소유 파일**: `pkg/orchestra/interactive.go`
**에이전트**: executor-B (T3와 함께)

- [ ] Step 5 (sendPrompts) 직전에 debate 전략 체크:
  ```go
  if cfg.Strategy == StrategyDebate && cfg.DebateRounds >= 2 {
      // delegate to interactive debate loop
  }
  ```
- [ ] `runInteractiveDebate()` 호출 후 결과를 OrchestraResult로 조립
- [ ] 변경 약 15줄

### T5: types.go에 RoundHistory 추가

**소유 파일**: `pkg/orchestra/types.go`
**에이전트**: executor-A 또는 executor-B (간단, 어느 에이전트든 가능)

- [ ] `OrchestraResult`에 `RoundHistory [][]ProviderResponse` 필드 추가
- [ ] 변경 약 5줄

### T6: `--rounds` CLI 플래그 추가

**소유 파일**: `internal/cli/orchestra.go`
**에이전트**: executor-C (독립 실행 가능)

- [ ] `--rounds N` IntVar 플래그 추가 (multi, orchestra 서브커맨드 모두)
- [ ] 검증: 1-10 범위, `--strategy debate` 필수
- [ ] 미지정 시 debate 전략 기본값 2 (기존 기본값 1에서 변경)
- [ ] `OrchestraConfig.DebateRounds`에 매핑
- [ ] 변경 약 20줄

---

## Phase 2: Hook 스크립트 라운드 인식 (P0 R5)

### T7: Hook 스크립트 AUTOPUS_ROUND 분기

**소유 파일**: `content/hooks/hook-claude-stop.sh`, `hook-gemini-afteragent.sh`, `hook-opencode-complete.ts`
**에이전트**: executor-C (독립 실행 가능, Phase 1과 병렬 가능)

- [ ] Shell hooks (claude, gemini): `AUTOPUS_ROUND` 체크 후 파일명 분기
  ```bash
  if [ -n "$AUTOPUS_ROUND" ]; then
    RESULT="${PROVIDER}-round${AUTOPUS_ROUND}-result.json"
    DONE="${PROVIDER}-round${AUTOPUS_ROUND}-done"
  else
    RESULT="${PROVIDER}-result.json"
    DONE="${PROVIDER}-done"
  fi
  ```
- [ ] TypeScript hook (opencode): `process.env.AUTOPUS_ROUND` 체크 후 동일 분기
- [ ] 각 파일 변경 약 8줄

---

## Phase 3: 부가 기능 (P1)

### T8: Judge 라운드 Interactive 실행 (R8)

**소유 파일**: `pkg/orchestra/interactive_debate.go` (확장)
**에이전트**: executor-B
**의존**: T3

- [ ] 모든 라운드 완료 후 judge 프로바이더 pane에 `buildJudgmentPrompt()` 전송
- [ ] Judge가 참여 프로바이더 중 하나면 pane 재사용, 별도면 추가 pane 생성
- [ ] hook 시그널로 수집 (round = "judge")

### T9: 라운드 진행 상태 표시 (R9)

**소유 파일**: `pkg/orchestra/interactive_debate.go` (확장)
**에이전트**: executor-B (T8과 함께)

- [ ] 라운드 시작: `[Round N/M] 시작...`
- [ ] 프로바이더 완료: `[Round N/M] {provider} 완료 ({duration})`
- [ ] 전체 완료: `[Debate 완료] {rounds}라운드, {total_duration}`

### T10: 조기 합의 감지 (R10)

**소유 파일**: `pkg/orchestra/interactive_debate.go` (확장)
**에이전트**: executor-B

- [ ] 라운드 완료 후 `MergeConsensus()` 66% 임계값으로 합의 체크
- [ ] 합의 시 남은 라운드 skip + 메시지 출력

### T11: 테스트 작성

**소유 파일**: `pkg/orchestra/interactive_debate_test.go`, `round_signal_test.go`
**에이전트**: executor-D (전용 테스트 에이전트)
**의존**: T1, T2, T3

- [ ] `RoundSignalName()` 파일명 생성 테스트 (하위 호환 포함)
- [ ] `CleanRoundSignals()` 정리 동작 테스트
- [ ] `runInteractiveDebate()` mock terminal 기반 통합 테스트
- [ ] `--rounds` 플래그 검증 테스트

---

## 실행 순서 및 병렬성

```
Phase 1:
  executor-A: T1 → T2 → T5     (signal layer)
  executor-B: [wait T1,T2] → T3 → T4   (debate loop)
  executor-C: T6 + T7           (CLI + hooks, 병렬 가능)

Phase 3:
  executor-B: T8 → T9 → T10    (judge, progress, consensus)
  executor-D: T11               (테스트, T1-T3 완료 후)
```

---

## 위험 완화

| 위험 | 완화 |
|------|------|
| interactive_debate.go 200줄 초과 | judge 로직을 별도 함수로 분리, progress 출력은 헬퍼 함수화 |
| 프롬프트 크기 폭발 | buildRebuttalPrompt에 2000자 truncate 적용 |
| pane 세션 불안정 | 각 라운드 시작 전 pollUntilPrompt로 상태 확인, 실패 시 skip |
| hook_signal.go 하위 호환 깨짐 | round=0 기본값으로 기존 파일명 유지 |
