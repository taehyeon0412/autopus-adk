# PRD: Interactive 멀티턴 핑퐁 Debate + Hook 자동주입

> Product Requirements Document — Standard (10-section format).

- **SPEC-ID**: SPEC-ORCH-008
- **Author**: Autopus Planner Agent
- **Status**: Draft
- **Date**: 2026-03-27
- **Extends**: SPEC-ORCH-007 (Hook 기반 멀티프로바이더 오케스트레이션)
- **Origin**: BS-003 (Interactive 멀티턴 핑퐁 Debate 브레인스토밍)

---

## 1. Problem & Context

**현재 상황**

SPEC-ORCH-007은 hook 파일 시그널 기반 결과 수집을 완성했다. 프로바이더 CLI가 응답을 완료하면 hook 스크립트가 `{provider}-result.json`과 `{provider}-done` 시그널 파일을 생성하고, 오케스트레이터가 파일 폴링으로 결과를 수집한다. debate 전략(`debate.go`)은 `runRebuttalRound()`을 통해 Phase 1(병렬 응답) -> Phase 2(반박) -> Phase 3(판정)의 3단계 토론을 지원한다.

**문제**

현재 debate의 multi-round 흐름에는 두 가지 구조적 한계가 있다:

1. **Interactive 모드에서 멀티턴 미지원**: `interactive.go`의 `RunInteractivePaneOrchestra()`는 프로바이더에 프롬프트를 한 번 전송하고 결과를 한 번 수집한 후 종료한다. Debate 전략이 `DebateRounds >= 2`일 때 rebuttal을 실행하려면 `debate.go`의 `runRebuttalRound()`이 `runProvider()`로 새 프로세스를 spawn한다. 이는 이미 열려 있는 interactive pane 세션을 활용하지 못하고, pane에서 사용자가 토론 과정을 실시간으로 관찰할 수 없게 한다.

2. **라운드 간 hook 시그널 충돌**: SPEC-ORCH-007의 hook 시그널은 `{provider}-done` 단일 파일이다. 멀티턴에서 Round 1과 Round 2가 동일한 `{provider}-done` 파일을 사용하면, Round 1의 잔여 시그널이 Round 2를 즉시 완료로 오판한다.

3. **핑퐁 토론의 부재**: 현재 rebuttal은 단일 라운드다. 프로바이더 A가 B의 응답을 보고 반박하지만, B가 A의 반박을 보고 재반박하는 멀티턴 핑퐁이 없다. 진정한 토론은 여러 라운드의 반복적 교차 반박이 필요하다.

**영향**

- `--multi --strategy debate` 실행 시 interactive pane에서 토론 과정을 관찰할 수 없다. 사용자는 최종 결과만 본다.
- 단일 rebuttal만으로는 프로바이더 간 심층적 논점 수렴이 불가능하다.
- BS-003에서 핵심 요구로 식별된 "프로바이더 간 멀티턴 핑퐁 토론"이 구현되지 않았다.

**변경 동기**

BS-003 브레인스토밍에서 식별한 핵심 아이디어: interactive pane에서 라운드별로 프롬프트를 주입하고, hook 시그널로 라운드별 완료를 감지하며, 결과를 다음 라운드의 rebuttal 프롬프트에 주입하는 멀티턴 핑퐁 루프. SPEC-ORCH-007이 hook 인프라를 완성했으므로, 이제 그 위에 멀티턴 루프를 구축할 시점이다.

---

## 2. Goals & Success Metrics

