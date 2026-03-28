# SPEC-PERM-001 구현 계획

## 태스크 목록

### Part 1: Go 바이너리 (P0)

- [ ] T1: `pkg/detect/permission.go` — `DetectPermissionMode()` 함수 구현
  - 파일 소유: `pkg/detect/permission.go` (신규, ~80줄)
  - 프로세스 트리 순회, 플래그 검색, fail-safe 반환
  - 환경변수 `AUTOPUS_PERMISSION_MODE` 오버라이드 지원 (P2 R7)
  - macOS `ps -o args= -p {PID}` 사용, 부모 PID 0 또는 1 도달 시 종료

- [ ] T2: `pkg/detect/permission_test.go` — 단위 테스트
  - 파일 소유: `pkg/detect/permission_test.go` (신규, ~100줄)
  - 케이스: 환경변수 오버라이드, 프로세스 검사 실패 시 safe 반환, 정상 감지
  - 내부 함수 테스트를 위해 `walkProcessTree` 등을 주입 가능하게 설계

- [ ] T3: `internal/cli/permission.go` — Cobra 서브커맨드
  - 파일 소유: `internal/cli/permission.go` (신규, ~60줄)
  - `auto permission detect` → stdout에 "bypass" 또는 "safe" 출력
  - `--json` 플래그로 JSON 출력 (P1 R5)
  - `newPermissionCmd()` 함수 export

- [ ] T4: `internal/cli/root.go` — 커맨드 등록
  - 파일 소유: `internal/cli/root.go` (수정, +1줄)
  - `root.AddCommand(newPermissionCmd())` 추가

### Part 2: Skill 업데이트 (P0)

- [ ] T5: `content/skills/agent-pipeline.md` — 동적 권한 모드 섹션 추가
  - 파일 소유: `content/skills/agent-pipeline.md` (수정)
  - Pipeline Overview 직전에 "Permission Mode Detection" 섹션 추가
  - Phase 0: `auto permission detect` 실행 → PERMISSION_MODE 변수 설정
  - Agent() 호출 시 mode 결정 규칙 문서화:
    - PERMISSION_MODE="bypass" → 모든 에이전트 bypassPermissions
    - PERMISSION_MODE="safe" → 기존 mode 유지

### Part 3: Router 템플릿 업데이트 (P0)

- [ ] T6: `templates/claude/commands/auto-router.md.tmpl` — Step 0.5 추가 및 조건부 mode
  - 파일 소유: `templates/claude/commands/auto-router.md.tmpl` (수정)
  - Route A 시작 전 Step 0.5 "Detect Permission Mode" 삽입
  - `auto permission detect` 실행 결과를 PERMISSION_MODE에 저장
  - 모든 `mode = "plan"` 인스턴스를 조건부로 변경:
    ```
    mode = PERMISSION_MODE == "bypass" ? "bypassPermissions" : "plan"
    ```
  - Pipeline Overview ASCII 다이어그램 업데이트

## 구현 전략

### 기존 코드 활용

- `pkg/detect/detect.go`의 패턴을 따름: 같은 패키지, 같은 네이밍 컨벤션
- `internal/cli/` 기존 커맨드 구조 (newXxxCmd 패턴) 준수
- `os/exec` + `ps` 조합은 기존 `detectBinary()` 패턴과 유사

### 변경 범위 최소화

- Go 코드 변경은 3개 신규 파일 + 1줄 수정으로 한정
- 템플릿/스킬 변경은 기존 구조를 유지하며 조건부 로직만 추가
- 기존 "safe" 모드 동작은 100% 동일하게 보존 (하위 호환성)

### 의존성 순서

```
T1 → T2 (테스트는 구현 후)
T1 → T3 (CLI는 라이브러리 함수 사용)
T3 → T4 (등록은 커맨드 구현 후)
T5, T6은 T1-T4와 독립적으로 병렬 가능
```

## 예상 소요

- Part 1 (Go 코드): T1~T4, 약 240줄 신규 코드
- Part 2 (Skill): T5, 약 30줄 추가
- Part 3 (Template): T6, 약 20줄 수정
