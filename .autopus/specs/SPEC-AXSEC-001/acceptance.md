# SPEC-AXSEC-001 수락 기준

## 시나리오

### S1: Shell metacharacter가 포함된 metric command 거부

- Given: `RunMetric`에 `echo ok; rm -rf /` 형태의 cmd가 전달됨
- When: `RunMetric(ctx, "echo ok; rm -rf /")` 호출
- Then: 에러 반환, 커맨드 실행되지 않음, 에러 메시지에 "disallowed shell metacharacter" 포함

### S2: 정상 metric command 실행 허용

- Given: `RunMetric`에 `echo '{"metric": 1.5}'` 형태의 안전한 cmd가 전달됨
- When: `RunMetric(ctx, "echo '{\"metric\": 1.5}'")` 호출
- Then: 정상 실행, MetricOutput 반환, 기존 동작과 동일

### S3: Pipe 문자 포함 metric command 거부

- Given: `RunMetric`에 `cat file | grep metric` 형태의 cmd가 전달됨
- When: `RunMetric(ctx, "cat file | grep metric")` 호출
- Then: 에러 반환, 커맨드 실행되지 않음

### S4: AllowShellMeta 옵션으로 bypass

- Given: 호출자가 `AllowShellMeta` 옵션을 명시적으로 전달
- When: metacharacter가 포함된 cmd로 호출
- Then: 검증 bypass, 커맨드 정상 실행

### S5: 악의적 branch name 거부

- Given: `addWorktreeWithRetry`에 `main; rm -rf /` 형태의 branch name이 전달됨
- When: `Create(ctx, "main; rm -rf /")` 호출
- Then: 에러 반환, git 명령 실행되지 않음

### S6: 정상 branch name 허용

- Given: `Create`에 `feature/my-branch` 형태의 유효한 branch name이 전달됨
- When: `Create(ctx, "feature/my-branch")` 호출
- Then: 정상 실행, worktree 생성됨

### S7: sanitizeBranchName 에러 반환

- Given: `sanitizeBranchName`에 `..exploit` 형태의 이름 전달
- When: `sanitizeBranchName("..exploit")` 호출
- Then: 에러 반환, 빈 문자열

### S8: 빈 문자열 branch name 처리

- Given: `Create`에 빈 문자열 branch name 전달
- When: `Create(ctx, "")` 호출
- Then: detach 모드로 정상 동작 (기존 동작 유지)

### S9: 기존 테스트 전체 통과

- Given: 모든 변경 사항이 적용된 코드베이스
- When: `go test ./...` 실행
- Then: 기존 테스트 포함 전체 통과, 실패 0건

### S10: 255자 초과 branch name 거부

- Given: 256자 이상의 branch name 전달
- When: `sanitizeBranchName` 호출
- Then: 에러 반환, "branch name too long" 메시지 포함
