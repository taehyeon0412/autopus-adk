# SPEC-ORCH-002 리서치

## 기존 코드 분석

### Terminal 인터페이스 (`pkg/terminal/terminal.go:22-35`)

`Terminal` 인터페이스는 5개 메서드를 제공:
- `SplitPane(ctx, direction) (PaneID, error)` — 모니터링 pane 2개 생성에 사용
- `SendCommand(ctx, paneID, cmd)` — `tail -f` 실행 및 대시보드 갱신에 사용
- `Close(ctx, name)` — pane 정리에 사용
- `Name() string` — cmux 감지 확인
- `Notify(ctx, message)` — 상태 알림에 활용 가능

### CmuxAdapter (`pkg/terminal/cmux.go`)

`CmuxAdapter`가 `Terminal` 인터페이스를 구현. `SplitPane`은 `cmux pane split --direction h` 실행 후 pane ID 반환. `SendCommand`는 `cmux send-keys` 사용.

### DetectTerminal (`pkg/terminal/detect.go:15-23`)

`DetectTerminal()` → cmux > tmux > plain 우선순위. 반환된 `Terminal.Name() == "cmux"` 이면 모니터링 활성화.

### pane_runner.go (`pkg/orchestra/pane_runner.go`)

SPEC-ORCH-001에서 구현된 pane 기반 orchestra 실행기. 핵심 참고 패턴:
- `splitProviderPanes()` (:77-99) — pane 생성 + 임시 파일 할당 패턴
- `cleanupPanes()` (:273-279) — pane 닫기 + 임시 파일 삭제 패턴
- `buildPaneCommand()` (:185-202) — shell escape를 포함한 pane 명령 구성

### Pipeline 체크포인트 (`internal/cli/pipeline.go`)

- `specCheckpointPath(specID)` (:22-24) — `.autopus/pipeline-state/{specID}.yaml` 경로
- `LoadCheckpointIfContinue()` (:29-65) — 체크포인트 로드 + stale 판별
- `pipeline.Checkpoint` 구조체 — Phase 상태, 태스크 완료율 등 포함

### Agent Pipeline 스킬 (`.claude/skills/autopus/agent-pipeline.md`)

5-Phase 파이프라인 정의. 에이전트 스폰 시 `Agent()` 호출의 prompt 파라미터에 컨텍스트를 주입하는 패턴이 이미 존재 (Phase 2 Profile Injection, :166-198). 동일 방식으로 로그 경로를 주입하면 됨.

### Agent Teams 스킬 (`.claude/skills/autopus/agent-teams.md`)

`--team` 모드의 Lead/Builder/Guardian 역할 정의. `SendMessage`로 통신. Builder-Guardian 직접 통신 패턴 (:128-153)에서 로그 기록 패턴을 유사하게 적용 가능.

## 설계 결정

### D1: Go 코드 변경 최소화 — 프롬프트 주입 방식 선택

**결정**: 에이전트 활동 로그는 에이전트 프롬프트에 로그 파일 경로를 주입하여, 에이전트가 직접 `echo` 또는 `Bash`로 기록하게 한다. Go 코드에서는 Phase 전환 등 핵심 이벤트만 기록.

**이유**: Agent Teams/서브에이전트는 Claude Code 내부 API로 동작하므로, Go 코드에서 에이전트 내부 활동을 직접 캡처할 수 없다. 프롬프트 주입이 가장 자연스러운 통합 방식.

**대안 검토**:
- (A) Go 코드에서 에이전트 표준 출력 파싱 → Agent tool은 stdout을 반환하지만 실시간 스트리밍은 불가. 대시보드 갱신에 부적합.
- (B) 별도 IPC 채널 → 과도한 복잡성. 에이전트가 셸 명령을 실행할 수 있으므로 파일 기반이 충분.

### D2: 대시보드를 CLI 명령으로 구현

**결정**: `auto pipeline dashboard {spec-id}` CLI 명령이 체크포인트 YAML을 읽어 상태를 렌더링. 메인 세션이 `SendCommand`로 pane에 재실행.

