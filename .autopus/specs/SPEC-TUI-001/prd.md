# PRD: SPEC-TUI-001 — `auto init` TUI Interactive Upgrade

**Status**: draft
**Created**: 2026-03-31
**Domain**: TUI
**Mode**: Standard (10+ sections)
**Predecessor**: SPEC-INITUX-001 (completed)

---

## 1. Problem & Context

### 현재 상태

`auto init`의 인터랙티브 프롬프트는 `bufio.Reader` 기반 번호 입력 방식을 사용한다.
사용자는 옵션을 읽고 번호를 타이핑한 후 Enter를 눌러야 하며, 선택 중 시각적 피드백이 없다.

**현재 구현 분석** (7개 파일, 총 ~801줄):

| 파일 | 줄 수 | 역할 |
|------|-------|------|
| `internal/cli/init.go` | 169 | 5단계 wizard flow orchestration |
| `internal/cli/init_helpers.go` | 79 | platform file generation, gitignore |
| `internal/cli/prompts.go` | 231 | `promptYesNo()`, `promptChoice()` — 번호 입력 기반 |
| `internal/cli/tui/banner.go` | 58 | banner, section header |
| `internal/cli/tui/status.go` | 104 | status messages (Success, Error, Step 등) |
| `internal/cli/tui/box.go` | 76 | styled boxes |
| `internal/cli/tui/style.go` | 84 | lipgloss styling, color definitions |

**핵심 문제점:**

1. **키보드 네비게이션 부재**: 화살표 키로 옵션을 탐색할 수 없고, 번호를 직접 타이핑해야 함
2. **선택 중 시각적 피드백 없음**: 현재 hover된 옵션이 무엇인지 알 수 없음
3. **진행 상태 불명확**: `tui.Step()`이 존재하지만 전체 flow에서 현재 위치가 시각적으로 부족
4. **에러 복구 미흡**: 잘못된 입력 시 graceful recovery가 없음
5. **모던 TUI 기대치 불충족**: `gh`, `charm`, `lazygit` 등 모던 CLI 도구 대비 UX 격차

### 기술적 배경

- 현재 dependency: `lipgloss v1.1.0` (Charmbracelet ecosystem)
- `bubbletea`, `bubbles`, `huh` 미사용 — 이들은 lipgloss와 같은 Charmbracelet 생태계
- Go 1.26, cobra v1.9.1
- Non-TTY 환경(CI, Claude Code Bash tool) 대응 코드 존재 (`isStdinTTY()`, `EnsureSafeEnv()`)

---

## 2. Goals & Success Metrics

### Goals

| ID | Goal | 측정 방식 |
|----|------|----------|
| G1 | 사용자가 키보드만으로 모든 설정을 완료 | 번호 타이핑 0회 |
| G2 | init 완료 시간 단축 | 표준 설정 기준 60초 이내 |
| G3 | 진행 상태를 항상 인지 | 모든 화면에 step indicator 표시 |
| G4 | Non-TTY 환경 무중단 | `--yes` 및 non-TTY 자동 감지로 fallback |
| G5 | Autopus 브랜딩 일관성 | 기존 color palette, 스타일 100% 재사용 |

### Success Metrics

- **M1**: interactive init에서 사용자의 manual text input이 0회 (선택은 모두 키보드 네비게이션)
- **M2**: init wizard 평균 완료 시간 < 45초 (5단계 기준)
- **M3**: non-TTY 환경에서 hang 또는 panic 발생 0건
- **M4**: 기존 `--yes`, `--quality`, `--no-review-gate` 플래그 모두 정상 동작

---

## 3. Target Users

### Primary

- **첫 설치 개발자**: autopus-adk를 처음 프로젝트에 설치하는 개발자
  - 기대: 빠르고 직관적인 초기 설정, 옵션의 의미를 즉시 파악
  - 환경: macOS/Linux 터미널, iTerm2, Terminal.app, Warp, VS Code terminal

### Secondary