| 목표 | 성공 지표 | 목표값 | 일정 |
|------|----------|--------|------|
| Interactive pane에서 멀티턴 debate 실행 | 라운드 루프 정상 완주율 | >= 90% (3라운드 기준) | Phase 1 |
| 라운드별 hook 시그널 분리 | 라운드 간 시그널 충돌 발생률 | 0% | Phase 1 |
| 핑퐁 rebuttal 프롬프트 주입 | 이전 라운드 응답이 다음 라운드 프롬프트에 포함 확인 | 100% | Phase 1 |
| --rounds N CLI 플래그 | 사용자가 라운드 수 지정 가능 | 1-10 범위 지원 | Phase 1 |
| Judge 라운드 interactive 실행 | 최종 판정이 interactive pane에서 실행 | 정상 동작 | Phase 2 |
| 사용자 실시간 토론 관찰 | cmux pane에서 각 라운드 진행 가시성 | 라운드별 프롬프트/응답 확인 가능 | Phase 2 |

**Anti-Goals** (성공이 아닌 것)

- 비대화식(non-interactive, sentinel) 모드의 debate 흐름을 변경하는 것은 아니다. 기존 `debate.go`의 `runProvider()` 기반 흐름은 그대로 유지한다.
- Hook 스크립트 자체의 구조를 재설계하는 것은 아니다. SPEC-ORCH-007의 hook 스크립트를 라운드 인식으로 확장만 한다.
- 실시간 스트리밍 결과 표시(토큰 단위)를 구현하는 것은 아니다. 결과는 라운드 완료 단위로 수집한다.

---

## 3. Target Users

| 사용자 그룹 | 역할 | 사용 빈도 | 핵심 기대 |
|------------|------|----------|----------|
| 로컬 개발자 | `auto orchestra --multi --strategy debate --rounds N` 사용자 | 일일 | 여러 AI 프로바이더가 토론하며 논점을 수렴하는 과정을 실시간으로 관찰 |
| 의사결정자 | 복잡한 설계 결정에 멀티프로바이더 토론 활용 | 주간 | N라운드 토론 후 최종 판정으로 신뢰도 높은 결론 도출 |
| Autopus-ADK 기여자 | debate 전략 확장 개발자 | 월간 | 라운드 루프 구조를 이해하고 새 토론 전략을 추가할 수 있음 |

**Primary User**: 로컬 개발자 (cmux pane에서 AI 프로바이더 간 멀티턴 토론을 실시간 관찰하며 결과 활용)

---

## 4. User Stories

### Story 1: 멀티턴 핑퐁 Debate 실행

**As a** 로컬 개발자,
**I want** `auto orchestra --multi --strategy debate --rounds 3`으로 3라운드 토론을 실행하도록,
**so that** 프로바이더들이 서로의 응답을 보고 반박하는 과정을 통해 더 정제된 결론을 얻을 수 있다.

**Acceptance Criteria**

- Given 3개 프로바이더(claude, gemini, opencode)가 설정된 상태에서, when `--rounds 3`으로 debate 실행 시, then Round 1(독립 응답) -> Round 2(교차 반박) -> Round 3(재반박) 순서로 3라운드가 실행된다.
- Given Round 1이 완료되면, when Round 2 시작 시, then 각 프로바이더의 pane에 다른 프로바이더들의 Round 1 응답이 포함된 rebuttal 프롬프트가 전송된다.
- Given 모든 라운드가 완료되면, when 결과 병합 시, then 최종 라운드의 응답을 기반으로 debate 결과가 생성된다.

---

### Story 2: 라운드별 Hook 시그널 분리

**As a** 시스템 (오케스트레이터),
**I want** 각 라운드마다 독립된 hook 시그널 파일을 사용하도록,
**so that** 이전 라운드의 시그널이 다음 라운드의 완료 감지를 방해하지 않는다.

**Acceptance Criteria**

- Given Round 1이 진행 중일 때, when claude가 응답을 완료하면, then `claude-round1-done` 시그널이 생성된다 (`claude-done`이 아닌).
- Given Round 2 시작 전에, when 오케스트레이터가 이전 라운드 시그널을 정리하면, then `claude-round1-done`과 `claude-round1-result.json`이 삭제 또는 무시된다.
- Given 멀티턴이 아닌 단일 실행(rounds=1)일 때, when 기존 hook 스크립트가 실행되면, then 기존 `{provider}-done` 형식과 하위 호환된다.

