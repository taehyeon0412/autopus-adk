# SPEC-TUI-001: `auto init` TUI Interactive Upgrade

**Status**: completed
**Created**: 2026-03-31
**Domain**: TUI
**Target Module**: autopus-adk
**PRD**: prd.md

## Overview

`auto init`의 인터랙티브 프롬프트를 `bufio.Reader` 기반 번호 입력 방식에서 Charmbracelet `huh` 라이브러리 기반 키보드 네비게이션 방식으로 업그레이드한다. 화살표 키로 옵션 탐색, Enter로 확인하는 모던 TUI 위자드를 제공하며, non-TTY 환경에서는 기존 default-value fallback을 유지한다.

## Requirements

### R1: Interactive Select Component
WHEN the user runs `auto init` in a TTY terminal, THE SYSTEM SHALL present each configuration step using interactive select components with keyboard navigation (arrow keys for movement, Enter for confirmation).

**Acceptance Criteria:**
- AC1.1: 모든 선택 단계에서 화살표 키(up/down)로 옵션 탐색이 가능하다
- AC1.2: Enter 키로 현재 포커스된 옵션을 확정한다
- AC1.3: 번호 타이핑 없이 모든 설정을 완료할 수 있다

### R2: Progress Indicator
WHEN the user is on any init step, THE SYSTEM SHALL display a progress indicator showing the current step number, total steps, and step name (e.g., `[2/5] Quality Gate`).

**Acceptance Criteria:**
- AC2.1: 모든 step 화면에 `[N/5] Step Name` 형태의 진행 표시가 렌더링된다
- AC2.2: step 전환 시 progress indicator가 즉시 갱신된다

### R3: Non-TTY Fallback
WHEN stdin is not a TTY or `--yes` flag is provided, THE SYSTEM SHALL skip all interactive prompts and use default values, maintaining full backward compatibility with the existing non-interactive behavior.

**Acceptance Criteria:**
- AC3.1: non-TTY 환경에서 `auto init`이 hang 없이 완료된다
- AC3.2: `--yes` 플래그 사용 시 모든 프롬프트가 건너뛰어진다
- AC3.3: non-TTY 모드에서 bubbletea `tea.Program`이 시작되지 않는다

### R4: Visual Highlight
WHEN the user selects an option in a select component, THE SYSTEM SHALL visually highlight the currently focused option with the Autopus brand color (`ColorViolet #7c3aed`) and show a cursor indicator.

**Acceptance Criteria:**
- AC4.1: 현재 포커스된 옵션이 `#7c3aed` 색상으로 강조된다
- AC4.2: 선택되지 않은 옵션과 시각적으로 구분 가능하다
- AC4.3: 커서 인디케이터(e.g., `>` 또는 `●`)가 현재 위치를 나타낸다

### R5: Inline Descriptions
WHEN the init wizard presents a choice, THE SYSTEM SHALL show option descriptions inline (e.g., "Ultra -- all agents use Opus" next to the option) without requiring the user to remember numbered indices.

**Acceptance Criteria:**
- AC5.1: 각 옵션에 설명 텍스트가 함께 표시된다
- AC5.2: 옵션 번호가 표시되지 않는다

### R6: Graceful Exit
WHEN the user presses `q` or `Ctrl+C` at any point during the wizard, THE SYSTEM SHALL gracefully exit without writing partial configuration, displaying a cancellation message.

**Acceptance Criteria:**
- AC6.1: `Ctrl+C` 입력 시 "init cancelled" 메시지가 출력된다
- AC6.2: 취소 시 `autopus.yaml`이 생성/변경되지 않는다
- AC6.3: terminal state가 정상적으로 복원된다

### R7: Step Transition Animation
WHEN transitioning between init steps, THE SYSTEM SHALL animate the transition with a brief visual effect to provide a polished user experience.

**Acceptance Criteria:**
- AC7.1: step 간 전환 시 시각적 피드백(fade, clear+redraw 등)이 있다
- AC7.2: 전환 애니메이션이 100ms를 초과하지 않는다

### R8: Branding Consistency
THE SYSTEM SHALL apply Autopus branding consistently across all TUI components, using the existing color palette defined in `tui/style.go` (`ColorViolet`, `ColorPink`, `ColorSuccess`, etc.).

**Acceptance Criteria:**
- AC8.1: wizard의 모든 컴포넌트가 `tui/style.go`에 정의된 색상을 사용한다
- AC8.2: 새로운 색상이 `tui/style.go` 외부에서 정의되지 않는다

### R9: Pre-configured Value Default
WHEN a step has a pre-configured value (from flags or existing `autopus.yaml`), THE SYSTEM SHALL pre-select that value as the default in the select component.