- **CI/CD 파이프라인**: non-interactive 모드로 init을 실행하는 자동화 환경
  - 기대: `--yes` 플래그로 무인 실행, exit code 기반 성공/실패 판단
- **재설정 개발자**: 기존 설정을 변경하려는 사용자
  - 기대: 기존 값이 pre-selected 된 상태로 wizard가 시작

---

## 4. User Stories / Job Stories

### JTBD Format

**JS-1**: WHEN I run `auto init` in my project, I WANT TO navigate options with arrow keys and confirm with Enter, SO THAT I can set up the harness without typing option numbers.

**JS-2**: WHEN I'm on a specific init step, I WANT TO see which step I'm on and how many remain, SO THAT I know the progress and can estimate time to completion.

**JS-3**: WHEN I accidentally select the wrong option, I WANT TO go back to the previous step, SO THAT I can correct my choice without restarting the entire wizard.

**JS-4**: WHEN I run `auto init` in a CI pipeline (non-TTY), I WANT the init to complete automatically with sensible defaults, SO THAT my pipeline doesn't hang waiting for input.

**JS-5**: WHEN I'm selecting a language, I WANT TO see all available options highlighted with my current selection visually distinct, SO THAT I can quickly identify and confirm my choice.

**JS-6**: WHEN init completes, I WANT TO see a summary of all my selections in a branded box, SO THAT I can verify the configuration before it's applied.

**JS-7**: WHEN I provide `--quality ultra` as a flag, I WANT that step to be skipped entirely, SO THAT the wizard only asks about unconfigured settings.

---

## 5. Functional Requirements (MoSCoW, EARS)

### P0 — Must Have

**FR-01**: WHEN the user runs `auto init` in a TTY terminal, THE SYSTEM SHALL present each configuration step using interactive select components with keyboard navigation (arrow keys for movement, Enter for confirmation).

**FR-02**: WHEN the user is on any init step, THE SYSTEM SHALL display a progress indicator showing the current step number, total steps, and step name (e.g., `[2/5] Quality Gate`).

**FR-03**: WHEN stdin is not a TTY or `--yes` flag is provided, THE SYSTEM SHALL skip all interactive prompts and use default values, maintaining full backward compatibility with the existing non-interactive behavior.

**FR-04**: WHEN the user selects an option in a select component, THE SYSTEM SHALL visually highlight the currently focused option with the Autopus brand color (`ColorViolet #7c3aed`) and show a cursor indicator.

**FR-05**: WHEN the init wizard presents a choice, THE SYSTEM SHALL show option descriptions inline (e.g., "Ultra — all agents use Opus" next to the option) without requiring the user to remember numbered indices.

**FR-06**: WHEN the user presses `q` or `Ctrl+C` at any point during the wizard, THE SYSTEM SHALL gracefully exit without writing partial configuration, displaying a cancellation message.

### P1 — Should Have

**FR-07**: WHEN transitioning between init steps, THE SYSTEM SHALL animate the transition with a brief visual effect (e.g., fade or slide) to provide a polished user experience.

**FR-08**: THE SYSTEM SHALL apply Autopus branding consistently across all TUI components, using the existing color palette defined in `tui/style.go` (`ColorViolet`, `ColorPink`, `ColorSuccess`, etc.).

**FR-09**: WHEN a step has a pre-configured value (from flags or existing `autopus.yaml`), THE SYSTEM SHALL pre-select that value as the default in the select component.

**FR-10**: WHEN the init wizard completes all steps, THE SYSTEM SHALL display a confirmation screen showing all selected values in a styled summary before writing the configuration.

### P2 — Could Have

**FR-11**: WHEN the init wizard reaches the final step, THE SYSTEM SHALL display a preview of the generated `autopus.yaml` content before confirmation.

**FR-12**: WHEN the user presses `Backspace` or a designated "back" key, THE SYSTEM SHALL navigate to the previous step with the previously selected value preserved.

### P3 — Won't Have (this iteration)

**FR-13**: Multi-page form (huh form groups) — 단일 페이지에서 모든 옵션을 한 번에 편집하는 UI는 이번 iteration에서 구현하지 않음.

