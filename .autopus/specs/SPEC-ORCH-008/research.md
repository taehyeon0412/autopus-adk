# SPEC-ORCH-008 리서치

---

## 기존 코드 분석

### interactive.go — 현재 구조와 확장 포인트

`RunInteractivePaneOrchestra()` (L18-109)는 8단계로 구성:
1. R8 fallback 체크 (plain terminal)
2. HookSession 생성 (hook 모드 시)
3. `splitProviderPanes()` — pane 분할
4. `startPipeCapture()` — pipe-pane 시작
5. `launchInteractiveSessions()` — CLI 바이너리 실행
6. `waitForSessionReady()` — 프롬프트 패턴 대기
7. `sendPrompts()` — 프롬프트 전송
8. 결과 수집 (hook 또는 ReadScreen) + merge

**확장 포인트**: Step 7 (sendPrompts) 직전에 debate 전략 분기를 삽입한다.
debate 모드에서는 Step 7-8을 `runInteractiveDebate()`로 대체하고, 라운드 루프 내에서 프롬프트 전송과 결과 수집을 반복한다.

현재 295줄. debate 분기 추가로 약 15줄 증가 → 여전히 300줄 제한 내.

### debate.go — 재사용 가능 함수

| 함수 | 위치 | 재사용 방식 |
|------|------|------------|
| `buildRebuttalPrompt(original, otherResponses)` | L95-104 | Round 2+ 프롬프트 생성에 직접 호출 |
| `buildJudgmentPrompt(topic, arguments)` | L108-116 | Judge 라운드에 직접 호출 |
| `buildDebateMerged(responses, cfg)` | L120-154 | 최종 결과 merge에 재사용 |
| `runRebuttalRound()` | L52-91 | interactive에서는 미사용 (runProvider 기반이므로) |
| `findOrBuildJudgeConfig()` | L158-168 | Judge pane 생성 시 재사용 |

### hook_signal.go — 라운드 확장 방안

`WaitForDone()` (L52-74): 현재 provider-specific done 파일만 지원 (`{provider}-done`).
라운드 확장 방안: variadic round 파라미터 추가.

```go
// Current: WaitForDone(timeout, providers...)
// New: WaitForDone(timeout, providers...) → round 0 (기존 동작)
// New: WaitForRoundDone(timeout, round, providers...) → 라운드 스코프
```

하위 호환을 위해 새 메서드 `WaitForRoundDone()`을 추가하는 방식이 안전하다.
기존 `WaitForDone()` 호출부(hook_watcher.go L70)를 변경할 필요 없음.

`ReadResult()` (L77-94): 동일 패턴으로 `ReadRoundResult(round, providers...)` 추가.

### hook_watcher.go — 라운드별 호출 패턴

`WaitAndCollectHookResults()` (L13-44): 현재 전체 프로바이더를 한 번에 수집.
Interactive debate에서는 이 함수를 직접 사용하지 않고, `interactive_debate.go`에서 라운드별로 `HookSession.WaitForRoundDone()` + `ReadRoundResult()`를 직접 호출한다.

### pane_runner.go — 재사용 가능 인프라

| 함수 | 위치 | 재사용 |
|------|------|--------|
| `splitProviderPanes()` | L89-109 | interactive_debate.go는 panes를 인자로 받음 (이미 분할됨) |
| `paneArgs()` | L178-183 | buildInteractiveLaunchCmd에서 사용 |
| `cleanupPanes()` | L277-283 | interactive.go의 defer에서 호출 (debate 루프 종료 후) |

### types.go — 변경 사항

`OrchestraConfig.DebateRounds` (L74): 이미 존재. 현재 debate.go에서만 사용.
`OrchestraResult`: `RoundHistory [][]ProviderResponse` 필드 추가 필요.

### CLI (orchestra.go) — --rounds 플래그 부재

현재 `--judge` 플래그는 있지만 `--rounds`는 없음. `DebateRounds`가 CLI에서 설정되지 않아 항상 기본값 0 (→ debate.go에서 1로 처리). `--rounds` 플래그를 추가하고 `buildOrchestraConfig()`에서 매핑 필요.

### Hook 스크립트 — 라운드 인식 패턴

3개 hook 스크립트 모두 하드코딩된 파일명 사용:
- `hook-claude-stop.sh` L34: `"${SESSION_DIR}/claude-result.json"`, L39: `"${SESSION_DIR}/claude-done"`
- `hook-gemini-afteragent.sh` L34: `"${SESSION_DIR}/gemini-result.json"`, L39: `"${SESSION_DIR}/gemini-done"`
- `hook-opencode-complete.ts` L30-31: `"opencode-result.json"`, `"opencode-done"`

모든 스크립트에 `AUTOPUS_ROUND` 환경변수 체크 + 파일명 분기 추가 필요.

---

## 설계 결정

### D1: 신규 파일 `interactive_debate.go` vs interactive.go 확장