---

### Story 3: Interactive Pane에서 실시간 토론 관찰

**As a** 로컬 개발자,
**I want** cmux pane에서 각 프로바이더의 라운드별 응답이 실시간으로 표시되도록,
**so that** 토론 진행 과정을 관찰하고 필요시 개입할 수 있다.

**Acceptance Criteria**

- Given debate가 interactive mode로 실행 중일 때, when Round 2 rebuttal 프롬프트가 전송되면, then 각 pane에서 "## Other debaters' arguments:" 섹션이 포함된 프롬프트가 표시된다.
- Given 3라운드 토론이 진행 중일 때, when 사용자가 pane을 관찰하면, then 각 라운드의 시작과 종료가 시각적으로 구분된다.

---

### Story 4: Judge 라운드

**As a** 로컬 개발자,
**I want** 모든 debate 라운드 완료 후 judge 프로바이더가 최종 판정을 내리도록,
**so that** 토론 결과를 종합한 공정한 결론을 얻을 수 있다.

**Acceptance Criteria**

- Given `--judge claude`로 judge가 지정된 상태에서, when 모든 debate 라운드가 완료되면, then judge pane에 전체 토론 기록이 포함된 판정 프롬프트가 전송된다.
- Given judge가 판정을 완료하면, when 결과 병합 시, then judge 응답이 `(judge)` 태그와 함께 최종 결과에 포함된다.

---

### Story 5: --rounds CLI 플래그

**As a** 로컬 개발자,
**I want** `--rounds N` 플래그로 토론 라운드 수를 지정하도록,
**so that** 주제 복잡도에 따라 토론 깊이를 조절할 수 있다.

**Acceptance Criteria**

- Given `--strategy debate` 없이 `--rounds 3`을 지정하면, when 실행 시, then 오류 메시지와 함께 `--rounds requires --strategy debate`을 출력한다.
- Given `--rounds 0` 또는 `--rounds 11`을 지정하면, when 검증 시, then 유효 범위(1-10) 오류를 출력한다.
- Given `--rounds`를 지정하지 않으면, when debate 전략 실행 시, then 기본값 2(1라운드 독립 응답 + 1라운드 rebuttal)로 동작한다.

---

## 5. Functional Requirements

### P0 -- Must Have

#### R1: Interactive Debate 라운드 루프

WHEN debate 전략이 interactive 모드에서 실행될 때, THE SYSTEM SHALL `RunInteractivePaneOrchestra()` 내에서 `DebateRounds` 횟수만큼 라운드 루프를 실행한다:
- Round 1: 원본 프롬프트를 각 pane에 전송하고 hook 결과를 수집한다.
- Round 2..N: 이전 라운드 결과를 기반으로 `buildRebuttalPrompt()`를 호출하여 rebuttal 프롬프트를 생성하고, 각 pane에 전송하고 hook 결과를 수집한다.
- 각 라운드 사이에 이전 라운드의 hook 시그널 파일을 정리한다.

#### R2: 라운드 스코프 Hook 시그널

WHEN 멀티턴 debate(rounds >= 2)가 실행될 때, THE SYSTEM SHALL hook 시그널 파일에 라운드 번호를 포함한다:
- 결과 파일: `/tmp/autopus/{session-id}/{provider}-round{N}-result.json`
- 완료 시그널: `/tmp/autopus/{session-id}/{provider}-round{N}-done`
- 환경변수 `AUTOPUS_ROUND`를 현재 라운드 번호로 설정하여 hook 스크립트에 전달한다.
- `AUTOPUS_ROUND`가 미설정이면 기존 `{provider}-done` 형식으로 fallback (하위 호환).

#### R3: Rebuttal 프롬프트 Interactive 주입

