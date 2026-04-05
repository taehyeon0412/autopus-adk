# SPEC-AXQUAL-001 리서치

## 기존 코드 분석

### resolvePlatform 함수

- **파일**: `internal/cli/pipeline_run.go:85-98`
- **시그니처**: `func resolvePlatform(platform string) string`
- **동작**: 
  1. `platform`이 비어있지 않으면 그대로 반환
  2. `claude`, `codex`, `gemini` 순서로 `exec.LookPath` 호출
  3. 모두 없으면 `"claude"` 반환
- **의존성**: `os/exec.LookPath` (PATH 환경변수 의존)
- **호출처**: `runPipeline` 함수 (pipeline_run.go:104)

### 기존 테스트

- **파일**: `internal/cli/pipeline_run_test.go` (117줄)
- **커버리지**: `TestPipelineRunCmd_DefaultPlatform`이 `resolvePlatform(cfg.Platform)`을 호출하지만, 
  PATH를 제어하지 않아 호스트 환경에 따라 결과가 달라짐
- **부재**: empty PATH, 특정 바이너리만 있는 경우, 우선순위 검증 없음

### 템플릿 TODO

- **agent_create.go:32,36,40**: `agentTemplate` const 내부 한국어 TODO 3개
  - "TODO: 이 에이전트의 역할과 책임을 정의하세요."
  - "TODO: 작업 처리 지침을 작성하세요."
  - "TODO: 완료 기준을 정의하세요."
  - 기존 주석: `@AX:NOTE [AUTO] @AX:REASON: agent template format must match .claude/agents/ conventions` (line 16)
- **skill_create.go:38**: `skillTemplate` const 내부 한국어 TODO 1개
  - "TODO: 이 스킬의 구현 지침을 작성하세요."
  - 기존 주석: `@AX:NOTE [AUTO] @AX:REASON: skill template format must match .claude/skills/ conventions` (line 16)

## 설계 결정

### D1: 환경 격리 vs 인터페이스 모킹

**선택**: 환경 격리 (`t.Setenv`)

**이유**: `resolvePlatform`은 간단한 유틸리티 함수로 `exec.LookPath`를 직접 호출한다. 
인터페이스를 주입하도록 리팩토링하면 프로덕션 코드 변경이 필요하고, 이 함수의 복잡도 대비 과도한 추상화가 된다.

**대안 검토**:
1. ~~인터페이스 모킹~~ — `LookPathFunc` 변수 주입: 코드 변경 범위가 크고 overkill
2. ~~테스트 스킵~~ — `exec.LookPath` 사용 함수는 CI에서도 테스트 가능해야 함
3. **환경 격리** (선택) — `t.Setenv("PATH", tmpDir)` + 더미 바이너리로 실제 LookPath 동작 검증. 
   Go 1.17+ `t.Setenv`는 테스트 종료 시 자동 복원하여 격리 보장.

### D2: 템플릿 TODO 처리 방식

**선택**: `@AX:EXCLUDE` 주석 추가

**이유**: 템플릿 내 TODO는 사용자가 생성한 에이전트/스킬 파일에서 편집할 placeholder다.
이를 제거하면 사용자 경험이 저하된다. AX 도구가 무시하도록 태그만 추가한다.

### D3: 테스트 파일 위치

**선택**: 기존 `pipeline_run_test.go`에 추가

**이유**: 파일이 현재 117줄로, 80줄 추가해도 ~197줄로 200줄 미만 목표 범위 내.
300줄 하드 리밋에도 여유가 있어 별도 파일 분리 불필요.

## 기술 참고

- `t.Setenv`: Go 1.17+, `t.Cleanup`으로 자동 복원. `t.Parallel()`과 함께 사용 불가 (Go에서 panic 발생) — 해당 테스트는 sequential로 실행해야 함.
- `exec.LookPath`: PATH 환경변수에서 실행 파일을 탐색. Unix에서는 실행 권한(`0755`)이 있어야 찾음.
- 더미 바이너리 생성: `os.WriteFile(path, []byte("#!/bin/sh\n"), 0755)` — 실행 가능한 최소 파일.
