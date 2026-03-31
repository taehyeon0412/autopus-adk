# Acceptance Criteria: SPEC-TUI-001

## Criteria Matrix

| ID | Requirement | Criterion | Verification Method |
|----|-------------|-----------|---------------------|
| AC-01 | R1 (Interactive Select) | 화살표 키(up/down)로 옵션 탐색 가능 | Manual test + teatest |
| AC-02 | R1 (Interactive Select) | Enter 키로 현재 포커스 옵션 확정 | Manual test + teatest |
| AC-03 | R1 (Interactive Select) | 번호 타이핑 없이 모든 설정 완료 가능 | Manual test |
| AC-04 | R2 (Progress Indicator) | 모든 step에 `[N/5] Step Name` 진행 표시 렌더링 | Visual inspection |
| AC-05 | R2 (Progress Indicator) | step 전환 시 indicator 즉시 갱신 | Visual inspection |
| AC-06 | R3 (Non-TTY Fallback) | non-TTY에서 hang 없이 완료 | Automated test (pipe stdin) |
| AC-07 | R3 (Non-TTY Fallback) | `--yes` 플래그로 모든 프롬프트 건너뜀 | Automated test |
| AC-08 | R3 (Non-TTY Fallback) | non-TTY에서 tea.Program 미시작 | Unit test (isStdinTTY mock) |
| AC-09 | R4 (Visual Highlight) | 포커스 옵션이 `#7c3aed` 색상 강조 | Visual inspection |
| AC-10 | R4 (Visual Highlight) | 커서 인디케이터 표시 | Visual inspection |
| AC-11 | R5 (Inline Descriptions) | 각 옵션에 설명 텍스트 함께 표시 | Visual inspection |
| AC-12 | R5 (Inline Descriptions) | 옵션 번호 비표시 | Visual inspection |
| AC-13 | R6 (Graceful Exit) | Ctrl+C 시 "init cancelled" 메시지 출력 | Manual test |
| AC-14 | R6 (Graceful Exit) | 취소 시 autopus.yaml 미생성/미변경 | Automated test (file check) |
| AC-15 | R6 (Graceful Exit) | 취소 시 terminal state 정상 복원 | Manual test |
| AC-16 | R7 (Step Transition) | step 간 시각적 전환 피드백 존재 | Visual inspection |
| AC-17 | R7 (Step Transition) | 전환 애니메이션 100ms 이내 | Timing measurement |
| AC-18 | R8 (Branding) | 모든 컴포넌트가 tui/style.go 색상 사용 | Code review |
| AC-19 | R8 (Branding) | 새 색상이 style.go 외부에서 정의되지 않음 | Code review |
| AC-20 | R9 (Pre-configured Default) | `--quality ultra` 시 해당 step skip | Automated test |
| AC-21 | R9 (Pre-configured Default) | 기존 autopus.yaml 값 pre-selected | Manual test |
| AC-22 | R10 (Completion Summary) | 선택값 요약이 branded box에 표시 | Visual inspection |
| AC-23 | R10 (Completion Summary) | 기존 SummaryTable 컴포넌트 재사용 | Code review |
| AC-24 | R11 (YAML Preview) | 최종 확인 전 YAML 프리뷰 렌더링 | Visual inspection |
| AC-25 | R11 (YAML Preview) | 프리뷰와 실제 저장 내용 동일 | Automated test (diff) |
| AC-26 | R12 (Back Navigation) | 뒤로가기 시 이전 step 이동 | Manual test + teatest |
| AC-27 | R12 (Back Navigation) | 이전 step 기존 선택값 유지 | Manual test + teatest |
| AC-28 | R12 (Back Navigation) | 첫 step에서 뒤로가기 시 무동작 | Manual test + teatest |
| AC-29 | R13 (Flag Compatibility) | 기존 플래그 모두 동일 동작 | Automated test (all flags) |
| AC-30 | R13 (Flag Compatibility) | 플래그 지정 step 자동 skip | Automated test |

## Priority Classification

### P0 -- Must Pass (Release Blocker)

| AC IDs | Description |
|--------|-------------|
| AC-01 ~ AC-03 | 키보드 네비게이션 기본 동작 |
| AC-06 ~ AC-08 | non-TTY 안전성 (CI 파이프라인 차단 방지) |
| AC-13 ~ AC-15 | graceful exit (데이터 손실 방지) |
| AC-29, AC-30 | 기존 플래그 호환성 (regression 방지) |

### P1 -- Should Pass

| AC IDs | Description |
|--------|-------------|
| AC-04, AC-05 | progress indicator |
| AC-09 ~ AC-12 | 시각적 강조 및 inline description |
| AC-18, AC-19 | 브랜딩 일관성 |
| AC-20, AC-21 | pre-configured defaults |
| AC-22, AC-23 | completion summary |

### P2 -- Nice to Have

| AC IDs | Description |
|--------|-------------|
| AC-16, AC-17 | step transition animation |
| AC-24, AC-25 | YAML preview |
| AC-26 ~ AC-28 | back navigation |