WHEN Round N(N >= 2)이 시작될 때, THE SYSTEM SHALL 각 프로바이더 pane에 rebuttal 프롬프트를 주입한다:
- `buildRebuttalPrompt(original, otherResponses)`를 호출하여 프롬프트를 생성한다.
- `sendPrompts()`와 동일한 방식으로 pane에 프롬프트 텍스트를 전송하고 Enter를 전송한다.
- 프롬프트 전송 전 pane이 입력 대기 상태(프롬프트 패턴 감지)임을 확인한다.

#### R4: --rounds N CLI 플래그

WHEN 사용자가 `--rounds N`을 지정하면, THE SYSTEM SHALL:
- `OrchestraConfig.DebateRounds`에 N을 설정한다.
- N의 유효 범위는 1-10이다. 범위 밖이면 오류를 반환한다.
- `--strategy debate`가 아닌 전략에서 `--rounds`를 지정하면 오류를 반환한다.
- 미지정 시 debate 전략의 기본 라운드 수는 2이다.

#### R5: Hook 스크립트 라운드 인식

WHEN `AUTOPUS_ROUND` 환경변수가 설정되어 있으면, THE SYSTEM SHALL 각 hook 스크립트가 시그널 파일명에 라운드 번호를 포함한다:
- Claude Stop hook: `{provider}-round{AUTOPUS_ROUND}-result.json`, `{provider}-round{AUTOPUS_ROUND}-done`
- Gemini AfterAgent hook: 동일 패턴
- opencode plugin: 동일 패턴
- `AUTOPUS_ROUND` 미설정 시 기존 `{provider}-result.json`, `{provider}-done` 형식 유지.

#### R6: 라운드 간 시그널 정리

WHEN 라운드 N이 완료되고 라운드 N+1이 시작되기 전에, THE SYSTEM SHALL:
- 라운드 N의 `{provider}-round{N}-done` 시그널 파일을 삭제한다.
- 결과 파일(`{provider}-round{N}-result.json`)은 보존한다 (최종 결과 병합에 필요).
- `AUTOPUS_ROUND` 환경변수를 N+1로 업데이트한다.

#### R7: Interactive Debate 결과 병합

WHEN 모든 debate 라운드가 완료되면, THE SYSTEM SHALL:
- 최종 라운드의 응답을 `mergeByStrategy(StrategyDebate, ...)`에 전달한다.
- 전체 라운드 히스토리(각 라운드의 모든 프로바이더 응답)를 `OrchestraResult`에 포함한다.
- `buildDebateMerged()`가 라운드 수와 함께 요약을 생성한다.

### P1 -- Should Have

#### R8: Judge 라운드 Interactive 실행

WHEN `JudgeProvider`가 설정되고 모든 debate 라운드가 완료되면, THE SYSTEM SHALL:
- Judge 프로바이더의 pane에 `buildJudgmentPrompt()`로 생성된 판정 프롬프트를 전송한다.
- Judge 응답을 hook 시그널로 수집한다 (라운드 번호 = "judge").
- Judge가 참여 프로바이더 중 하나인 경우, 해당 pane을 재사용한다. 별도 프로바이더인 경우, 추가 pane을 생성한다.

#### R9: 라운드 진행 상태 표시

WHILE debate 라운드가 진행 중일 때, THE SYSTEM SHALL:
- 각 라운드 시작 시 `[Round N/M] 시작...` 메시지를 stdout에 출력한다.
- 각 프로바이더의 라운드 완료 시 `[Round N/M] {provider} 완료 ({duration})` 메시지를 출력한다.
- 전체 debate 완료 시 `[Debate 완료] {rounds}라운드, {total_duration}` 요약을 출력한다.

#### R10: 조기 합의 감지

WHEN 라운드 N 완료 후 모든 프로바이더의 응답이 실질적으로 동일하면, THE SYSTEM SHALL:
- 남은 라운드를 건너뛰고 조기 종료한다.
- `MergeConsensus()` 66% 임계값을 활용하여 합의 여부를 판단한다.
- 조기 종료 시 `[Early Consensus] Round {N}에서 합의 도달, 남은 {M}라운드 건너뜀` 메시지를 출력한다.