---

## 6. Non-Functional Requirements

### Performance

- **NFR-01**: init wizard의 시작 시간(첫 화면 렌더링까지)은 500ms 이내여야 한다.
- **NFR-02**: step 간 전환 시간은 100ms 이내여야 한다 (애니메이션 제외).
- **NFR-03**: bubbletea model의 메모리 사용량은 추가 10MB를 초과하지 않아야 한다.

### Accessibility

- **NFR-04**: `NO_COLOR` 환경변수 설정 시 모든 색상을 비활성화하고 plain text로 표시해야 한다.
- **NFR-05**: 256-color 미지원 터미널에서도 readable한 fallback 렌더링을 제공해야 한다.

### Terminal Compatibility

- **NFR-06**: 최소 지원 터미널 목록:
  - macOS: Terminal.app, iTerm2, Warp, Ghostty
  - Linux: GNOME Terminal, Konsole, Alacritty, kitty
  - Multiplexer: tmux, screen
  - IDE: VS Code integrated terminal, JetBrains terminal
- **NFR-07**: minimum terminal width 40 columns에서 정상 렌더링되어야 한다 (기존 `bannerWidth = 40` 기준).

### Reliability

- **NFR-08**: bubbletea program의 panic은 recover되어 graceful error message로 전환되어야 한다.
- **NFR-09**: SIGINT/SIGTERM 수신 시 terminal state를 원래대로 복원해야 한다 (alternate screen buffer cleanup).

---

## 7. Technical Constraints

### Library Constraints

| Constraint | Detail |
|------------|--------|
| **TC-01**: Charmbracelet 생태계 필수 | lipgloss v1.x 이미 채택, bubbletea/bubbles/huh는 같은 생태계 |
| **TC-02**: `charmbracelet/huh` 우선 검토 | huh는 form/wizard 패턴에 최적화된 high-level library |
| **TC-03**: Go 1.23+ | go.mod에 go 1.26 명시 |
| **TC-04**: 300줄 파일 제한 | 모든 소스 파일은 300줄 이하, 200줄 미만 권장 |

### Architecture Constraints

| Constraint | Detail |
|------------|--------|
| **TC-05**: 기존 `tui/` 패키지 유지 | `style.go`, `banner.go`, `status.go`, `box.go`는 다른 커맨드에서도 사용 — 파괴적 변경 금지 |
| **TC-06**: cobra 커맨드 구조 유지 | `newInitCmd()` 함수 시그니처 및 flag 정의 유지 |
| **TC-07**: `config.Save()` 호출 패턴 유지 | 각 step마다 intermediate save가 아닌, 최종 확인 후 한 번에 저장하는 것도 허용 |
| **TC-08**: Non-TTY 감지 로직 재사용 | 기존 `isStdinTTY()`, `EnsureSafeEnv()` 활용 |

### Migration Constraints

| Constraint | Detail |
|------------|--------|
| **TC-09**: `prompts.go`의 `promptChoice()`, `promptYesNo()` 교체 | 다른 커맨드에서 사용하지 않는다면 제거, 사용한다면 deprecated wrapper 유지 |
| **TC-10**: 기존 flag 100% 호환 | `--yes`, `--quality`, `--no-review-gate`, `--platforms`, `--project`, `--dir` 모두 동일 동작 |

---

## 8. Out of Scope

| ID | 제외 항목 | 사유 |
|----|----------|------|
| **OOS-01** | `auto update` 커맨드의 TUI 업그레이드 | 별도 SPEC으로 분리 |
| **OOS-02** | `auto doctor` 커맨드의 TUI 업그레이드 | 별도 SPEC으로 분리 |
| **OOS-03** | Custom theme 시스템 (user-configurable colors) | 현재 Autopus branding만 지원 |
| **OOS-04** | Mouse click 지원 | 키보드 네비게이션만 구현 |
| **OOS-05** | Wizard state persistence (중단 후 재개) | init은 짧은 프로세스이므로 불필요 |

---

## 9. Risks & Open Questions