**결정**: 신규 파일 `interactive_debate.go`에 라운드 루프 로직 분리.

**이유**:
- `interactive.go`는 현재 295줄. 라운드 루프(~150줄)를 추가하면 300줄 하드 리밋 초과.
- 단일 실행과 멀티턴 로직의 관심사가 다름 (단일: send → collect, 멀티턴: loop → rebuttal → collect per round).
- interactive.go에는 분기 코드(~15줄)만 추가.

**대안**:
- interactive.go에 inline으로 추가 → 300줄 초과로 불가.
- debate.go에 추가 → debate.go는 비대화식 전용, interactive pane 의존성 없음.

### D2: WaitForRoundDone() 새 메서드 vs WaitForDone() 시그니처 변경

**결정**: `WaitForRoundDone(timeout, round, providers...)` 신규 메서드 추가.

**이유**:
- `WaitForDone()` 시그니처를 변경하면 hook_watcher.go(L70)의 기존 호출부 수정 필요.
- 새 메서드는 하위 호환을 완전히 유지하며, 기존 단일 실행 경로에 영향 없음.
- `round_signal.go`의 `RoundSignalName()`을 내부에서 호출하여 파일명 결정.

**대안**:
- WaitForDone에 variadic opts 추가 → API가 복잡해지고, 기존 호출부에서 혼동 가능.
- Option struct 패턴 → 과도한 엔지니어링. 라운드 1개 파라미터만 추가하면 됨.

### D3: AUTOPUS_ROUND 전달 방식 — pane export vs 프로세스 환경변수

**결정**: 라운드마다 pane에 `export AUTOPUS_ROUND=N` SendCommand 전송.

**이유**:
- interactive 모드에서 프로바이더 CLI는 이미 실행 중인 프로세스. `os.Setenv()`는 오케스트레이터 프로세스에만 적용되고, pane 내 프로세스에는 전달되지 않음.
- pane에 `export` 명령을 전송하면 해당 pane의 shell 환경에 변수가 설정됨.
- hook 스크립트는 프로바이더 CLI가 완료 시 실행되므로, pane shell의 환경변수를 상속함.

**대안**:
- 오케스트레이터에서 os.Setenv() → pane 프로세스에 전달 안 됨. 불가.
- 시그널 파일명을 오케스트레이터가 직접 계산 → hook 스크립트가 기존 형식으로 파일을 생성하므로 불일치. 양쪽 다 인식 필요.

### D4: 프롬프트 크기 제한

**결정**: `buildRebuttalPrompt()`에 응답당 최대 2000자 truncate 적용.

**이유**:
- 3프로바이더 * 5라운드면 누적 프롬프트가 수만 자로 폭발 가능.
- 대부분의 AI CLI는 프롬프트 크기 제한이 있음 (Claude: ~200K tokens, Gemini: ~1M tokens, 하지만 실용적으로 10K자 이내가 적절).
- 2000자면 핵심 논점은 충분히 전달됨.

**대안**:
- 무제한 → 프롬프트 폭발로 CLI 거부 또는 품질 저하.
- 500자 → 너무 짧아 맥락 손실.
- 요약 모델 호출로 압축 → 추가 API 비용 + 지연. 향후 개선으로 연기.

### D5: debate 기본 라운드 수 변경 (1 → 2)

**결정**: `--strategy debate` 시 기본 라운드 수를 1에서 2로 변경.

**이유**:
- debate 전략을 선택한다는 것 자체가 멀티턴 토론 의도.
- rounds=1이면 단순 병렬 실행과 차이 없음 (rebuttal이 없으므로).
- PRD Q5에서 "debate 전략 선택 자체가 멀티턴 의도이므로 기본 2가 적절"로 합의.
- 기존 사용자에게 breaking change이나, debate 전략 사용자 수가 적고 (v0.12+에서 도입), 기본 동작이 더 유용해짐.

---

## 위험 분석

### 높은 위험: Pane 세션 상태 불안정

멀티라운드에서 pane이 "입력 대기" 상태가 아닌 경우 프롬프트가 무시되거나 중복 전송될 수 있다.

**완화**: 각 라운드 시작 전 `pollUntilPrompt()`로 프롬프트 패턴 감지. 30초 대기 후 실패 시 해당 프로바이더 skip.

### 중간 위험: Hook 스크립트 업데이트 누락

사용자가 기존 hook 스크립트를 수동 수정했을 경우, `auto hooks install`로 업데이트하지 않으면 라운드 인식이 작동하지 않는다.

**완화**: `AUTOPUS_ROUND` 미설정 시 기존 형식 fallback. 오케스트레이터가 양쪽 파일명을 모두 폴링.

### 낮은 위험: interactive.go 파일 크기 초과

현재 295줄. debate 분기 15줄 추가 → 310줄 → 300줄 하드 리밋 초과.

**완화**: `cleanScreenOutput()` 등 유틸 함수를 `interactive_util.go`로 분리하여 250줄 이하 유지.
