# SPEC-AXQUAL-001 구현 계획

## 태스크 목록

### P0 — resolvePlatform 단위 테스트

- [ ] T1: `pipeline_run_test.go`에 `TestResolvePlatform` table-driven 테스트 함수 작성
  - Subtask: PATH를 격리하기 위해 `t.Setenv("PATH", ...)` 사용
  - Subtask: 임시 디렉토리에 더미 바이너리 생성하여 `exec.LookPath` 동작 검증
  - Cases: explicit platform, binary found (claude), binary found (codex only), binary found (gemini only), multiple binaries (priority), empty PATH (fallback)
- [ ] T2: `pipeline_run.go:101`의 `@AX:TODO` 태그 제거
- [ ] T3: 테스트 실행 및 전체 통과 확인 (`go test ./internal/cli/ -run TestResolvePlatform -v`)

### P1 — 템플릿 TODO 문서화

- [ ] T4: `agent_create.go`의 `agentTemplate` const 위 주석에 `@AX:EXCLUDE` 추가
  - 이유: 사용자 대상 placeholder이며 코드 결함이 아님
- [ ] T5: `skill_create.go`의 `skillTemplate` const 위 주석에 `@AX:EXCLUDE` 추가
  - 이유: 동일 패턴

## 구현 전략

### resolvePlatform 테스트 접근법

`resolvePlatform`은 `exec.LookPath`를 내부에서 직접 호출하므로, 인터페이스 추상화 없이 **환경 격리** 방식으로 테스트한다.

1. `t.Setenv("PATH", tmpDir)` — Go 1.17+ `t.Setenv`는 테스트 종료 시 자동 복원
2. `tmpDir`에 실행 가능한 더미 파일 생성 (`os.WriteFile` + `os.Chmod 0755`)
3. 각 케이스마다 PATH에 포함할 바이너리를 다르게 구성

이 방식은 함수 시그니처를 변경하지 않으며, 기존 프로덕션 코드 수정 없이 테스트 가능하다.

### 템플릿 TODO 문서화

기존 `@AX:NOTE` 주석이 이미 있으므로 (`agent_create.go:16`, `skill_create.go:16`), 해당 라인에 `@AX:EXCLUDE` 속성을 추가하여 AX 스캔 도구가 무시하도록 한다.

## 변경 범위

- 수정 파일: 4개 (pipeline_run.go, pipeline_run_test.go, agent_create.go, skill_create.go)
- 신규 파일: 0개
- 예상 추가 코드: ~80줄 (테스트 코드)
- 기존 동작 변경: 없음
