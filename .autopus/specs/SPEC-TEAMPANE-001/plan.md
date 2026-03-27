# SPEC-TEAMPANE-001 구현 계획

## 태스크 목록

- [ ] T0: Shell-escape 함수 export (선행 작업)
  - `pkg/orchestra/pane_shell.go`의 `shellEscapeArg()`, `sanitizeProviderName()` 등을 export (첫 글자 대문자)
  - 또는 `pkg/shellutil/` 공통 패키지로 추출
  - 기존 `pane_runner.go` 호출부도 함께 수정 (rename)

- [ ] T1: `PipelineMonitor` 인터페이스 및 `TeammatePaneInfo` 타입 정의
  - `PipelineMonitor` 인터페이스: Start(), UpdateAgent(name, status), Close(), LogPath()
  - 인터페이스를 `pkg/pipeline/monitor.go`에 추가 (기존 파일, 인터페이스 선언만 추가)
  - `TeammatePaneInfo` 구조체: role, PaneID, logPath
  - `team_pane.go`에 배치

- [ ] T2: `LayoutPlan` 및 순차적 분할 구현
  - `LayoutPlan` 구조체 (role 이름 목록 기반)
  - `planLayout(teammates []string)` — dashboard를 초기 패널로, 나머지 Vertical split
  - `applyLayout(ctx, term, plan)` — 순차 SplitPane(Horizontal) 호출, PaneID 수집
  - 단위 테스트: 3명/4명/5명 split 횟수 검증
  - `team_layout.go`에 배치

- [ ] T3: `TeamMonitorSession` 코어 구현
  - `NewTeamMonitorSession(specID, term, teammates)` 생성자
  - `Start()` — planLayout → applyLayout → 각 패널에 tail -f 전송
  - `UpdateAgent(name, status)` — PipelineMonitor 인터페이스 구현, 대시보드 패널 갱신
  - `Close()` — cleanupTeammatePanes + 로그 파일 삭제
  - `isMultiplexer()` — `term.Name() != "plain"` 체크
  - plain 터미널 graceful degradation (Start → no-op, UpdateTeammate → no-op)
  - `team_monitor.go`에 배치

- [ ] T4: `TeamDashboardData` 및 폭 인식 렌더링
  - `TeamDashboardData` 구조체 (기존 DashboardData 확장)
  - `TeammateStatus` 타입 (role, phase, status, icon)
  - `RenderTeamDashboard(data, maxWidth)` — maxWidth < 38이면 compact 모드
  - 기존 `RenderDashboard()` 변경 없음
  - `team_dashboard.go`에 배치

- [ ] T5: CLI 통합 준비 및 이벤트 연동
  - `--team` 감지 시 `TeamMonitorSession` 인스턴스화 분기점 문서화
  - `EventAgentSpawn`/`EventAgentDone` 이벤트와 `UpdateTeammate()` 연동 패턴
  - 기존 `MonitorSession`에 `PipelineMonitor` 인터페이스 적합성 검증

- [ ] T6: 전체 테스트 작성
  - `team_monitor_test.go`: mock Terminal로 Start/UpdateTeammate/Close 테스트
  - `team_layout_test.go`: 3명/4명/5명 split 횟수, PaneID 수집 검증
  - `team_pane_test.go`: 패널 생성/정리, 로그 파일 네이밍 검증
  - `team_dashboard_test.go`: 일반 모드/compact 모드 렌더링 출력 검증
  - plain 터미널 graceful degradation 테스트
  - SplitPane 실패 시 cleanup 및 fallback 테스트

## 구현 전략

### 접근 방법
1. **순차적 Split 전략**: Terminal API가 `SplitPane(ctx, direction)`만 지원하므로 pane_runner.go와 동일하게 Vertical split을 순차 실행. 2x2 그리드는 포기하고 수직 스택 레이아웃 채택.
2. **Composition + Interface**: `PipelineMonitor` 인터페이스를 정의하여 `MonitorSession`과 `TeamMonitorSession`이 동일한 타입으로 사용될 수 있게 함. 기존 코드에 인터페이스만 추가하고 MonitorSession 구현체는 변경하지 않음.
3. **Shell-escape 공유**: `pkg/orchestra/pane_shell.go`의 함수를 export하여 `pkg/pipeline/`에서도 사용 가능하게 함 (T0 선행 작업). 중복 구현 방지.
4. **보안**: shell-escape 패턴은 `pane_shell.go`의 검증된 로직을 동일하게 적용.
5. **고정 레이아웃**: 초기 팀 구성 기반으로 한 번만 생성. 동적 패널 추가/제거 없음 (late-spawned 팀원은 Lead 패널에 로그).

### 변경 범위
- **신규 파일 4개**: team_monitor.go, team_layout.go, team_pane.go, team_dashboard.go
- **신규 테스트 4개**: 각 파일의 _test.go
- **기존 파일 변경 2개**: `monitor.go`에 PipelineMonitor 인터페이스 추가, `pane_shell.go`에 함수 export
- **파일 크기**: 모든 파일 200줄 이하 목표, 300줄 하드리밋 준수

### tmux 제한사항
- 개별 패널 닫기 미지원 → 파이프라인 완료 시 세션 일괄 정리
- 실패한 팀원 패널 보존 불가 → 실패 메시지를 패널에 echo한 후 전체 종료 시 함께 닫힘
- 이 제한사항은 R5와 R7에 명시됨
