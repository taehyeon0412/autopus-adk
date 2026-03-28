# SPEC-ORCH-013 리서치

## 기존 코드 분석

### R1 관련: Judge Timeout 구조

**`pkg/orchestra/interactive_debate_helpers.go:52-74` — `runJudgeRound`**
- v0.20.1에서 이미 `context.Background()` 기반 fresh context 사용 중
- `judgeTimeout = max(cfg.TimeoutSeconds, 60s)` — 최소 60초 보장
- `runProvider(judgeCtx, judgeCfg, judgment)` — 내부적으로 `cmd.Run()` 반환이 completion signal
- 현재 코드는 이미 올바른 방향이나, `perRoundTimeout` 함수가 judge budget을 고려하지 않음

**`pkg/orchestra/interactive_debate_helpers.go:106-118` — `perRoundTimeout`**
- `totalSeconds / rounds` 단순 분배 — judge 시간을 별도 확보하지 않음
- 최소 45초 floor 적용
- 문제: debate 3라운드에 120초 전체를 분배하면 parent context가 만료됨
  - 다만 judge는 `context.Background()` 사용하므로 실제로는 영향 없음
  - 그래도 debate에 더 정확한 budget을 할당하는 것이 의미론적으로 올바름

**`pkg/orchestra/interactive_debate.go:148-153` — judge 호출부**
- `runJudgeRound(ctx, cfg, panes, hookSession, finalResponses, rounds)` — 첫 인자로 parent ctx 전달
- 하지만 `runJudgeRound`는 첫 인자를 `_`로 무시하고 fresh context 생성
- 시그니처와 실제 사용이 불일치 — `_` 파라미터 정리 권장

### R2 관련: waitForCompletion 구조

**`pkg/orchestra/interactive.go:273-298` — `waitForCompletion`**
- 2-phase consecutive match: prompt 1회 감지 → `candidateDetected = true` → 2회 연속 감지 → 완료
- 문제: 이전 라운드 prompt `> `가 이미 화면에 있으면 첫 poll에서 즉시 candidate, 두 번째 poll에서 confirm
- 해결: baseline screen과 비교하여 content 변화가 없으면 match를 건너뜀

**`pkg/orchestra/interactive.go:230-267` — `waitAndCollectResults`**
- `waitForCompletion(ctx, cfg.Terminal, pi, patterns)` 호출
- baseline 전달 메커니즘 없음 → 시그니처 변경 필요

**`pkg/orchestra/interactive_debate.go:159-208` — `executeRound`**
- 라인 191: `cfg.Terminal.SendLongText(ctx, pi.paneID, prompt)` — 프롬프트 전송
- 라인 207: `waitAndCollectResults(ctx, cfg, panes, patterns, time.Now())` — 결과 수집
- 전송 직전에 ReadScreen snapshot을 수집하여 waitAndCollectResults에 전달 가능

### R3 관련: FormatDebate diff 생성

**`pkg/orchestra/merger.go:109-137` — `FormatDebate`**
- 라인 123-136: `findDifferences(responses)` 호출 — raw responses 직접 전달
- `findDifferences`는 `responses[i].Output`을 `splitLines` + `normalizeLine`으로 비교
- `normalizeLine`은 공백/대소문자만 정규화하고 ANSI/noise는 처리하지 않음
- 해결: `findDifferences` 호출 전에 responses 복사본의 Output에 `cleanScreenOutput` 적용

**`pkg/orchestra/interactive_detect.go:134-138` — `cleanScreenOutput`**
- `SanitizeScreenOutput` → `stripInlineNoise` → `filterPromptLines` 순서
- 이 함수를 diff 생성 전에 적용하면 noise 문제 해결

### R4 관련: cliNoisePatterns 현황

**`pkg/orchestra/interactive_detect.go:29-52` — `cliNoisePatterns`**
- gemini CLI noise: 10개 패턴
- opencode TUI noise: 4개 패턴 (`Build · gpt`, `Build GPT-`, `⬝+ esc`, `ctrl+[a-z]`)
- cmux status bar: 1개 패턴

**누락된 패턴:**
1. Shell login banner: `Last login: ... on ttys...` — macOS 터미널 기본 배너
2. User@host prompt: `bitgapnam@Mac ~ %` — zsh default prompt
3. OpenCode TUI chrome 추가: `gpt-5.4 OpenAI` 라인 등 (기존 `Build GPT-` 패턴과 겹치지 않는 변형)

## 설계 결정

### D1: waitForCompletion에 baseline 추가 vs 별도 함수

**선택: baseline 파라미터 추가**
- 기존 함수 시그니처를 확장하여 baseline string을 받도록 변경
- 대안: 별도 `waitForCompletionWithBaseline` 함수 생성
  - 코드 중복이 발생하고 유지보수 부담 증가
  - 기존 호출자가 baseline 없이 호출하는 경우 빈 문자열 전달로 기존 동작 유지

### D2: FormatDebate에서 원본 수정 vs 복사본 정제

**선택: 복사본 정제**
- `responses` 슬라이스를 shallow copy하여 Output만 cleanScreenOutput 적용
- 원본 responses는 buildDebateResult, OrchestraResult.Responses 등에서 원본 유지 필요
- 대안: findDifferences 내부에서 정제
  - 관심사 분리 위반 — findDifferences는 비교 로직에만 집중해야 함

### D3: cliNoisePatterns 확장 vs 별도 opencode 전용 필터

**선택: 기존 배열 확장**
- `cliNoisePatterns`은 `isPromptLine`에서 일괄 적용되므로 배열에 추가만으로 동작
- 대안: provider별 필터 맵
  - 현재 아키텍처에서 provider 정보가 isPromptLine까지 전달되지 않아 대규모 리팩토링 필요
  - 향후 provider별 정제가 필요하면 그때 분리
