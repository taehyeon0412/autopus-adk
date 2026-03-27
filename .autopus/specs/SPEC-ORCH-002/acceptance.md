# SPEC-ORCH-002 수락 기준

## 시나리오

### S1: cmux 환경에서 파이프라인 시작 시 모니터링 pane 생성

- Given: cmux가 설치되어 있고, 사용자가 `/auto go SPEC-XXX`를 실행
- When: 파이프라인이 시작됨
- Then: 2개의 수평 분할 pane이 생성됨 (로그 pane + 대시보드 pane)
- And: 로그 pane에서 `tail -f /tmp/autopus-pipeline-SPEC-XXX.log`가 실행 중
- And: 대시보드 pane에서 초기 상태 (Phase 1: pending)가 표시됨

### S2: cmux 미설치 환경에서 graceful skip

- Given: cmux가 설치되어 있지 않음 (plain 또는 tmux)
- When: 파이프라인이 시작됨
- Then: 모니터링 pane이 생성되지 않음
- And: 파이프라인이 정상적으로 실행됨 (에러 없음)

### S3: 구조화된 로그 기록

- Given: 파이프라인이 실행 중이고, 로그 파일이 존재
- When: Phase 전환 또는 에이전트 스폰 이벤트 발생
- Then: 로그 파일에 `[2026-03-25 14:30:00] [planner] [Phase 1] Task decomposition started` 형식의 엔트리가 추가됨
- And: 로그 pane에서 실시간으로 새 엔트리가 표시됨

### S4: 에이전트 프롬프트에 로그 경로 주입

- Given: 에이전트가 스폰되려 함
- When: 에이전트 프롬프트가 구성됨
- Then: 프롬프트에 `## Pipeline Monitor` 섹션과 로그 파일 경로가 포함됨
- And: 에이전트가 해당 파일에 자체 로그를 기록할 수 있음

### S5: 대시보드 갱신

- Given: 대시보드 pane이 활성 상태
- When: Phase 1 → Phase 2 전환 발생
- Then: 대시보드 pane에 Phase 1: done, Phase 2: running 상태가 반영됨
- And: 경과 시간이 업데이트됨

### S6: 파이프라인 완료 시 정리

- Given: 모니터링 pane 2개가 활성 상태
- When: 파이프라인이 완료됨 (성공 또는 실패)
- Then: 두 pane이 닫힘
- And: `/tmp/autopus-pipeline-SPEC-XXX.log` 파일이 삭제됨

### S7: 로그 기록 실패 시 파이프라인 계속 실행

- Given: 로그 파일 경로에 쓰기 권한이 없음
- When: 로그 기록 시도
- Then: 경고 메시지가 stderr에 출력됨
- And: 파이프라인 실행이 중단되지 않음

### S8: `auto pipeline dashboard` CLI 명령

- Given: 파이프라인 체크포인트 파일이 존재
- When: `auto pipeline dashboard SPEC-XXX` 실행
- Then: Phase별 상태, 활성 에이전트, 경과 시간이 포맷팅되어 stdout에 출력됨

### S9: `--team` 모드에서 동일한 모니터링

- Given: `--team` 플래그로 Agent Teams 모드 활성화
- When: 팀원이 SendMessage로 통신
- Then: 각 팀원의 활동이 동일한 로그 파일에 기록됨
- And: 대시보드에 팀원별 상태가 표시됨

### S10: 대시보드 렌더링 포맷

- Given: Phase 2 실행 중, executor 2개 활성
- When: 대시보드가 렌더링됨
- Then: 다음 형식으로 출력:
  ```
  ╔══════════════════════════════════════╗
  ║  SPEC-XXX Pipeline Dashboard        ║
  ╠══════════════════════════════════════╣
  ║ Phase 1: Planning       ✓ done      ║
  ║ Phase 2: Implementation ▶ running   ║
  ║   executor-1: T1 (pkg/foo.go)       ║
  ║   executor-2: T2 (pkg/bar.go)       ║
  ║ Phase 3: Testing        ○ pending   ║
  ║ Phase 4: Review         ○ pending   ║
  ╠══════════════════════════════════════╣
  ║ Elapsed: 2m 34s                     ║
  ╚══════════════════════════════════════╝
  ```
