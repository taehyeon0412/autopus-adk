# Implementation Plan: SPEC-TUI-001

## Task Breakdown

| Task ID | Description | Agent | Mode | File Ownership | Profile | Complexity |
|---------|-------------|-------|------|----------------|---------|------------|
| T1 | Add huh/bubbletea dependencies to go.mod | executor | single | go.mod, go.sum | go | low |
| T2 | Create wizard model and orchestration | executor | single | tui/wizard_steps.go | go | high |
| T3 | Create wizard-specific huh theme/styles | executor | single | tui/wizard_styles.go | go | medium |
| T4 | Refactor init.go to use huh wizard | executor | single | cli/init.go | go | medium |
| T5 | Deprecate prompts.go and migrate warnParentRuleConflicts | executor | single | cli/prompts.go | go | medium |
| T6 | Add non-TTY fallback and cancellation tests | executor | single | tui/wizard_steps_test.go | go | medium |
| T7 | Integration validation and cleanup | executor | single | - | go | low |

## Task Details

### T1: Add huh/bubbletea dependencies
**Description**: `go get github.com/charmbracelet/huh@latest`를 실행하여 huh 및 transitive dependencies(bubbletea, bubbles)를 go.mod에 추가한다. `go mod tidy`로 정리 후 바이너리 크기 변화를 측정한다.
**Files**: `go.mod`, `go.sum`
**Dependencies**: none
**Acceptance Criteria**: go.mod에 huh, bubbletea, bubbles가 direct/indirect로 포함되고 `go build ./...`이 성공한다.

### T2: Create wizard step definitions and huh form builders
**Description**: `internal/cli/tui/wizard_steps.go`에 init wizard의 5개 step(Language, Quality, ReviewGate, Methodology, Confirmation)을 huh `Select` 및 `Confirm` 컴포넌트로 정의한다. 각 step은 `huh.NewSelect()` 또는 `huh.NewConfirm()`으로 생성되며, inline description을 포함한다. 전체 flow는 `huh.NewForm()` 또는 step-by-step `huh.NewGroup()`으로 구성한다. back navigation(R12)을 지원하기 위해 step별 개별 form 실행 + 루프 구조를 채택한다.

**Files**: `internal/cli/tui/wizard_steps.go` (~200줄)
**Dependencies**: T1 (huh dependency 추가 필요)
**Acceptance Criteria**:
- AC1.1, AC1.2, AC1.3: 키보드 네비게이션으로 모든 step 완료 가능
- AC5.1, AC5.2: inline description 표시, 번호 비표시
- AC9.1, AC9.2: 플래그/기존값에 의한 pre-select 및 step skip
- AC12.1, AC12.2, AC12.3: back navigation 동작

**Key Design**:
```go
// InitWizardResult holds all wizard selections.
type InitWizardResult struct {
    CommentsLang  string
    CommitsLang   string
    AILang        string
    Quality       string
    ReviewGate    bool
    Methodology   string
    Cancelled     bool
}

// RunInitWizard runs the interactive init wizard using huh forms.
func RunInitWizard(opts InitWizardOpts) (*InitWizardResult, error)
```

### T3: Create wizard-specific huh theme/styles
**Description**: `internal/cli/tui/wizard_styles.go`에 huh form의 Autopus 브랜드 테마를 정의한다. huh의 `huh.ThemeCharm()` 기반으로 커스텀 테마를 생성하여 `ColorViolet`, `ColorPink` 등 기존 색상을 적용한다. select cursor, focused option, title 등의 스타일을 Autopus 브랜딩에 맞춘다.

**Files**: `internal/cli/tui/wizard_styles.go` (~80줄)
**Dependencies**: T1
**Acceptance Criteria**:
- AC4.1, AC4.2, AC4.3: 포커스 옵션 ColorViolet 강조, 커서 인디케이터
- AC8.1, AC8.2: 기존 색상 팔레트 재사용

### T4: Refactor init.go to use huh wizard
**Description**: `internal/cli/init.go`의 `RunE` 함수를 리팩토링하여, TTY 환경에서는 `tui.RunInitWizard()`를 호출하고 결과를 `config.Save()`에 전달하도록 변경한다. Non-TTY/`--yes` 모드에서는 기존 default-value 로직을 유지한다. step별 `promptXxx()` 호출을 제거하고 wizard 결과 기반으로 cfg를 설정한다.

**Files**: `internal/cli/init.go` (~120줄로 축소)
**Dependencies**: T2, T3
**Acceptance Criteria**:
- AC2.1, AC2.2: progress indicator 표시 (huh 내장 또는 WizardHeader 연동)
- AC3.1, AC3.2, AC3.3: non-TTY fallback 유지
- AC6.1, AC6.2, AC6.3: 취소 시 partial config 미생성
- AC10.1, AC10.2: 완료 summary 표시
- AC13.1, AC13.2: 기존 플래그 호환성 유지

### T5: Deprecate prompts.go and migrate warnParentRuleConflicts
**Description**: `prompts.go`에서 `promptChoice()`와 `promptYesNo()`를 deprecated로 마킹한다. `warnParentRuleConflicts()`는 huh `Confirm` 컴포넌트를 사용하도록 마이그레이션하되, non-TTY fallback을 유지한다. `promptLanguageSettings()`, `promptQualityMode()`, `promptReviewGate()`, `promptMethodology()`는 wizard_steps.go로 대체되므로 제거한다. `isStdinTTY()`는 유틸리티로 유지하거나 별도 파일로 이동한다.

**Files**: `internal/cli/prompts.go` (대폭 축소 또는 제거)
**Dependencies**: T4
**Acceptance Criteria**:
- `promptChoice()`, `promptYesNo()` 호출이 0건
- `warnParentRuleConflicts()`가 huh 기반으로 동작
- 기존 동작과의 backward compatibility 유지

### T6: Add non-TTY fallback and cancellation tests
**Description**: wizard의 non-TTY fallback 경로와 취소 시나리오에 대한 테스트를 작성한다. `teatest` 또는 huh의 테스트 유틸리티를 활용하여 키 입력 시뮬레이션 기반 통합 테스트를 추가한다.

**Files**: `internal/cli/tui/wizard_steps_test.go` (~150줄)
**Dependencies**: T2, T3
**Acceptance Criteria**:
- non-TTY 환경에서 wizard가 시작되지 않는 것을 검증
- 취소 시 `Cancelled: true` 반환 검증
- 각 step의 default value 선택 검증

### T7: Integration validation and cleanup
**Description**: 전체 flow의 end-to-end 검증. `go build ./...`, `go test ./...` 통과 확인. 불필요한 import 정리. 바이너리 크기 비교. 주요 터미널(iTerm2, Terminal.app, VS Code terminal)에서 수동 확인 체크리스트 실행.

**Files**: -
**Dependencies**: T1-T6
**Acceptance Criteria**:
- `go build ./...` 성공
- `go test ./...` 성공
- 모든 신규 파일 300줄 미만
- 바이너리 크기 증가 5MB 이하

## Execution Order

```
T1 (dependencies)
 |
 ├── T3 (styles -- independent)
 └── T2 (steps -- can start after T1)
      |
      T4 (init.go refactor -- needs T2, T3)
      |
      T5 (prompts.go deprecation -- needs T4)
      |
      T6 (tests -- needs T2, T3)
      |
      T7 (validation -- needs all)
```

T2 and T3 can run in parallel after T1.
