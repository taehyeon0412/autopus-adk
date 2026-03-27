# SPEC-ORCH-003 구현 계획

## 태스크 목록

- [ ] T1: `pkg/orchestra/job.go` — Job struct 및 persistence 함수 구현
  - Job struct 정의 (ID, Strategy, Prompt, Judge, Providers, CreatedAt, TimeoutAt, Terminal)
  - ProviderJob sub-struct (Name, PaneID, OutputFile)
  - SaveJob(): job.json을 temp 디렉토리에 직렬화
  - LoadJob(jobID): /tmp/autopus-orch-{jobID}/job.json 로드
  - CheckStatus(): sentinel 기반 상태 판정 (running/partial/done/timeout/error)
  - CollectResults(): 출력 파일 읽기 + 전략별 merge 위임
  - Cleanup(): pane 종료 + temp 디렉토리 삭제

- [ ] T2: `pkg/orchestra/job_test.go` — Job lifecycle 테스트
  - TestSaveLoadJob: 직렬화/역직렬화 왕복
  - TestCheckStatus: 각 상태값 시나리오
  - TestCollectResults: sentinel 파싱 및 merge 호출
  - TestCleanup: 파일/디렉토리 삭제 확인

- [ ] T3: `pkg/orchestra/pane_runner.go` 수정 — RunPaneOrchestraDetached() 추가
  - splitProviderPanes() 재사용
  - sendPaneCommands() 재사용
  - cleanupPanes() 호출하지 않고 Job으로 pane 정보 저장
  - SaveJob() 호출 후 즉시 반환

- [ ] T4: `internal/cli/orchestra_job.go` — CLI 서브커맨드 구현
  - newOrchestraStatusCmd(): job 상태 조회
  - newOrchestraWaitCmd(): sentinel polling + 완료 대기
  - newOrchestraResultCmd(): 결과 수집 + merge + 출력 (--cleanup 옵션)

- [ ] T5: `internal/cli/orchestra.go` 수정 — auto-detach 분기
  - runOrchestraCommand()에 terminal 타입 체크 추가
  - pane terminal + --no-detach 없음 → RunPaneOrchestraDetached() 호출
  - job ID 출력 후 즉시 반환
  - 서브커맨드 등록: status, wait, result

- [ ] T6: `internal/cli/orchestra_job_test.go` — CLI 테스트
  - TestStatusCmd: mock job으로 상태 출력 검증
  - TestResultCmd: mock 출력 파일로 merge 결과 검증

## 구현 전략

### 기존 코드 활용
- `splitProviderPanes()`, `sendPaneCommands()` — 그대로 재사용 (pane 생성 + 명령 전송)
- `waitForSentinel()`, `hasSentinel()`, `readOutputFile()` — CheckStatus/CollectResults에서 재사용
- `mergeByStrategy()` — CollectResults에서 전략별 병합에 재사용
- `sanitizeProviderName()`, `shellEscapeArg()` — 기존 보안 유틸 재사용
- `randomHex()` — job ID 생성에 재사용

### 변경 범위
- `pane_runner.go`: 약 30줄 추가 (RunPaneOrchestraDetached 함수)
- `orchestra.go` (CLI): 약 20줄 수정 (분기 + 서브커맨드 등록)
- 나머지는 모두 새 파일

### 의존 관계
- T1 → T2 (Job struct가 먼저 있어야 테스트 가능)
- T1 → T3 (RunPaneOrchestraDetached가 Job을 사용)
- T1, T3 → T4 (CLI가 Job과 detached runner를 호출)
- T4 → T5 (서브커맨드가 있어야 등록 가능)
- T4 → T6 (CLI 테스트)

### 병렬 실행 가능 태스크
- T1 + T4 초기 구조 (interface만)
- T2 + T6 (테스트는 각각 독립)