#### R11: OrchestraConfig 라운드 히스토리 구조

WHEN debate 전략이 interactive 멀티턴으로 실행될 때, THE SYSTEM SHALL:
- `OrchestraResult`에 `RoundHistory [][]ProviderResponse` 필드를 추가한다.
- 각 라운드의 모든 프로바이더 응답을 라운드 순서대로 저장한다.
- 기존 `Responses` 필드는 최종 라운드의 응답으로 설정한다 (하위 호환).

### P2 -- Could Have

#### R12: 라운드별 결과 비교 뷰

WHEN 모든 라운드가 완료되면, THE SYSTEM SHALL 라운드별 응답 변화를 요약하는 비교 뷰를 제공한다.

#### R13: 라운드별 개별 타임아웃

WHERE `PerRoundTimeout`이 설정되면, THE SYSTEM SHALL 전체 타임아웃과 별개로 각 라운드에 개별 타임아웃을 적용한다. 라운드 타임아웃 초과 시 해당 라운드를 부분 결과로 완료 처리한다.

#### R14: 토론 기록 파일 저장

WHEN debate가 완료되면, THE SYSTEM SHALL 전체 토론 기록(라운드별 프롬프트 + 응답)을 `.autopus/debate-history/{session-id}.json`에 저장한다.

---

## 6. Non-Functional Requirements

| 카테고리 | 요구사항 | 목표 |
|---------|---------|------|
| 성능 | 라운드 간 전환 지연 (시그널 정리 + 프롬프트 주입) | < 3초 |
| 성능 | 라운드별 hook 시그널 감지 지연 | done 파일 생성 후 500ms 이내 |
| 성능 | 3라운드 * 3프로바이더의 총 오버헤드 (AI 응답 시간 제외) | < 30초 |
| 안정성 | N라운드 루프에서 특정 라운드 실패 시 | 실패 라운드까지의 결과로 부분 완료, 전체 crash 방지 |
| 안정성 | 특정 프로바이더 1개가 타임아웃 시 | 다른 프로바이더 결과는 정상 수집, 해당 프로바이더만 skip |
| 보안 | 라운드별 시그널 파일 권한 | 기존 0o600 유지 |
| 보안 | 환경변수 AUTOPUS_ROUND 주입 | pane 내부에서만 유효, 외부 프로세스에 영향 없음 |
| 호환성 | rounds=1 또는 미지정 시 | SPEC-ORCH-007의 단일 수집 흐름과 동일하게 동작 |
| 호환성 | 비대화식 debate 전략 (runDebate) | 기존 runProvider() 기반 흐름 변경 없음 |
| 파일 크기 | 신규/수정 소스 파일 | 300줄 하드 리밋, 200줄 목표 |

---

## 7. Technical Constraints

**기술 스택 제약**

- Go 1.26+, 외부 의존성 최소화 (stdlib 우선)
- 파일 감시: `os.Stat()` 폴링 200ms 간격 (기존 hook_signal.go 패턴 재사용)
- Hook 스크립트: POSIX shell 호환, `AUTOPUS_ROUND` 환경변수로 라운드 인식
- 파일 크기 제한: 소스 파일 300줄 하드 리밋, 200줄 목표

**기존 코드 활용**

| 기존 코드 | 활용 방식 |
|----------|----------|
| `interactive.go` RunInteractivePaneOrchestra() | 라운드 루프로 확장 (신규 파일에 루프 로직 분리) |
| `debate.go` buildRebuttalPrompt() | interactive 라운드 루프에서 직접 호출 |
| `debate.go` buildJudgmentPrompt() | interactive judge 라운드에서 직접 호출 |
| `hook_signal.go` HookSession | WaitForDone에 라운드 파라미터 추가 |
| `hook_watcher.go` WaitAndCollectHookResults() | 라운드별 호출로 확장 |
| `types.go` OrchestraConfig.DebateRounds | 기존 필드 재사용, CLI 연결만 추가 |