**이유**: Go 바이너리로 렌더링하면 상태를 정확히 반영할 수 있고, 에이전트 외부에서도 독립적으로 실행 가능. 체크포인트 파일은 이미 존재하므로 추가 상태 관리 불필요.

**대안 검토**:
- (A) watch + shell script 기반 대시보드 → 체크포인트 YAML 파싱이 shell에서 복잡
- (B) 웹 대시보드 → 과도한 범위. cmux pane에서의 텍스트 UI가 적절

### D3: 단일 로그 파일 — append-only

**결정**: 모든 에이전트가 `/tmp/autopus-pipeline-{spec-id}.log`에 append. 파일 잠금 없음.

**이유**: POSIX에서 작은 크기의 `write()` 호출은 pipe/regular file에서 atomic. 에이전트가 `echo "..." >> file` 하는 수준에서는 interleaving 위험 없음. 로그는 디버깅 보조 목적이므로 완벽한 ordering 불필요.

### D4: pkg/pipeline 패키지에 배치

**결정**: 모니터링 관련 코드를 `pkg/pipeline/`에 배치. `pkg/orchestra/`는 멀티프로바이더 실행 전용.

**이유**: 이 기능은 Agent Pipeline의 Phase 진행을 모니터링하는 것이므로, orchestra(멀티프로바이더 실행)보다 pipeline 도메인에 더 적합. 단, Terminal 인터페이스는 `pkg/terminal`에서 import.

### D5: cmux 전용 — tmux에서는 미활성화

**결정**: `DetectTerminal().Name() == "cmux"`인 경우에만 모니터링 활성화. tmux에서는 활성화하지 않음.

**이유**: 사용자가 cmux를 메인 터미널로 사용 중 (memory: `user_terminal_cmux.md`). tmux 지원은 필요 시 후속 SPEC에서 추가.

### D6: JSONL 이벤트 로그 + 텍스트 로그 듀얼 기록

**결정**: JSONL(`events.jsonl`)과 텍스트(`pipeline.log`) 두 가지를 동시 기록.

**이유**: JSONL은 구조화된 파싱과 세션 재생에 적합. 텍스트 로그는 `tail -f`로 즉시 확인 가능. codex 브레인스토밍에서 "event-sourced 상태 모델"의 기반으로 JSONL이 필수적이라는 제안을 반영.

### D7: Late-bind pane 패턴 — 상태 모델 우선

**결정**: 상태 모델(MonitorState)을 먼저 생성하고, cmux pane은 상태에 late-bind. pane이 없어도 상태 모델은 독립적으로 동작.

**이유**: codex 제안 — "먼저 pane을 만들고 에이전트를 띄우는 대신, 팀 상태 모델을 먼저 생성하고 그 상태에 맞춰 pane을 late-bind하면 plain/tmux/cmux fallback이 쉬워진다."

## 브레인스토밍 인사이트 (BS-004 + Orchestra Debate)

### codex 핵심 제안
- **Event-sourced TeamViz core**: `TeamStateStore + EventBus + Renderer` 아키텍처. 모든 상태 변경은 append-only event로 직렬화
- **Role wrapper + sidecar logs**: 각 pane에서 agent를 직접 띄우지 말고 wrapper가 `stdout.log`, `events.jsonl`, sentinel 관리
- **Main pane Control Tower**: phase, blockers, last message, role status, ETA 주기 렌더링
- **Replay artifacts**: 세션 종료 후 동일 이벤트 로그를 재생하거나 첨부 가능한 artifact로 보존

### gemini 핵심 제안
- **Role-Themed Quadrant Layout**: 2x2 grid (Lead=Cyan, Builder=Green, Tester=Yellow, Guardian=Red) — ICE 9.0
- **Active Agent Auto-Focus**: 현재 활성 에이전트 pane을 동적 확대
- **Intercept "Red Alert" Mode**: Guardian 검증 실패 시 main pane을 점유하여 알림

### 후속 구현 후보 (이번 SPEC 범위 밖)
- `auto pipeline replay {spec-id}` — JSONL 이벤트 재생 명령
- bubbletea TUI 기반 인터랙티브 대시보드
- Guardian "Red Alert" 모드
- 역할별 pane 동적 크기 조정
