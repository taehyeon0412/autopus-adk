# SPEC-PERM-001: auto permission detect 서브커맨드 및 agent-pipeline 동적 권한 상승

**Status**: completed
**Created**: 2026-03-28
**Domain**: PERM
**Module**: autopus-adk

## 목적

현재 agent-pipeline에서 서브에이전트 권한 모드가 하드코딩되어 있다. `planner`, `validator`, `reviewer`, `security-auditor`는 `mode: "plan"`으로, `executor`, `tester`, `annotator`는 `mode: "bypassPermissions"`로 고정된다. 메인 Claude Code 세션이 `--dangerously-skip-permissions`로 실행된 경우에도 `plan` 모드 에이전트들은 매번 도구 사용 승인을 요청하여 자동화를 방해한다.

이 SPEC은 부모 프로세스의 권한 모드를 감지하여, `--dangerously-skip-permissions` 세션에서는 모든 서브에이전트가 `bypassPermissions`로 동작하도록 동적 권한 상속을 구현한다.

## 범위

### In-Scope

- Go 바이너리: `auto permission detect` CLI 서브커맨드
- Go 라이브러리: `pkg/detect/permission.go` 권한 모드 감지 함수
- Skill: `content/skills/agent-pipeline.md` 동적 모드 결정 로직
- Template: `templates/claude/commands/auto-router.md.tmpl` 조건부 mode 파라미터

### Out-of-Scope

- Claude Code 이외 플랫폼(Codex, Gemini CLI)의 권한 감지
- 런타임 권한 변경 (세션 중간에 모드 전환)
- 파인그레인 권한 제어 (에이전트별 개별 권한 매핑)

## 요구사항

### P0 (Must Have)

- **R1**: WHEN `auto permission detect` 커맨드가 실행되면, THE SYSTEM SHALL 부모 프로세스 트리를 검사하여 `--dangerously-skip-permissions` 플래그 존재 여부를 판단하고, "bypass" 또는 "safe"를 stdout에 출력한다.
- **R2**: WHEN 프로세스 트리 검사에 실패하면 (권한 부족, 프로세스 종료 등), THE SYSTEM SHALL 안전한 기본값 "safe"를 반환한다.
- **R3**: WHEN agent-pipeline Phase 0에서 권한 모드가 "bypass"로 감지되면, THE SYSTEM SHALL 모든 서브에이전트의 mode를 "bypassPermissions"로 설정한다.
- **R4**: WHEN 권한 모드가 "safe"이면, THE SYSTEM SHALL 기존 하드코딩된 mode 값을 유지한다 (plan/bypassPermissions 혼합).

### P1 (Should Have)

- **R5**: WHEN `auto permission detect --json` 플래그가 지정되면, THE SYSTEM SHALL JSON 형식 `{"mode": "bypass"|"safe", "parent_pid": N, "flag_found": bool}`으로 결과를 출력한다.
- **R6**: WHILE auto-router.md.tmpl이 파이프라인을 시작할 때, THE SYSTEM SHALL Step 0.5에서 `auto permission detect`를 실행하고 결과를 `PERMISSION_MODE` 변수에 저장한다.

### P2 (Could Have)

- **R7**: WHERE 환경변수 `AUTOPUS_PERMISSION_MODE`가 설정되어 있으면, THE SYSTEM SHALL 프로세스 트리 검사를 건너뛰고 해당 값을 사용한다.

## 아키텍처

### 감지 전략

```
DetectPermissionMode()
  1. 환경변수 AUTOPUS_PERMISSION_MODE 확인 → 있으면 즉시 반환
  2. os.Getppid()로 부모 PID 획득
  3. 프로세스 트리 상향 순회 (macOS: ps -o args= -p {PID})
  4. 각 조상 프로세스의 args에서 "--dangerously-skip-permissions" 검색
  5. 발견 시 "bypass", 미발견 시 "safe" 반환
  6. 오류 발생 시 "safe" 반환 (fail-safe)
```

### 변경 파일 목록

| 파일 | 역할 | 변경 유형 |
|------|------|-----------|
| `pkg/detect/permission.go` | 권한 모드 감지 함수 | 신규 |
| `pkg/detect/permission_test.go` | 단위 테스트 | 신규 |
| `internal/cli/permission.go` | `auto permission detect` Cobra 커맨드 | 신규 |
| `internal/cli/root.go` | `newPermissionCmd()` 등록 | 수정 (1줄) |
| `content/skills/agent-pipeline.md` | Phase 0 권한 감지 섹션, 동적 mode | 수정 |
| `templates/claude/commands/auto-router.md.tmpl` | Step 0.5 + 조건부 mode | 수정 |
