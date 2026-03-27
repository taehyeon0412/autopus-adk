# SPEC-TERM-001 수락 기준

## 시나리오

### S1: cmux 환경에서 멀티에이전트 파이프라인 실행
- Given: cmux가 설치되어 있고 PATH에서 접근 가능하다
- When: 사용자가 `auto go SPEC-XXX --multi`를 실행한다
- Then: DetectTerminal()이 cmux adapter를 반환하고, cmux workspace가 생성되며, 각 에이전트가 독립 패인에서 실행된다

### S2: tmux fallback 동작
- Given: cmux가 설치되어 있지 않고, tmux가 설치되어 있다
- When: 사용자가 `auto go SPEC-XXX --multi`를 실행한다
- Then: DetectTerminal()이 tmux adapter를 반환하고, tmux session이 생성되며, split-window로 패인이 분할된다

### S3: plain mode fallback (graceful degradation)
- Given: cmux와 tmux 모두 설치되어 있지 않다
- When: 사용자가 `auto go SPEC-XXX --multi`를 실행한다
- Then: DetectTerminal()이 plain adapter를 반환하고, 경고 메시지가 stderr에 출력되며, 기존 in-process 방식으로 파이프라인이 실행된다

### S4: --multi 미사용 시 기존 동작 유지
- Given: 기존 파이프라인 코드가 정상 동작한다
- When: 사용자가 `auto go SPEC-XXX`를 --multi 없이 실행한다
- Then: Terminal adapter가 활성화되지 않고, 기존 서브에이전트(Agent tool) 방식으로 실행된다

### S5: auto agent run 서브커맨드
- Given: `.autopus/runs/T1/context.yaml`에 유효한 태스크 정보가 있다
- When: 사용자가 `auto agent run T1`을 실행한다
- Then: context.yaml을 읽어 태스크를 실행하고, `.autopus/runs/T1/result.yaml`에 결과를 기록한다

### S6: auto agent run — 태스크 파일 없음
- Given: `.autopus/runs/T99/context.yaml`이 존재하지 않는다
- When: 사용자가 `auto agent run T99`를 실행한다
- Then: 에러 메시지 "task context not found: T99"가 출력되고, exit code 1로 종료된다

### S7: cmux workspace 생성 실패
- Given: cmux가 설치되어 있지만 Socket API가 응답하지 않는다
- When: CreateWorkspace가 호출된다
- Then: 에러가 반환되고, 파이프라인은 plain 모드로 fallback하여 계속 실행된다

### S8: Phase 전환 시 레이아웃 변경 (P1)
- Given: Phase 1(planner, 1패인)이 완료되었다
- When: Phase 2(executor)로 전환된다
- Then: 기존 패인이 정리되고, executor 수만큼 새로운 패인이 생성된다

### S9: 동시 패인 수 초과
- Given: max_panes가 5로 설정되어 있다
- When: 7개의 executor 태스크가 Phase 2에서 실행된다
- Then: 5개가 먼저 패인에서 실행되고, 나머지 2개는 큐에 대기하며, 패인 완료 시 순차 실행된다

### S10: tmux 중첩 세션 방지
- Given: 이미 tmux 세션 안에서 실행 중이다 (TMUX 환경변수 존재)
- When: tmux adapter가 새 세션을 생성하려 한다
- Then: 중첩 세션 대신 현재 세션에서 window/pane을 분할한다

### S11: 패인 프로세스 비정상 종료
- Given: executor 패인에서 `auto agent run T3`가 실행 중이다
- When: 해당 프로세스가 비정상 종료(exit code != 0)한다
- Then: result.yaml에 실패 상태가 기록되고, 대시보드 패인에 실패가 표시되며, 다른 패인은 영향 없이 계속 실행된다

### S12: Terminal interface 단위 테스트
- Given: cmux, tmux, plain adapter가 모두 구현되어 있다
- When: `go test ./pkg/terminal/...`을 실행한다
- Then: 모든 adapter가 Terminal interface를 구현하고, 테스트 커버리지가 85% 이상이다
