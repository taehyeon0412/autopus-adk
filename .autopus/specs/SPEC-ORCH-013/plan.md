# SPEC-ORCH-013 구현 계획

## 태스크 목록

### R1: Debate/Judge Timeout 분리

- [ ] T1: `runPaneDebate`에서 judge 호출 시 parent context 대신 `context.Background()` 기반 fresh context 사용 확인 — 현재 v0.20.1에서 이미 적용됨, 코드 리뷰로 검증
- [ ] T2: `perRoundTimeout` 산출 시 judge 시간을 별도 budget으로 분리 — totalSeconds에서 judge budget(최소 60초)을 먼저 차감한 나머지를 rounds에 분배. 예: total=120, judgeReserve=60, debateBudget=60, perRound=60/3=20 → floor 45초 적용 → 45초/round
- [ ] T3: `runJudgeRound` 로그에 "event-based completion" 명시 — `cmd.Run()` 반환이 primary signal임을 주석과 로그에 표현

### R2: Claude ReadScreen 타이밍

- [ ] T4: `waitForCompletion`에 `baselineScreen` 파라미터 추가 — 프롬프트 전송 직전 ReadScreen snapshot을 기준선으로 전달
- [ ] T5: 2-phase 로직 앞에 "screen changed" 조건 추가 — `baselineScreen`과 현재 screen이 다를 때만 prompt match 시도
- [ ] T6: `waitAndCollectResults`와 `executeRound`에서 prompt 전송 전 baseline snapshot 수집 후 `waitForCompletion`에 전달. `RunInteractivePaneOrchestra` 내 호출(line 94, 98)도 baseline="" (빈 문자열)로 전달하여 기존 동작 유지

### R3: Diff 섹션 Noise 정제

- [ ] T7: `FormatDebate` 내 `findDifferences` 호출 전에 각 response.Output에 `cleanScreenOutput` 적용
- [ ] T8: `findDifferences` 함수 시그니처는 유지하되, 입력 데이터가 이미 정제됨을 전제하는 주석 추가

### R4: OpenCode 출력 정제

- [ ] T9: `cliNoisePatterns`에 shell login banner 패턴 추가: `(?i)^Last login:` (macOS/Linux 모두 커버)
- [ ] T10: `cliNoisePatterns`에 user@host prompt 패턴 추가: `^\w+@[\w.-]+\s*[%$#]\s*$`
- [ ] T11: opencode TUI chrome 패턴 보강: `(?i)^\s*gpt-[\d.]+\s+OpenAI` 등 누락 패턴 추가
- [ ] T12: 추가된 패턴에 대한 단위 테스트 작성

## 구현 전략

### 접근 방법

1. **R1 (T1-T3)**: v0.20.1에서 이미 `runJudgeRound`가 `context.Background()` 기반 fresh context를 사용하고 있어 핵심 수정은 완료됨. `perRoundTimeout`이 judge 시간까지 포함해 분배하지 않도록 확인하고, 주석/로그를 명확화하는 수준.

2. **R2 (T4-T6)**: `waitForCompletion`의 시그니처에 baseline string을 추가하는 것이 핵심. 기존 2-phase consecutive match 로직은 유지하되, "screen content changed" 게이트를 앞단에 추가. baseline과 현재 screen이 동일하면 prompt match를 건너뜀.

3. **R3 (T7-T8)**: `FormatDebate`에서 `findDifferences` 호출 전 response를 복제하여 Output을 cleanScreenOutput으로 정제. 원본 responses는 변경하지 않음 (다른 용도에서 원본 필요).

4. **R4 (T9-T12)**: 기존 `cliNoisePatterns` 배열에 패턴 추가만으로 해결. 새 패턴은 기존 `isPromptLine` 함수를 통해 자동 적용됨.

### 기존 코드 활용

- `cleanScreenOutput()` — R3에서 diff 정제에 재사용
- `isPromptLine()` / `cliNoisePatterns` — R4에서 확장
- `waitForCompletion()` — R2에서 baseline 로직 추가
- `runJudgeRound()` — R1에서 확인/강화

### 변경 범위

- 5개 파일 수정 (신규 파일 없음)
- 약 50-80 라인 추가/수정
- 기존 테스트 호환성 유지 필수
