# SPEC-TEAMPANE-001 리서치

## 기존 코드 분석

### Terminal Interface (`autopus-adk/pkg/terminal/terminal.go`)
- `Terminal` 인터페이스: `Name()`, `CreateWorkspace()`, `SplitPane()`, `SendCommand()`, `Notify()`, `Close()`
- `Direction` 상수: `Horizontal`(좌우), `Vertical`(상하)
- `PaneID` 타입: string wrapper
- **중요**: `SplitPane(ctx, direction)` — 현재 활성 패널만 분할. 특정 패널 ID 지정 분할 미지원
- **결론**: 인터페이스 변경 없이 순차적 split 전략으로 구현

### DetectTerminal (`autopus-adk/pkg/terminal/detect.go`)
- `DetectTerminal()` — cmux > tmux > plain 우선순위
- `isInstalled` 함수 변수로 테스트 시 mock 가능

### CmuxAdapter (`autopus-adk/pkg/terminal/cmux.go`)
- `SplitPane()`: `cmux new-split {right|down}` 명령 → `surface:N` ref 반환
- `SendCommand()`: `cmux send --surface {ref} {cmd}` — 패널에 명령 전송
- `Close()`: surface ref면 `close-surface`, workspace ref면 `close-workspace`
- **핵심**: 개별 패널(surface) 닫기 지원 → R5 실패 패널 보존 가능

### TmuxAdapter (`autopus-adk/pkg/terminal/tmux.go`)
- `SplitPane()`: `tmux split-window -t {session} {-h|-v}` — 패널 ID 반환
- `SendCommand()`: `tmux send-keys -t {session}:{paneID} {cmd} Enter`
- `Close()`: `tmux kill-session -t {name}` — **세션 단위만 지원**
- **제약**: 개별 패널 닫기 불가 → R5 실패 패널 보존은 cmux에서만 가능

### MonitorSession (`autopus-adk/pkg/pipeline/monitor.go`)
- `isCmux()`: `term.Name() == "cmux"` 체크 — **현재 tmux는 스킵됨**
- `Start()`: Vertical split(log pane) + Horizontal split(dashboard pane) = 2패널
- `Close()`: `cleanupPanes()`(pane 리스트 nil 설정) + 로그 파일 삭제 + `term.Close()`
- `MonitorState`: `Phase string`, `Agents map[string]string`

### Dashboard (`autopus-adk/pkg/pipeline/dashboard.go`)
- `DashboardData`: Phases, Agents, Blocker, Elapsed
- `RenderDashboard()`: box-drawing 문자로 대시보드 렌더링
- `phaseOrder`: 5개 phase (Planning, Test Scaffold, Implementation, Testing, Review)
- `statusIcon()`: done(✓), running(▶), failed(✗), pending(○) 아이콘
- `boxWidth = 38` — 고정 폭 (narrow pane에서 깨질 수 있음)

### Pane Runner (`autopus-adk/pkg/orchestra/pane_runner.go`)
- `splitProviderPanes()`: 프로바이더별 **순차 Horizontal split** + 임시 파일 생성
- `sendPaneCommands()`: 각 패널에 명령 전송, 실패 시 skipWait 표시
- `cleanupPanes()`: context.Background()로 패널 닫기 + 파일 삭제
- **핵심 패턴**: split → send → collect → cleanup (이 패턴을 팀 모드에 재활용)

### Shell Escape (`autopus-adk/pkg/orchestra/pane_shell.go`)
- `shellEscapeArg()`, `shellEscapeArgs()`: 쉘 인젝션 방지
- `sanitizeProviderName()`: 경로 탐색 방지
- `uniqueHeredocDelimiter()`: heredoc 충돌 방지
- **패키지 경계**: 모두 unexported → `pkg/pipeline/`에서 직접 호출 불가
- **결정**: 필요한 함수를 `pkg/pipeline/team_pane.go`에 자체 구현

### Events (`autopus-adk/pkg/pipeline/events.go`)
- `EventAgentSpawn`, `EventAgentDone`: 에이전트 생명주기 이벤트
- `Event.Agent` 필드로 팀원 식별 가능
- **연동 포인트**: 이벤트 발생 시 `TeamMonitorSession.UpdateTeammate()` 호출

