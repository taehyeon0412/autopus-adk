# PRD: Orchestra 멀티턴 토론 프로토콜 개선

> Product Requirements Document — Minimal (5-section format).

- **SPEC-ID**: SPEC-ORCH-009
- **Author**: Autopus Planner Agent
- **Status**: Draft
- **Date**: 2026-03-27
- **Extends**: SPEC-ORCH-008 (Interactive 멀티턴 핑퐁 Debate + Hook 자동주입)
- **Origin**: BS-004 (Orchestra 멀티턴 토론 프로토콜 개선 브레인스토밍)

---

## 1. Problem

Brainstorm 멀티턴 debate는 인프라가 존재하나 연결되지 않아 작동하지 않는다.

**근본 원인 3가지:**

1. **rounds=0 하드코딩**: `orchestra_brainstorm.go`(line 31)에서 `runOrchestraCommand()`에 rounds=0을 직접 전달한다. `resolveRounds()`는 `orchestra_helpers.go`(line 71)에 구현되어 있고, debate 전략일 때 기본값 2를 반환하지만, brainstorm 커맨드에서는 호출하지 않는다. `plan` 커맨드는 `resolveRounds()`를 올바르게 사용하고 있어(orchestra.go line 90), brainstorm만 누락된 상태다.

2. **Debate 라우팅 미활성화**: `interactive.go`(line 26)에서 `cfg.DebateRounds >= 2`일 때만 `runInteractiveDebate()`로 라우팅한다. rounds=0이 전달되므로 이 조건이 항상 false가 되어, 이미 구현된 `interactive_debate.go`의 멀티턴 루프(`runPaneDebate`, `executeRound`)에 도달하지 못한다.

3. **ReadScreen UI 오염**: 응답 캡처 시 `ReadScreen()`이 터미널 pane의 원시 출력을 그대로 반환한다. 여기에는 프롬프트 문자(`>`, `$`), ANSI 이스케이프 시퀀스, 상태바 등 UI chrome이 포함된다. 이 오염된 출력이 rebuttal 프롬프트에 주입되면 프로바이더가 UI 아티팩트를 응답 내용으로 오해하거나 prompt injection 벡터로 악용될 수 있다.

**영향:**
- `auto orchestra brainstorm --strategy debate`는 단일 라운드만 실행하고 종료
- debate 인프라(`interactive_debate.go`, `debate.go`)가 brainstorm에서 사용 불가
- SPEC-ORCH-008에서 구현한 핑퐁 토론 기능이 brainstorm 경로에서 죽은 코드로 남아있음

---

## 2. Requirements

모든 요구사항은 P0 (Must Have) 등급이며 EARS 형식을 따른다.

### REQ-1: Brainstorm debate 기본 라운드 적용

> WHEN brainstorm 커맨드가 debate 전략을 사용할 때,
> THE SYSTEM SHALL `resolveRounds()`를 호출하여 기본 2라운드를 적용한다.

**수락 기준:**
- `orchestra_brainstorm.go`에서 `resolveRounds(flagStrategy, 0)`를 호출하고 결과를 rounds 파라미터로 전달
- `--rounds N` 플래그가 명시적으로 전달된 경우 해당 값을 우선 적용
- 기존 `plan` 커맨드의 패턴(`orchestra.go` line 90)과 동일한 호출 구조 사용

**영향 파일:**
- `autopus-adk/internal/cli/orchestra_brainstorm.go`

### REQ-2: Rebuttal 프롬프트 연결 검증

> WHEN debate 라운드가 완료되면,
> THE SYSTEM SHALL 이전 라운드의 다른 프로바이더 응답을 rebuttal 프롬프트에 포함하여 다음 라운드에 전달한다.

**수락 기준:**
- `cfg.DebateRounds >= 2` 조건이 brainstorm 경로에서 true로 평가됨
- `runInteractiveDebate()`가 호출되고, 각 라운드에서 `buildRebuttalPrompt()`로 교차 응답이 주입됨
- 2라운드 기본값에서 최소 1회의 rebuttal이 발생

**영향 파일:**
- `autopus-adk/pkg/orchestra/interactive.go` (라우팅 — 기존 코드 변경 불필요, REQ-1 수정으로 자동 활성화)
- `autopus-adk/pkg/orchestra/interactive_debate.go` (실행 — 기존 코드 변경 불필요)
- `autopus-adk/pkg/orchestra/debate.go` (rebuttal 빌드 — 기존 코드 변경 불필요)

### REQ-3: ReadScreen 출력 정제

> WHEN ReadScreen으로 프로바이더 응답을 수집할 때,
> THE SYSTEM SHALL UI chrome(프롬프트 문자, ANSI 이스케이프 시퀀스, 상태바)을 제거한 정제된 텍스트만 반환한다.