**파일 구조 계획**

| 파일 | 역할 | 줄 수 목표 |
|------|------|-----------|
| `pkg/orchestra/interactive_debate.go` (신규) | Interactive debate 라운드 루프 메인 로직 | < 200 |
| `pkg/orchestra/round_signal.go` (신규) | 라운드 스코프 시그널 관리 (정리, 라운드별 파일명) | < 100 |
| `pkg/orchestra/interactive.go` (수정) | debate 전략 분기 추가 (interactive_debate.go 호출) | 변경 < 20줄 |
| `pkg/orchestra/hook_signal.go` (수정) | WaitForDone에 round 파라미터 추가 | 변경 < 30줄 |
| `pkg/orchestra/types.go` (수정) | OrchestraResult에 RoundHistory 추가 | 변경 < 10줄 |
| `content/hooks/hook-claude-stop.sh` (수정) | AUTOPUS_ROUND 인식 파일명 분기 | 변경 < 10줄 |
| `content/hooks/hook-gemini-afteragent.sh` (수정) | AUTOPUS_ROUND 인식 파일명 분기 | 변경 < 10줄 |
| `content/hooks/hook-opencode-complete.ts` (수정) | AUTOPUS_ROUND 인식 파일명 분기 | 변경 < 10줄 |
| `internal/cli/` (수정) | --rounds 플래그 추가 | 변경 < 20줄 |

**호환성 요구사항**

- macOS (darwin) 및 Linux 지원
- cmux 및 tmux 터미널 모두 지원
- rounds=1인 경우 기존 SPEC-ORCH-007 흐름과 동일하게 동작
- 비대화식 debate 전략 (`runDebate()`)은 변경 없음

---

## 8. Out of Scope

이 릴리즈에서 다루지 않는 항목:

- **비대화식(sentinel/process) debate 멀티턴 변경**: `debate.go`의 `runDebate()` / `runRebuttalRound()` 기존 흐름은 유지
- **사용자 개입 기능**: 토론 중간에 사용자가 프롬프트를 수정하거나 추가 지시를 내리는 기능
- **실시간 토큰 스트리밍**: 라운드 완료 단위 수집만 지원, 토큰 단위 스트리밍은 미지원
- **4개 이상 프로바이더 동시 debate**: 현재 3프로바이더 아키텍처 유지
- **자동 라운드 수 결정**: AI가 토론 수렴도를 판단하여 자동으로 라운드를 조절하는 기능
- **Windows 지원**: POSIX shell hook 스크립트 기반

**향후 반복으로 연기**

- 사용자 실시간 개입 (moderator 모드)
- 토론 수렴도 기반 자동 라운드 조절
- 토론 결과의 구조화된 리포트 생성

---

## 9. Risks & Open Questions

### Risks

| 위험 | 심각도 | 확률 | 완화 전략 |
|------|--------|------|----------|
| 라운드 간 pane 세션 상태 불안정 | High | Medium | 각 라운드 시작 전 pane 입력 대기 상태를 확인 (`pollUntilPrompt`). 실패 시 해당 프로바이더를 skip하고 나머지로 계속 |
| 멀티턴에서 프롬프트 크기 폭발 | Medium | High | 라운드가 누적될수록 rebuttal 프롬프트에 이전 응답이 포함되어 크기가 급증한다. 각 프로바이더 응답을 최대 2000자로 truncate하여 프롬프트 크기를 제한 |
| Hook 스크립트가 AUTOPUS_ROUND를 인식하지 못하는 환경 | Medium | Low | `AUTOPUS_ROUND` 미설정 시 기존 `{provider}-done` 형식으로 fallback. 오케스트레이터가 양쪽 파일명을 모두 확인 |
| 3라운드 * 3프로바이더의 총 실행 시간 과다 | Medium | Medium | 기본 라운드당 타임아웃을 전체 타임아웃/라운드수로 분배. 조기 합의 감지(R10)로 불필요한 라운드 skip |
| 프로바이더 CLI가 동일 세션에서 연속 프롬프트를 거부 | Low | Low | 대부분의 AI CLI는 대화 모드를 지원한다. 거부 시 해당 프로바이더를 fallback으로 전환 |