### Risks

| ID | Risk | Impact | Likelihood | Mitigation |
|----|------|--------|------------|------------|
| **R1** | bubbletea 학습 곡선으로 구현 지연 | Medium | Medium | huh library는 bubbletea 위의 high-level abstraction — 학습 부담 감소 |
| **R2** | 일부 터미널에서 렌더링 깨짐 | High | Low | `NO_COLOR` fallback, 최소 width 검증, CI에서 non-TTY 자동 감지 |
| **R3** | alternate screen buffer 사용 시 출력 손실 | Medium | Medium | alternate screen 미사용 또는 선택적 사용 권장 |
| **R4** | lipgloss v1.x의 OSC 11 hang 이슈 재발 | High | Low | 기존 `EnsureSafeEnv()` 패턴 유지, bubbletea도 동일 패턴 적용 |
| **R5** | `prompts.go` 제거 시 다른 커맨드 영향 | Medium | Low | 사용처 전수 조사 후 결정 |

### Open Questions

| ID | Question | Owner | Deadline |
|----|----------|-------|----------|
| **OQ-1** | `charmbracelet/huh`와 raw `bubbletea+bubbles` 중 어느 것을 사용할 것인가? | Tech Lead | SPEC review 시 |
| **OQ-2** | alternate screen buffer를 사용할 것인가? (wizard를 전체 화면으로 표시) | UX | SPEC review 시 |
| **OQ-3** | Language Settings에서 4개 언어 외 추가 언어 지원이 필요한가? | PM | v2 결정 |
| **OQ-4** | `promptChoice()`/`promptYesNo()`를 다른 커맨드에서 사용하는지 확인 필요 | Dev | 구현 전 |

---

## 10. Pre-mortem

> "이 기능이 6개월 후 실패했다면, 그 이유는 무엇이었을까?"

### Scenario 1: "터미널 호환성 지옥"
bubbletea 기반 TUI가 특정 터미널/multiplexer 조합에서 깨져서, 사용자들이 `--yes` 모드만 사용하게 됨. Interactive wizard가 실질적으로 사용되지 않는 dead code가 됨.
**예방**: 주요 터미널 5종에서 수동 테스트, CI에서 non-TTY 자동 감지 테스트 필수.

### Scenario 2: "Non-TTY regression"
bubbletea의 `tea.Program` 시작 로직이 non-TTY 환경에서 hang을 일으킴. 기존 `EnsureSafeEnv()` 우회 코드가 bubbletea에는 적용되지 않아 CI 파이프라인이 중단됨.
**예방**: non-TTY 분기에서는 bubbletea를 절대 시작하지 않고, 기존 plain-text fallback 경로를 유지.

### Scenario 3: "의존성 폭발"
bubbletea + bubbles + huh 의존성 추가로 바이너리 크기가 크게 증가하고, 의존성 충돌이나 보안 취약점이 발생.
**예방**: `go mod tidy` 후 바이너리 크기 비교, 의존성 트리 검토.

### Scenario 4: "300줄 제한과의 충돌"
bubbletea Model의 `Init()`, `Update()`, `View()` 메서드가 복잡해지면서 파일이 300줄을 초과. 인위적 분할로 가독성이 오히려 저하됨.
**예방**: step별로 별도 Model 파일 분리 (e.g., `tui/init_language.go`, `tui/init_quality.go`), 공통 로직은 shared helper로 추출.

### Scenario 5: "huh API 변경"
huh library가 아직 v0.x이므로 breaking change가 발생하여 유지보수 부담 증가.
**예방**: huh 버전 pinning, 자체 wrapper layer로 격리, 대안으로 raw bubbletea+bubbles 검토.

---

## 11. Practitioner Q&A

### Q1: huh vs raw bubbletea — 어떤 것을 선택해야 하는가?

**A**: `charmbracelet/huh`를 우선 검토한다. huh는 form/wizard 패턴에 특화된 high-level library로, select, multi-select, confirm 등의 컴포넌트를 선언적으로 구성할 수 있다. init wizard의 요구사항(5단계 sequential form)에 정확히 부합한다. 만약 huh의 customization 한계(예: step 간 애니메이션, custom progress bar)가 있다면 해당 부분만 raw bubbletea로 구현한다.

