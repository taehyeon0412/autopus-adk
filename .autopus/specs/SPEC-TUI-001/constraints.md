# Constraints: SPEC-TUI-001

## Hard Constraints

- **C1**: 모든 소스 파일은 300줄 이하 (프로젝트 file-size-limit 규칙)
- **C2**: Non-TTY 환경에서 bubbletea `tea.Program`을 시작하지 않아야 한다 (hang/panic 방지)
- **C3**: `EnsureSafeEnv()` 패턴을 유지하여 lipgloss OSC 11 hang 이슈를 방지해야 한다
- **C4**: 기존 CLI 플래그(`--yes`, `--quality`, `--no-review-gate`, `--platforms`, `--project`, `--dir`) 100% 호환
- **C5**: cobra 커맨드 구조 유지 -- `newInitCmd()` 함수 시그니처 및 flag 정의 변경 금지
- **C6**: 기존 `tui/` 패키지의 공개 API(`style.go`, `banner.go`, `status.go`, `box.go`) 파괴적 변경 금지
- **C7**: `config.Save()` 호출은 모든 step 완료 후 한 번만 수행 (partial save 금지)
- **C8**: Charmbracelet 생태계 라이브러리만 사용 (lipgloss, bubbletea, bubbles, huh)
- **C9**: alternate screen buffer 미사용 -- wizard 결과가 터미널 히스토리에 남아야 함

## Soft Constraints

- **C10**: 바이너리 크기 증가 5MB 이하 (huh + bubbletea + bubbles 의존성 추가 기준)
- **C11**: init wizard 시작 시간(첫 화면 렌더링) 500ms 이내
- **C12**: step 간 전환 시간 100ms 이내 (애니메이션 제외)
- **C13**: bubbletea model 메모리 사용량 추가 10MB 미만
- **C14**: 소스 파일 200줄 미만 권장 (300줄 hard limit 대비 여유 확보)
- **C15**: 최소 터미널 너비 40 columns에서 정상 렌더링 (기존 `bannerWidth = 40` 기준)
- **C16**: `NO_COLOR` 환경변수 설정 시 모든 색상 비활성화

## Terminal Compatibility Constraints

- **C17**: 최소 지원 터미널: Terminal.app, iTerm2, Warp, Ghostty (macOS), GNOME Terminal, Konsole, Alacritty, kitty (Linux)
- **C18**: multiplexer 지원: tmux, screen
- **C19**: IDE 터미널 지원: VS Code integrated terminal, JetBrains terminal
- **C20**: 256-color 미지원 터미널에서 readable fallback 렌더링 제공

## Migration Constraints

- **C21**: `promptChoice()`, `promptYesNo()`는 다른 커맨드에서 미사용 확인 완료 -- 제거 또는 deprecated 마킹 허용
- **C22**: `warnParentRuleConflicts()`는 init.go에서 호출되므로 마이그레이션 필수 -- huh Confirm으로 전환
- **C23**: `isStdinTTY()` 유틸리티는 여러 곳에서 필요하므로 유지 (위치 이동 가능)
- **C24**: `langCodes`, `langLabels` 데이터는 wizard_steps.go로 이동

## Signal Handling Constraints

- **C25**: SIGINT/SIGTERM 수신 시 terminal state 원래대로 복원 (raw mode cleanup)
- **C26**: bubbletea program의 panic은 recover되어 graceful error message로 전환
