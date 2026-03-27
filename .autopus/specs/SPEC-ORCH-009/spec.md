# SPEC-ORCH-009: Orchestra 멀티턴 토론 프로토콜 개선

**Status**: completed
**Created**: 2026-03-27
**Domain**: ORCH
**Extends**: SPEC-ORCH-008
**Origin**: BS-004

## 목적

Brainstorm 커맨드의 멀티턴 debate가 rounds=0 하드코딩으로 인해 비활성화되어 있다.
SPEC-ORCH-008에서 구현한 멀티턴 핑퐁 인프라(interactive_debate.go, debate.go, round_signal.go)가
brainstorm 경로에서 죽은 코드로 남아있으며, ReadScreen 출력에 ANSI 이스케이프와 UI chrome이
혼입되어 rebuttal 프롬프트 품질을 저하시킨다.

## 요구사항

### P0: REQ-1 — Brainstorm debate 기본 라운드 적용

> WHEN brainstorm 커맨드가 debate 전략을 사용할 때,
> THE SYSTEM SHALL `resolveRounds()`를 호출하여 기본 2라운드를 적용한다.

- `orchestra_brainstorm.go`에서 `resolveRounds(flagStrategy, rounds)`를 호출하고 결과를 `runOrchestraCommand`에 전달
- `plan` 커맨드의 패턴(`orchestra.go` line 90)과 동일한 호출 구조 사용
- rounds=0(미지정) + debate 전략 → 기본값 2 적용

### P0: REQ-2 — Rebuttal 프롬프트 연결 검증

> WHEN debate 라운드가 완료되면,
> THE SYSTEM SHALL 이전 라운드의 다른 프로바이더 응답을 rebuttal 프롬프트에 포함하여 다음 라운드에 전달한다.

- `interactive.go:26`의 `cfg.DebateRounds >= 2` 조건이 brainstorm 경로에서 true로 평가됨
- `runInteractiveDebate()` → `runPaneDebate()` → `executeRound()` 경로가 활성화됨
- REQ-1 수정으로 자동 활성화 — 기존 코드 변경 불필요

### P0: REQ-3 — ReadScreen 출력 정제 강화

> WHEN ReadScreen으로 프로바이더 응답을 수집할 때,
> THE SYSTEM SHALL UI chrome(프롬프트 문자, ANSI 이스케이프 시퀀스, 상태바, OSC 시퀀스)을 제거한 정제된 텍스트만 반환한다.

- ANSI 이스케이프 시퀀스(`\x1b[...m`, CSI, OSC 등) 완전 제거
- 상태바 라인(tmux status-line 패턴) 제거
- trailing whitespace 및 연속 빈 행 정리
- 정제 함수는 독립적으로 테스트 가능한 순수 함수로 구현

### P1: REQ-4 — Brainstorm --rounds 플래그 추가

> WHEN 사용자가 brainstorm 커맨드를 실행할 때,
> THE SYSTEM SHALL `--rounds N` 플래그를 통해 debate 라운드 수를 제어할 수 있도록 한다.

- `plan` 커맨드와 동일한 `--rounds` 플래그 패턴 사용
- 값 범위: 1-10 (기존 validation 로직 재사용)
- 미지정 시 `resolveRounds()`가 debate에 대해 기본값 2 반환

### P2: REQ-5 — 프로바이더별 프롬프트 어댑터 개선

> WHERE 프로바이더가 PromptViaArgs=true 속성을 가질 때,
> THE SYSTEM SHALL 프로바이더별 프롬프트 전달 방식을 최적화하여 프롬프트 수신 실패를 방지한다.

- OpenCode의 `run` 모드와 interactive 모드의 프롬프트 전달 차이 문서화
- `buildInteractiveLaunchCmd()`에서 opencode의 `run` 플래그 스킵 로직 검증
- 향후 프로바이더별 어댑터 인터페이스 설계 방향 제시 (본 SPEC에서는 문서화만)

## 생성 파일 상세

| 파일 | 역할 | 변경량 |
|------|------|--------|
| `internal/cli/orchestra_brainstorm.go` | --rounds 플래그 추가, resolveRounds() 호출 | ~10줄 수정 |
| `pkg/orchestra/screen_sanitizer.go` | 전용 출력 정제 함수 (신규) | ~60줄 |
| `pkg/orchestra/screen_sanitizer_test.go` | 테이블 드리븐 테스트 (신규) | ~80줄 |
| `pkg/orchestra/interactive_detect.go` | cleanScreenOutput()에서 신규 sanitizer 호출 | ~5줄 수정 |