## 설계 결정

### D1: 순차적 Split 전략 (그리드 → 스택)
- **결정**: 2x2 그리드 대신 순차 Horizontal split으로 수직 스택 레이아웃 채택
- **이유**: Terminal.SplitPane(ctx, direction)이 현재 활성 패널만 분할하며, 특정 패널 ID 지정 분할을 지원하지 않음. 그리드 구현을 위해서는 Terminal 인터페이스 변경이 필요하나 이는 @AX:ANCHOR로 표시된 안정적 경계를 침범함
- **이점**: pane_runner.go와 동일한 검증된 패턴 재활용, Terminal 인터페이스 변경 불필요
- **대안 기각**: Terminal 인터페이스에 SplitPaneAt(ctx, paneID, direction) 추가 → 모든 어댑터 수정 필요, 과도한 변경

### D2: isMultiplexer 일반화
- **결정**: `term.Name() != "plain"` 체크로 cmux와 tmux 모두 지원
- **이유**: 기존 MonitorSession은 cmux만 지원하지만, tmux도 SplitPane/SendCommand를 구현하므로 배제할 이유 없음

### D3: 고정 레이아웃 (동적 패널 변경 삭제)
- **결정**: 초기 팀 구성 기반 고정 레이아웃. 동적 패널 추가/제거 미지원
- **이유**: 터미널 멀티플렉서에서 mid-session 패널 추가/삭제는 기존 레이아웃을 붕괴시킬 위험이 있음. late-spawned 팀원은 Lead 패널에 로그 공유로 충분
- **리뷰 피드백 반영**: gemini "Dynamic Pane Management Risks" — 동적 변경은 예측 불가능한 레이아웃을 야기

### D4: 패키지 경계 존중
- **결정**: `pkg/pipeline/`에 새 파일 배치, `pkg/orchestra/`의 unexported 함수 사용하지 않음
- **이유**: shell-escape 등 필요한 유틸은 소량(~20줄)이므로 자체 구현이 적절. cross-package 의존성 추가보다 중복 코드가 낫다
- **리뷰 피드백 반영**: claude "SPEC 생성 파일 경로가 잘못됨"

### D5: PipelineMonitor 인터페이스
- **결정**: `PipelineMonitor` 인터페이스를 `team_monitor.go`에 정의하여 MonitorSession과 TeamMonitorSession 공통화
- **이유**: 파이프라인 코드에서 `--team` 여부에 따라 분기하지 않고 동일한 인터페이스로 주입 가능
- **리뷰 피드백 반영**: claude "기존 MonitorSession과의 인터페이스 관계 불명확"
- **기존 코드 영향**: MonitorSession에 인터페이스 적합성만 확인 (메서드 시그니처 이미 호환)

### D6: tmux 개별 패널 정리 — known limitation
- **결정**: tmux에서는 R5(실패 패널 보존)를 미지원하고 문서화
- **이유**: TmuxAdapter.Close()가 세션 단위만 지원. kill-pane 명령을 추가하려면 TmuxAdapter 수정이 필요하며 이는 R6(기존 코드 무변경) 원칙에 위배
- **완화**: 실패 메시지를 패널에 echo하여 Close() 전까지는 확인 가능
- **리뷰 피드백 반영**: claude+gemini "tmux 개별 패널 정리 vs R5 충돌"

### D7: Dashboard 폭 인식 렌더링
- **결정**: `RenderTeamDashboard(data, maxWidth)`에 maxWidth 파라미터 추가
- **이유**: 순차 스택 레이아웃에서 패널 폭은 터미널 전체 폭이므로 일반적으로 문제없지만, 안전 장치로 compact 모드 지원
- **리뷰 피드백 반영**: gemini "Hardcoded Dashboard Width"

### D8: 로그 파일 네이밍 컨벤션
- **결정**: `os.CreateTemp("", "autopus-team-{specID}-{role}-")` 패턴
- **이유**: specID + role 조합으로 식별 가능하고, os.CreateTemp가 유일성 보장
- **리뷰 피드백 반영**: gemini "Log File Naming and Collision"
