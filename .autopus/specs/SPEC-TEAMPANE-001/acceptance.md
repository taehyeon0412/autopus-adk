# SPEC-TEAMPANE-001 수락 기준

## 시나리오

### S1: cmux 터미널에서 3인 팀 패널 생성
- Given: cmux 터미널이 감지되고, 팀 구성이 ["lead", "builder", "guardian"]
- When: `TeamMonitorSession.Start()`가 호출됨
- Then: SplitPane(Vertical)이 3회 호출됨 (lead, builder, guardian)
- Then: 초기 패널이 dashboard로 사용됨
- Then: 각 패널에 `tail -f autopus-team-{specID}-{role}-*.log` 명령이 전송됨
- Then: 대시보드 패널에 팀원 상태가 표시됨

### S2: cmux 터미널에서 4인 팀 패널 생성
- Given: cmux 터미널이 감지되고, 팀 구성이 ["lead", "builder-1", "builder-2", "guardian"]
- When: `TeamMonitorSession.Start()`가 호출됨
- Then: SplitPane(Vertical)이 4회 호출됨
- Then: 5개 패널이 존재 (dashboard + lead + builder-1 + builder-2 + guardian)

### S3: tmux 터미널 폴백
- Given: cmux가 없고 tmux 터미널이 감지됨
- When: `TeamMonitorSession.Start()`가 호출됨
- Then: tmux의 SplitPane을 사용하여 동일한 순차적 분할 구조가 생성됨
- Then: 각 패널에 로그 스트리밍이 활성화됨

### S4: plain 터미널 graceful degradation
- Given: cmux와 tmux 모두 없어 PlainAdapter가 반환됨
- When: `TeamMonitorSession.Start()`가 호출됨
- Then: 패널 생성 없이 nil 에러로 정상 반환됨
- Then: 후속 `UpdateAgent()` 호출이 무시됨 (패닉 없음)
- Then: `Close()` 호출이 에러 없이 반환됨

### S5: 팀원 상태 업데이트
- Given: TeamMonitorSession이 시작되고 Builder 패널이 활성 상태
- When: `UpdateAgent("builder", "Phase 2 - implementing")`가 호출됨
- Then: 대시보드 패널에 Builder 상태가 갱신됨

### S6: 파이프라인 완료 시 정리
- Given: TeamMonitorSession이 4개 패널로 실행 중
- When: `Close()`가 호출됨
- Then: 모든 팀원 패널이 닫힘
- Then: 대시보드 패널이 닫힘
- Then: 모든 임시 로그 파일이 삭제됨

### S7: 팀원 실패 시 패널 표시 (cmux)
- Given: cmux에서 TeamMonitorSession이 실행 중이고 Builder 패널이 활성 상태
- When: Builder 팀원이 실패함
- Then: Builder 패널에 `[FAILED] builder: {error}` 메시지가 전송됨
- Then: Builder 패널은 파이프라인 Close()까지 닫히지 않음
- Then: 다른 팀원 패널은 정상 동작을 계속함

### S7b: 팀원 실패 시 패널 표시 (tmux — 제한)
- Given: tmux에서 TeamMonitorSession이 실행 중
- When: Builder 팀원이 실패함
- Then: Builder 패널에 `[FAILED] builder: {error}` 메시지가 전송됨
- Then: 개별 패널 보존은 미지원 — Close() 시 전체 세션이 함께 정리됨

### S8: 기존 MonitorSession 비간섭
- Given: `--team` 플래그 없이 일반 파이프라인이 실행됨
- When: 기존 `MonitorSession.Start()`가 호출됨
- Then: 기존 2-패널 동작(dashboard + log tail)이 변경 없이 유지됨
- Then: `TeamMonitorSession` 코드가 로드되지만 실행되지 않음

### S9: 패널 생성 실패 시 폴백
- Given: cmux 터미널이 감지되었으나 SplitPane()이 실패함
- When: `TeamMonitorSession.Start()`가 호출됨
- Then: 이미 생성된 패널이 정리됨
- Then: plain 터미널 모드로 폴백하여 에러 없이 반환됨

### S10: 레이아웃 계산 — 3인 팀
- Given: 팀원 목록이 ["lead", "builder", "guardian"]
- When: `planLayout(["lead", "builder", "guardian"])`이 호출됨
- Then: 3개의 Vertical split을 수행하는 LayoutPlan이 반환됨
- Then: role 순서가 ["lead", "builder", "guardian"]

### S11: 레이아웃 계산 — 5인 팀
- Given: 팀원 목록이 ["lead", "builder-1", "builder-2", "builder-3", "guardian"]
- When: `planLayout([...])`이 호출됨
- Then: 5개의 Vertical split을 수행하는 LayoutPlan이 반환됨

### S12: 로그 파일 네이밍 유일성
- Given: specID가 "SPEC-TEAMPANE-001"이고 role이 "builder-1"
- When: 로그 파일이 생성됨
- Then: 파일명이 `autopus-team-SPEC-TEAMPANE-001-builder-1-` 접두사를 가짐
- Then: os.CreateTemp에 의해 유일한 접미사가 추가됨

### S13: compact 대시보드 렌더링
- Given: 패널 폭이 30자로 제한됨
- When: `RenderTeamDashboard(data, 30)`이 호출됨
- Then: compact 모드로 렌더링됨 (boxWidth 조정)
- Then: 텍스트 잘림 없이 정상 표시됨

### S14: PipelineMonitor 인터페이스 호환성
- Given: `TeamMonitorSession`과 `MonitorSession`이 모두 존재
- When: 각각을 `PipelineMonitor` 타입으로 할당
- Then: 컴파일 에러 없이 인터페이스를 만족함
