# SPEC-TERM-001 리서치

## 기존 코드 분석

### 프로젝트 구조
- **모듈**: `github.com/insajin/autopus-adk` (Go 1.26)
- **CLI 프레임워크**: Cobra (`github.com/spf13/cobra v1.9.1`)
- **테스트**: `github.com/stretchr/testify v1.11.1`
- **TUI**: Charmbracelet lipgloss (`github.com/charmbracelet/lipgloss v1.1.0`)
- **진입점**: `cmd/auto/main.go` → `cli.Execute()`

### 관련 파일 및 패턴

#### 바이너리 감지 패턴
- `pkg/detect/detect.go` — `IsInstalled(binary string) bool` (6+ 소비자)
- `DetectPlatforms() []Platform` — PATH에서 코딩 CLI 감지
- **재활용**: cmux/tmux 바이너리 감지에 `IsInstalled()` 직접 사용 가능

#### 외부 프로세스 실행 패턴
- `pkg/orchestra/runner.go` — `runProvider()` 함수가 `exec.CommandContext` 래핑
- `newCommand()` 헬퍼로 테스트 시 mock 가능한 구조
- `PromptViaArgs` 플래그로 stdin vs args 입력 분기
- **참조**: cmux/tmux CLI 래핑 시 동일한 패턴 적용

#### 에이전트 커맨드 등록
- `internal/cli/agent_create.go` — `newAgentCmd()` (line 56-65)
  - 현재 `agent create` 서브커맨드만 등록
  - `agent run` 서브커맨드 추가 위치: `cmd.AddCommand(newAgentRunSubCmd())` (line 63 부근)
- `internal/cli/root.go` — `root.AddCommand(newAgentCmd())` (line 57)

#### 파이프라인 상태 관리
- `pkg/pipeline/types.go` — `Checkpoint` 구조체 (Phase, TaskStatus, GitCommitHash)
- `internal/cli/pipeline.go` — 체크포인트 로드/저장, specCheckpointPath()
- **참조**: `.autopus/runs/<task-id>/` 디렉토리의 context/result 파일 구조 설계에 활용

#### --multi 플래그 현황
- `templates/claude/commands/auto-router.md.tmpl` — `--multi` 플래그 정의 (스킬 레벨)
- 현재 Go 바이너리에는 `--multi` 관련 코드 없음 — 스킬 레벨 프롬프트에서만 사용
- **결정**: Go 바이너리의 pipeline 실행 코드에 `--multi` 플래그를 직접 추가해야 함

#### 워크트리 안전 규칙과의 관계
- `.claude/rules/autopus/worktree-safety.md` — 동시 워크트리 제한 5개
- **동기화**: `max_panes` 기본값을 5로 설정하여 워크트리 제한과 일치

### cmux API 분석

#### 사용 가능한 명령어 (예상)
```bash
cmux workspace create <name>        # 워크스페이스 생성
cmux workspace list                 # 워크스페이스 목록
cmux workspace remove <name>        # 워크스페이스 삭제
cmux pane split --direction h|v     # 패인 분할
cmux pane list                      # 패인 목록
cmux pane focus <id>                # 패인 포커스
cmux notify "<message>"             # 알림 표시
cmux send-keys <pane-id> "<cmd>"    # 특정 패인에 명령 전송
```

#### cmux Socket API 특성
- cmux는 manaflow-ai에서 개발한 터미널 멀티플렉서
- Socket 기반 API 제공 (CLI 명령이 소켓을 통해 통신)
- 사용자가 cmux를 메인 터미널로 사용 중 (MEMORY.md 참조)

#### 주의사항
- cmux API가 변경될 수 있음 → adapter 패턴으로 격리
- cmux 버전 체크 로직 필요 → `cmux --version` 파싱

### tmux API 분석

#### 사용할 명령어
```bash
tmux new-session -d -s <name>                    # 세션 생성 (detached)
tmux split-window -t <session> -h|-v             # 패인 분할
tmux send-keys -t <session>:<pane> "<cmd>" Enter  # 명령 전송
tmux display-message "<msg>"                      # 메시지 표시
tmux kill-session -t <name>                       # 세션 종료
tmux list-panes -t <session> -F "#{pane_id}"     # 패인 목록
```

#### 중첩 세션 처리
- `TMUX` 환경변수가 설정되어 있으면 이미 tmux 세션 안
- 이 경우 `new-session` 대신 `new-window`로 현재 세션에 추가

### 설계 결정

#### D1: Adapter 패턴 선택
- **결정**: 각 터미널 멀티플렉서를 독립 adapter로 구현
- **이유**: cmux API 변경에 강건, 테스트 용이, 새 멀티플렉서 추가 용이
- **대안 검토**: Strategy 패턴 — 인터페이스가 동일하므로 Adapter와 실질적 차이 없음

#### D2: CLI 래핑 vs Socket 직접 통신
- **결정**: v1에서는 `os/exec`로 cmux/tmux CLI 래핑
- **이유**: cmux Socket API의 Go SDK가 아직 없음, CLI 래핑이 가장 안정적
- **대안 검토**: Unix socket 직접 통신 — cmux 내부 프로토콜 의존성 높아 v2로 연기

#### D3: 프로세스 간 통신 방식
- **결정**: 파일 기반 (.autopus/runs/) — context.yaml 입력, result.yaml 출력
- **이유**: 프로세스 독립성 보장, 디버깅 용이, 실패 복구 가능
- **대안 검토**: Unix socket IPC — 구현 복잡도 대비 이점 적음. Named pipe — 양방향 통신 불필요

#### D4: `auto agent run`을 독립 서브커맨드로 분리
- **결정**: `auto agent run <task-id>` 서브커맨드 신규 생성
- **이유**: 각 패인에서 독립 프로세스로 실행해야 하므로, CLI 진입점 필요
- **대안 검토**: 라이브러리 호출 — 패인에서 프로세스 실행 시 CLI가 더 자연스럽고 장애 격리 가능

#### D5: max_panes 제한값
- **결정**: 기본 5개, worktree-safety 규칙과 동기화
- **이유**: 각 패인이 worktree를 사용할 수 있으므로 동일 제한 적용
- **대안 검토**: 무제한 — 리소스 소모 과다, 터미널 가독성 저하

#### D6: pkg/terminal 위치 선택
- **결정**: `pkg/terminal/` (public package)
- **이유**: 향후 다른 도구에서도 재사용 가능, `pkg/detect`와 같은 수준
- **대안 검토**: `internal/terminal/` — 재사용성 제한, 현재 pkg/ 아래에 다른 패키지들도 public으로 노출 중

### 리스크 및 완화

| 리스크 | 영향 | 완화 |
|--------|------|------|
| cmux API 비호환 변경 | adapter 동작 불능 | adapter 패턴 격리 + 버전 체크 |
| tmux 없는 환경 (Docker 등) | visual pipeline 불가 | plain fallback + 경고 |
| 패인 프로세스 좀비화 | 리소스 누수 | timeout + process group kill |
| 동시 exec 과다 | 시스템 리소스 부족 | max_panes 제한 + 큐잉 |
| 파일 기반 IPC race condition | 데이터 손실 | file lock 또는 atomic write |