**Acceptance Criteria:**
- AC9.1: `--quality ultra` 플래그 사용 시 해당 step이 건너뛰어진다
- AC9.2: 기존 `autopus.yaml`의 값이 있으면 해당 옵션이 pre-selected 된다

### R10: Completion Summary
WHEN the init wizard completes all steps, THE SYSTEM SHALL display a confirmation screen showing all selected values in a styled summary before writing the configuration.

**Acceptance Criteria:**
- AC10.1: 모든 step 완료 후 선택된 값 요약이 branded box에 표시된다
- AC10.2: 기존 `SummaryTable` 컴포넌트를 재사용한다

### R11: YAML Preview
WHEN the init wizard reaches the final step, THE SYSTEM SHALL display a preview of the generated `autopus.yaml` content before confirmation.

**Acceptance Criteria:**
- AC11.1: 최종 확인 전 YAML 프리뷰가 렌더링된다
- AC11.2: 프리뷰는 실제 저장될 내용과 동일하다

### R12: Back Navigation
WHEN the user presses `Backspace` or a designated "back" key, THE SYSTEM SHALL navigate to the previous step with the previously selected value preserved.

**Acceptance Criteria:**
- AC12.1: 뒤로가기 시 이전 step으로 이동한다
- AC12.2: 이전 step의 기존 선택값이 유지된다
- AC12.3: 첫 번째 step에서 뒤로가기 시 아무 동작도 하지 않는다

### R13: Flag Backward Compatibility
THE SYSTEM SHALL maintain 100% backward compatibility with existing CLI flags: `--yes`, `--quality`, `--no-review-gate`, `--platforms`, `--project`, `--dir`.

**Acceptance Criteria:**
- AC13.1: 모든 기존 플래그가 동일하게 동작한다
- AC13.2: 플래그로 지정된 값에 해당하는 step은 건너뛰어진다

## Architecture Decisions

### AD1: TUI Framework Selection
**Decision**: `charmbracelet/huh`를 primary framework으로 채택한다.
**Rationale**: huh는 bubbletea 위의 high-level abstraction으로, form/wizard 패턴에 최적화되어 있다. select, confirm 등의 컴포넌트를 선언적으로 구성할 수 있으며, 기존 lipgloss 생태계와 완전 호환된다. init wizard의 요구사항(5단계 sequential select)에 정확히 부합한다.
**Alternatives**: (1) raw bubbletea+bubbles -- 더 높은 customization 가능하지만 boilerplate가 많고 학습 곡선이 높음. (2) raw bufio.Reader 유지 -- 현재 방식으로는 PRD의 핵심 목표(키보드 네비게이션)를 달성할 수 없음.

### AD2: File Structure
**Decision**: `internal/cli/tui/` 하위에 wizard 관련 파일 3개를 신규 생성하고, init.go를 리팩토링한다.
```
internal/cli/tui/
  wizard.go          (기존 유지 -- SummaryTable, WizardHeader)
  wizard_steps.go    (NEW -- step definitions & huh form builders, ~200줄)
  wizard_styles.go   (NEW -- wizard-specific huh theme/styles, ~80줄)
internal/cli/
  init.go            (169줄 -> ~120줄 -- huh wizard 호출로 간소화)
  prompts.go         (231줄 -> deprecated or removed)
```
**Rationale**: 300줄 파일 제한을 준수하면서 관심사를 분리한다. step 정의, 스타일, orchestration을 별도 파일로 나눠 각 파일이 200줄 미만을 유지한다.

### AD3: Alternate Screen Buffer
**Decision**: alternate screen buffer를 사용하지 않는다.
**Rationale**: init wizard는 짧은 프로세스(5단계)이며, 완료 후 설정 summary가 터미널 히스토리에 남는 것이 유용하다. alternate screen 사용 시 결과가 사라져 불편하다.

### AD4: prompts.go 처리 전략
**Decision**: `promptChoice()`와 `promptYesNo()`는 `prompts.go` 내에서만 사용되므로(init.go에서 간접 호출), 전체 파일을 deprecated로 마킹하고, `warnParentRuleConflicts()`는 huh confirm 컴포넌트로 마이그레이션한다.
**Rationale**: 코드베이스 검색 결과 두 함수는 `prompts.go` 내부에서만 호출되며, 외부 패키지나 다른 커맨드에서 참조하지 않는다.

## Dependencies

- `github.com/charmbracelet/huh` (bubbletea + bubbles 의존, form/wizard high-level library)
- `github.com/charmbracelet/bubbletea` (huh의 transitive dependency)
- `github.com/charmbracelet/bubbles` (huh의 transitive dependency)
- 기존 `github.com/charmbracelet/lipgloss v1.1.0` (유지)
- 기존 `golang.org/x/term` (non-TTY 감지용, 유지)