### Open Questions

| # | 질문 | 담당 | 기한 | 상태 |
|---|------|------|------|------|
| Q1 | Rebuttal 프롬프트에 이전 라운드 응답을 몇 글자까지 포함해야 하는가? 전체 포함 vs truncate? | planner | 2026-04-10 | Open — 기본 2000자 truncate 제안 |
| Q2 | 조기 합의 감지의 "실질적 동일" 판정 기준은? 텍스트 유사도 vs 키워드 매칭? | planner | 2026-04-10 | Open — 기존 MergeConsensus 66% 임계값 재사용 제안 |
| Q3 | Judge 프로바이더가 참여 프로바이더와 동일한 경우 pane 재사용이 안전한가? 같은 pane에서 debate 후 judge 프롬프트를 보내면 컨텍스트 오염 문제는? | executor | 2026-04-10 | Open |
| Q4 | `AUTOPUS_ROUND` 환경변수를 pane에서 export할 때, 프로바이더 CLI 실행 전에 설정해야 하는가 아니면 런타임에 동적 변경이 가능한가? | executor | 2026-04-05 | Open — 라운드마다 `export AUTOPUS_ROUND=N` 전송 제안 |
| Q5 | debate 기본 라운드 수를 1(현재)에서 2로 변경하면 기존 사용자에게 breaking change인가? | planner | 2026-04-05 | Open — debate 전략 선택 자체가 멀티턴 의도이므로 기본 2가 적절 |

---

## 10. Practitioner Q&A

**Q1: Interactive debate와 기존 debate의 차이점은?**
A: 기존 `debate.go`의 `runDebate()`는 `runProvider()`로 새 프로세스를 spawn하여 각 라운드를 실행한다. Interactive debate는 이미 열린 pane 세션에 프롬프트를 재주입하여 동일 세션 내에서 멀티턴을 진행한다. 사용자는 cmux pane에서 토론 과정을 실시간으로 관찰할 수 있다.

**Q2: 라운드 루프의 구체적 실행 흐름은?**
A:
1. Round 1: `sendPrompts(cfg.Prompt)` -> `WaitAndCollectHookResults(round=1)` -> `responses[1]` 저장
2. Round 2: `cleanRoundSignals(round=1)` -> `setRoundEnv(round=2)` -> 각 프로바이더별 `buildRebuttalPrompt(cfg.Prompt, otherResponses)` -> `sendPrompts(rebuttalPrompt)` -> `WaitAndCollectHookResults(round=2)` -> `responses[2]` 저장
3. Round N: 반복
4. (Optional) Judge: `buildJudgmentPrompt(topic, allRoundResponses)` -> judge pane에 전송 -> 수집

**Q3: `interactive_debate.go`의 메인 함수 시그니처는?**
A: `func runInteractiveDebate(ctx context.Context, cfg OrchestraConfig, panes []paneInfo, hookSession *HookSession) ([]ProviderResponse, [][]ProviderResponse, error)` — 최종 응답과 라운드 히스토리를 반환한다.

