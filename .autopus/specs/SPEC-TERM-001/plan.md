# SPEC-TERM-001 구현 계획

## 태스크 목록

### Phase 1: Core Interface & Types (P0)
- [x] T1: `pkg/terminal/terminal.go` — Terminal interface, Direction, PaneID 타입 정의
- [x] T2: `pkg/terminal/detect.go` — DetectTerminal() 구현 (cmux > tmux > plain 우선순위)
- [x] T3: `pkg/terminal/plain.go` — Plain no-op adapter 구현

### Phase 2: Adapters (P0)
- [x] T4: `pkg/terminal/cmux.go` — cmux Socket API adapter 구현
- [x] T5: `pkg/terminal/tmux.go` — tmux CLI adapter 구현

### Phase 3: Agent Run Subcommand (P0)
- [x] T6: `internal/cli/agent_run.go` — `auto agent run <task-id>` 서브커맨드 구현
- [x] T7: `internal/cli/agent_create.go` 수정 — newAgentCmd()에 run 서브커맨드 등록

### Phase 4: Pipeline Integration (P0)
- [x] T8: Pipeline 실행 코드에 Terminal adapter 연동 — --multi 시 패인 분할 + agent run 실행

### Phase 5: Phase Layout & Dashboard (P1)
- [x] T9: Phase 전환 시 레이아웃 동적 변경 로직
- [x] T10: Dashboard pane 구현 — 파이프라인 진행 상황 실시간 표시

### Phase 6: Tests
- [x] T11: `pkg/terminal/terminal_test.go` — interface 준수 테스트
- [x] T12: `pkg/terminal/detect_test.go` — DetectTerminal 단위 테스트
- [x] T13: `pkg/terminal/cmux_test.go` — cmux adapter 단위 테스트 (mock exec)
- [x] T14: `pkg/terminal/tmux_test.go` — tmux adapter 단위 테스트 (mock exec)
- [x] T15: `pkg/terminal/plain_test.go` — plain adapter 단위 테스트
- [x] T16: `internal/cli/agent_run_test.go` — agent run 서브커맨드 테스트

## 구현 전략

### 기존 코드 활용
- `pkg/detect.IsInstalled()` — cmux/tmux 바이너리 존재 여부 확인에 재활용
- `pkg/orchestra/runner.go`의 `newCommand()` 패턴 — 외부 프로세스 실행 래핑 패턴 참조
- `internal/cli/agent_create.go`의 `newAgentCmd()` — `run` 서브커맨드를 여기에 등록
- `pkg/pipeline/types.go`의 Checkpoint 구조 — task-id 기반 상태 관리에 참조

### 변경 범위
- **신규 파일**: 6개 소스 + 6개 테스트 = 12개
- **수정 파일**: 2-3개 (agent_create.go, pipeline 통합 코드)
- **기존 동작 영향 없음**: `--multi` 플래그가 없으면 기존 동작 그대로 유지

### 접근 방법
1. Interface-first 설계: Terminal interface를 먼저 확정하고, 각 adapter를 독립적으로 구현
2. Adapter 패턴: 각 adapter는 Terminal interface를 구현하므로 교체 가능
3. Command 래핑: cmux/tmux 모두 CLI 래핑 방식 (os/exec)으로 구현하여 외부 의존성 최소화
4. 테스트 전략: exec.Command를 mockable하게 주입 (orchestra.newCommand 패턴 참조)
5. 파일 크기 제한: 각 파일 200줄 이하 목표, 300줄 초과 금지

### Task 간 독립성
- T1-T3: 독립 실행 가능 (T2는 T1의 타입에 의존)
- T4-T5: T1 완료 후 병렬 실행 가능
- T6-T7: T1 완료 후 독립 실행 가능
- T8: T1-T7 완료 후 실행
- T9-T10: T8 완료 후 실행
- T11-T16: 각 구현 태스크와 병렬 또는 후속 실행 가능

### 파일 기반 프로세스 간 통신
```
.autopus/runs/<task-id>/
├── context.yaml   — 입력: task 설명, SPEC ID, phase, 관련 파일 목록
└── result.yaml    — 출력: 상태(success/fail), 변경 파일, 에러 메시지
```

### 동시 패인 수 제한
- 기본값: `max_panes = 5` (worktree-safety 규칙과 동기화)
- 초과 시 큐잉하여 패인 완료 후 순차 실행