**수락 기준:**
- ANSI 이스케이프 시퀀스(`\x1b[...m` 등)가 제거됨
- 프롬프트 패턴(`> `, `$ `, `% ` 등 행 시작 패턴)이 제거됨
- 빈 행과 trailing whitespace가 정리됨
- 정제 함수는 독립적으로 테스트 가능한 순수 함수로 구현

**영향 파일:**
- `autopus-adk/pkg/orchestra/screen_sanitizer.go` (신규)
- `autopus-adk/pkg/orchestra/screen_sanitizer_test.go` (신규)
- `autopus-adk/pkg/orchestra/interactive.go` (ReadScreen 호출 후 정제 함수 적용)

---

## 3. Technical Notes

### 기존 인프라 활용

핵심 수정은 `orchestra_brainstorm.go`의 1줄 변경이다. `resolveRounds()` 호출을 추가하면 기존 debate 인프라가 자동으로 활성화된다:

```
// Before (line 31)
return runOrchestraCommand(..., 0, ...)

// After
resolvedRounds := resolveRounds(flagStrategy, 0)
return runOrchestraCommand(..., resolvedRounds, ...)
```

`--rounds` 플래그도 brainstorm 커맨드에 추가하여 사용자가 라운드 수를 제어할 수 있도록 한다 (`plan` 커맨드와 동일 패턴).

### 영향 범위

| 파일 | 변경 유형 | 변경량 |
|------|----------|--------|
| `internal/cli/orchestra_brainstorm.go` | 수정 | ~10줄 (resolveRounds 호출 + --rounds 플래그 추가) |
| `pkg/orchestra/screen_sanitizer.go` | 신규 | ~50줄 (ANSI strip + prompt strip + trim) |
| `pkg/orchestra/screen_sanitizer_test.go` | 신규 | ~80줄 (테이블 드리븐 테스트) |
| `pkg/orchestra/interactive.go` | 수정 | ~3줄 (ReadScreen 결과에 sanitizer 적용) |

### 활성화되는 기존 코드 경로

REQ-1 수정으로 다음 코드 경로가 brainstorm에서 자동 활성화:

1. `interactive.go:26` — `DebateRounds >= 2` 조건 통과
2. `interactive_debate.go` — `runInteractiveDebate()` → `runPaneDebate()` → `executeRound()` 루프
3. `debate.go` — `buildRebuttalPrompt()` 호출로 교차 응답 주입
4. `round_signal.go` — 라운드별 시그널 파일 관리

---

## 4. Out of Scope

- **새 프로바이더 CLI 지원**: 기존 등록된 프로바이더만 대상
- **Hook 자동 설치 자동화**: SPEC-ORCH-008 범위, 본 SPEC에서는 hook이 이미 설치된 환경을 전제
- **Judge 라운드 개선**: 기존 judge 로직 그대로 사용
- **UI/UX 변경**: 터미널 레이아웃, pane 배치 등은 변경하지 않음
- **성능 최적화**: ReadScreen 폴링 간격, 타임아웃 조정 등은 별도 SPEC으로

---

## 5. Key Q&A

### Q1: ReadScreen 정제 시 프롬프트 패턴 오탐이 발생하면?

프로바이더 응답 본문에 `> `로 시작하는 인용문이 포함될 수 있다. 정제 함수는 **행 시작 위치의 단일 프롬프트 문자만** 제거하고, markdown 인용문(`> text`)은 보존해야 한다. 구체적으로:
- `> ` 단독 행(응답 없이 프롬프트만): 제거
- `> some text` (1단어 이상 뒤따름): CLI 프롬프트와 markdown 인용 구분이 어려우므로, 첫 번째와 마지막 행의 프롬프트 패턴만 제거하고 중간 행은 보존

### Q2: Opencode prompt injection 위험은?

ReadScreen에서 수집한 원시 텍스트를 그대로 rebuttal 프롬프트에 주입하면, 프로바이더 A의 응답에 포함된 악의적 지시("ignore previous instructions...")가 프로바이더 B에 전달될 수 있다. 대응:
- REQ-3의 sanitizer가 제어 문자와 이스케이프 시퀀스를 제거하여 1차 방어
- Rebuttal 프롬프트 템플릿에서 응답 내용을 명시적으로 `[RESPONSE FROM {provider}]` 블록으로 감싸 컨텍스트 분리
- 근본적 해결(샌드박스 프롬프트)은 향후 SPEC으로 분리

### Q3: brainstorm에 --rounds 플래그를 추가하면 기존 사용자에게 영향이 있는가?

없다. 기본값은 `resolveRounds()`가 debate 전략에 대해 2를 반환하므로, 플래그를 명시하지 않으면 현재와 동일하게 동작한다(단, 현재는 rounds=0으로 debate가 비활성화된 상태이므로, 수정 후 debate가 실제로 작동하기 시작하는 것이 유일한 행동 변화).
