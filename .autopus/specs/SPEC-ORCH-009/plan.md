# SPEC-ORCH-009 구현 계획

## 태스크 목록

### Phase 1: P0 — 핵심 연결 수정 (1줄 변경으로 인프라 활성화)

- [ ] T1: `orchestra_brainstorm.go`에서 `resolveRounds()` 호출 추가
  - line 31의 `0` 대신 `resolveRounds(flagStrategy, 0)` 호출
  - `rounds` 변수 선언 및 `runOrchestraCommand`에 전달
  - 변경량: ~3줄

- [ ] T2: 기존 테스트 통과 확인
  - `orchestra_brainstorm_test.go` 실행하여 기존 테스트 깨지지 않음 확인
  - `interactive_debate_test.go` 실행하여 debate 경로 테스트 통과 확인

### Phase 2: P0 — ReadScreen 출력 정제 강화

- [ ] T3: `screen_sanitizer.go` 신규 생성
  - `SanitizeScreenOutput(raw string) string` — 공개 진입점
  - `stripANSIExtended(s string) string` — CSI, OSC, DCS 등 확장 이스케이프 제거
  - `stripStatusBar(s string) string` — tmux/터미널 상태바 라인 제거
  - `collapseBlankLines(s string) string` — 연속 빈 행 정리
  - `trimTrailingWhitespace(s string) string` — 행별 trailing whitespace 제거

- [ ] T4: `screen_sanitizer_test.go` 신규 생성
  - 테이블 드리븐 테스트: ANSI 색상 코드, CSI 커서 이동, OSC 시퀀스
  - 상태바 패턴 제거 검증
  - markdown 인용문(`> text`) 보존 검증
  - 빈 입력, 정상 텍스트(변경 없음) 엣지 케이스

- [ ] T5: `interactive_detect.go`의 `cleanScreenOutput()` 수정
  - 기존 `stripANSI()` + `filterPromptLines()` 대신 `SanitizeScreenOutput()` 호출
  - 기존 `stripANSI()`, `filterPromptLines()` 함수는 하위 호환성을 위해 유지

### Phase 3: P1 — Brainstorm --rounds 플래그

- [ ] T6: `orchestra_brainstorm.go`에 `--rounds` 플래그 추가
  - `rounds int` 변수 선언
  - `cmd.Flags().IntVar(&rounds, "rounds", 0, ...)` 등록
  - `resolveRounds(flagStrategy, rounds)` 호출에 사용자 입력값 전달
  - `plan` 커맨드의 패턴과 동일한 구조

- [ ] T7: brainstorm --rounds 테스트 추가
  - `--rounds 3 --strategy debate`에서 rounds=3이 전달되는지 검증
  - `--rounds 0`(기본값)에서 debate 전략 시 rounds=2 되는지 검증

### Phase 4: P2 — 프로바이더 어댑터 문서화

- [ ] T8: research.md에 프로바이더별 프롬프트 전달 방식 분석 기록
  - OpenCode `run` 모드 vs interactive 모드 차이
  - `buildInteractiveLaunchCmd()`의 `run` 플래그 스킵이 올바른지 검증
  - 향후 어댑터 인터페이스 설계 방향 제시

## 구현 전략

### 핵심 원칙: 최소 변경으로 기존 인프라 활성화

이 SPEC의 핵심 수정은 `orchestra_brainstorm.go`의 1줄 변경이다. `resolveRounds()` 호출만 추가하면 다음 기존 코드 경로가 자동 활성화된다:

1. `interactive.go:26` — `DebateRounds >= 2` 조건 통과
2. `interactive_debate.go` — `runInteractiveDebate()` → `runPaneDebate()` → `executeRound()` 루프
3. `debate.go` — `buildRebuttalPrompt()` 호출로 교차 응답 주입
4. `round_signal.go` — 라운드별 시그널 파일 관리

### 기존 코드 활용

- `resolveRounds()` — `orchestra_helpers.go:71`에 이미 구현됨
- `--rounds` 플래그 패턴 — `orchestra.go` plan 커맨드에서 이미 사용 중
- `stripANSI()` — `interactive_detect.go`에 기본 구현 있음, 확장 필요
- debate 인프라 전체 — SPEC-ORCH-008에서 구현 완료

### 변경 범위

총 수정/신규 파일: 4개 (+ 1개 테스트)
총 변경량: ~155줄 (신규 ~140줄 + 수정 ~15줄)