### Q2: 기존 `prompts.go`는 어떻게 처리하는가?

**A**: 구현 전에 `promptChoice()`와 `promptYesNo()`의 사용처를 전수 조사한다. `init.go` 외에 사용하는 곳이 없다면 제거한다. 다른 커맨드에서 사용한다면 deprecated 마킹 후 유지하되, 새로운 TUI 컴포넌트로의 마이그레이션 이슈를 별도 생성한다.

### Q3: bubbletea `tea.Program`과 cobra의 통합은 어떻게 하는가?

**A**: cobra의 `RunE` 함수 내에서 `tea.NewProgram(model).Run()`을 호출한다. TTY 체크 후 non-TTY이면 기존 default-value 경로를 타고, TTY이면 bubbletea program을 시작한다. bubbletea program 종료 후 반환된 final model에서 사용자 선택값을 추출하여 `config.Save()`에 전달한다.

### Q4: 파일 구조는 어떻게 설계하는가?

**A**: 300줄 제한을 고려한 파일 분할 전략:

```
internal/cli/tui/
├── style.go          (84줄, 유지)
├── banner.go         (58줄, 유지)
├── status.go         (104줄, 유지)
├── box.go            (76줄, 유지)
├── wizard.go         (NEW — wizard orchestration model, ~150줄)
├── wizard_steps.go   (NEW — step definitions & views, ~200줄)
└── wizard_styles.go  (NEW — wizard-specific styles, ~80줄)

internal/cli/
├── init.go           (169줄 → ~120줄 — wizard 호출로 간소화)
├── init_helpers.go   (79줄, 유지)
└── prompts.go        (231줄 → 제거 또는 deprecated)
```

### Q5: alternate screen buffer를 사용해야 하는가?

**A**: 사용하지 않는 것을 권장한다. init wizard는 짧은 프로세스(5단계)이며, 완료 후 설정 summary가 터미널 히스토리에 남는 것이 사용자에게 유용하다. alternate screen을 사용하면 완료 후 결과가 사라져 불편하다. 단, `WithAltScreen()` 옵션은 향후 설정으로 제공 가능.

### Q6: 테스트는 어떻게 작성하는가?

**A**: bubbletea의 `tea.ProgramTest` 또는 `teatest` 패키지를 활용하여 키 입력 시뮬레이션 기반 통합 테스트를 작성한다. 각 step의 Model은 순수 함수(`Update`)로 구현되므로 unit test도 가능하다. non-TTY fallback 경로는 기존 테스트 패턴 유지.

---

## Appendix: Current Dependency Tree (Charmbracelet)

```
현재:
  github.com/charmbracelet/lipgloss v1.1.0

추가 예정:
  github.com/charmbracelet/bubbletea  (lipgloss 의존)
  github.com/charmbracelet/bubbles    (bubbletea 의존)
  github.com/charmbracelet/huh        (bubbletea + bubbles 의존, 검토 후 결정)
```

## Appendix: Current Init Flow vs Target Flow

```
현재 (bufio.Reader):
  ┌──────────────┐    번호 입력     ┌──────────────┐    번호 입력
  │  Step 1      │ ──────────────→ │  Step 2      │ ──────────────→ ...
  │  텍스트 출력  │   "1" + Enter   │  텍스트 출력  │   "2" + Enter
  └──────────────┘                 └──────────────┘

목표 (bubbletea/huh):
  ┌──────────────────────────────────────┐
  │  [1/5] Language Settings             │
  │                                      │
  │  Code comments language?             │
  │                                      │
  │    ● English           ← selected    │
  │    ○ Korean (한국어)                  │
  │    ○ Japanese (日本語)               │
  │    ○ Chinese (中文)                  │
  │                                      │
  │  ↑↓ navigate  Enter confirm          │
  └──────────────────────────────────────┘
```
