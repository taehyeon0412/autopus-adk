# SPEC-ORCH-003 수락 기준

## 시나리오

### S1: Auto-Detach 동작 (cmux/tmux 터미널)
- Given: cmux 또는 tmux 터미널이 감지된 환경
- When: `auto orchestra brainstorm "feature"` 실행
- Then: pane N개가 생성되고, 각 pane에 프로바이더 명령이 전송되고, `/tmp/autopus-orch-{8hex}/job.json`이 생성되고, job ID가 stdout에 출력되고, 명령이 2초 이내에 반환된다

### S2: --no-detach 강제 블로킹
- Given: cmux 터미널이 감지된 환경
- When: `auto orchestra brainstorm --no-detach "feature"` 실행
- Then: 기존 RunPaneOrchestra와 동일하게 블로킹 실행되며, 결과가 직접 stdout에 출력된다

### S3: Plain 터미널 하위 호환
- Given: plain 터미널 환경 (cmux/tmux 없음)
- When: `auto orchestra brainstorm "feature"` 실행
- Then: 기존 RunOrchestra 블로킹 동작이 100% 유지되며 detach 관련 메시지가 나타나지 않는다

### S4: Status 조회
- Given: detach 모드로 생성된 job (일부 프로바이더 완료)
- When: `auto orchestra status {jobID}` 실행
- Then: 각 프로바이더의 완료/미완료 상태와 전체 상태(running/partial/done/timeout)가 표시된다

### S5: Wait 명령
- Given: detach 모드로 생성된 job (프로바이더 실행중)
- When: `auto orchestra wait {jobID}` 실행
- Then: sentinel 파일을 polling하여 모든 프로바이더가 완료되거나 timeout에 도달하면 최종 상태를 반환한다

### S6: Result 수집 및 Merge
- Given: 모든 프로바이더가 완료된 job
- When: `auto orchestra result {jobID}` 실행
- Then: 각 프로바이더의 출력 파일을 읽고, job.json의 strategy에 따라 merge하여 결과를 stdout에 출력한다

### S7: Result --cleanup
- Given: 완료된 job
- When: `auto orchestra result {jobID} --cleanup` 실행
- Then: 결과 출력 후 모든 pane이 종료되고, `/tmp/autopus-orch-{jobID}/` 디렉토리가 삭제된다

### S8: Timeout 처리
- Given: detach 모드 job이 timeout_at을 초과
- When: `auto orchestra status {jobID}` 실행
- Then: status가 "timeout"으로 표시되고, 완료된 프로바이더의 부분 결과는 보존된다

### S9: Job 디렉토리 무결성
- Given: detach 모드로 job 생성
- When: job.json을 파싱
- Then: id, strategy, prompt, providers(name, pane_id, output_file), created_at, timeout_at, terminal 필드가 모두 존재한다

### S10: 모든 서브커맨드 적용
- Given: cmux 터미널 환경
- When: `auto orchestra review`, `auto orchestra plan`, `auto orchestra secure` 각각 실행
- Then: brainstorm과 동일하게 auto-detach가 적용된다

### S11: 존재하지 않는 Job ID
- Given: 존재하지 않는 job ID
- When: `auto orchestra status {invalidID}` 실행
- Then: "job not found" 에러 메시지와 함께 비정상 종료 코드를 반환한다