**Q4: hook 스크립트의 라운드 인식은 어떻게 구현되는가?**
A: 각 hook 스크립트가 `$AUTOPUS_ROUND` 환경변수를 확인한다. 설정되어 있으면 `{provider}-round{AUTOPUS_ROUND}-result.json`과 `{provider}-round{AUTOPUS_ROUND}-done`을 생성한다. 미설정이면 기존 `{provider}-result.json`과 `{provider}-done`을 생성한다. 예시 (shell):
```bash
if [ -n "$AUTOPUS_ROUND" ]; then
  RESULT_FILE="${PROVIDER}-round${AUTOPUS_ROUND}-result.json"
  DONE_FILE="${PROVIDER}-round${AUTOPUS_ROUND}-done"
else
  RESULT_FILE="${PROVIDER}-result.json"
  DONE_FILE="${PROVIDER}-done"
fi
```

**Q5: 라운드 타임아웃은 어떻게 관리되는가?**
A: 전체 타임아웃(`OrchestraConfig.TimeoutSeconds`)을 라운드 수로 나누어 라운드별 타임아웃을 산출한다. 예: 전체 120초, 3라운드면 라운드당 40초. `context.WithTimeout()`으로 라운드 단위 컨텍스트를 생성한다. P2 R13이 구현되면 `PerRoundTimeout` 필드로 사용자가 직접 지정할 수 있다.

**Q6: 프롬프트 크기 폭발 문제의 구체적 해결 방안은?**
A: `buildRebuttalPrompt()`에 max length 파라미터를 추가한다. 각 프로바이더 응답의 `Output`을 최대 2000자로 truncate한 후 프롬프트에 포함한다. 3프로바이더 * 2000자 = 6000자 + 원본 프롬프트, 총 약 8000자 이내로 유지한다.

**Q7: `interactive.go`에서 debate 분기는 어떻게 추가되는가?**
A: `RunInteractivePaneOrchestra()`의 Step 5 (sendPrompts) ~ Step 6-7 (waitAndCollect) 사이에 전략 체크를 추가한다:
```go
if cfg.Strategy == StrategyDebate && cfg.DebateRounds >= 2 {
    return runInteractiveDebate(ctx, cfg, panes, hookSession)
}
```
이로써 단일 실행(rounds=1)은 기존 흐름을 타고, 멀티턴은 `interactive_debate.go`의 루프를 탄다.

**Q8: 비대화식 모드에서 `--rounds` 플래그의 동작은?**
A: `--rounds`는 `OrchestraConfig.DebateRounds`에 값을 설정한다. 비대화식 모드에서는 기존 `debate.go`의 `runDebate()`가 `DebateRounds`를 참조하여 rebuttal 라운드를 실행한다. 즉 `--rounds 3`이면 Round 1 + 2회 rebuttal = 3라운드가 실행된다. 단, 비대화식에서는 `runProvider()` 기반이므로 각 라운드마다 새 프로세스가 spawn된다.

---

## Quality Validation Checklist

| # | 검증 항목 | 상태 |
|---|----------|------|
| 1 | 10개 섹션 모두 작성됨 | PASS |
| 2 | EARS 형식 요구사항 (WHEN/WHILE/WHERE) | PASS |
| 3 | MoSCoW 우선순위 (P0/P1/P2) 분류됨 | PASS |
| 4 | User Stories에 Given-When-Then 수락 기준 포함 | PASS |
| 5 | 선행 SPEC(SPEC-ORCH-007)과의 관계 명시 | PASS |
| 6 | Anti-Goals 정의됨 | PASS |
| 7 | Risks에 심각도/확률/완화 전략 포함 | PASS |
| 8 | Open Questions에 담당자/기한 포함 | PASS |
| 9 | Technical Constraints에 파일 크기 제한 명시 | PASS |
| 10 | Out of Scope에 미지원 항목과 연기 항목 구분 | PASS |
| 11 | 기존 코드 활용 방안 구체적 명시 | PASS |
| 12 | 파일 구조 계획에 줄 수 목표 포함 | PASS |
| 13 | 하위 호환성 보장 방안 명시 (rounds=1 fallback) | PASS |
| 14 | Practitioner Q&A가 구현 세부사항을 다룸 | PASS |
