# SPEC-AXSEC-001: @AX:WARN Security Hardening

**Status**: completed
**Created**: 2026-04-05
**Domain**: AXSEC

## 목적

코드베이스 스캔에서 발견된 `@AX:WARN` 태그 2건에 대한 defense-in-depth 보안 강화.
현재 코드는 "호출자가 신뢰된 입력만 전달한다"는 가정에 의존하지만, 컴파일 타임 보장이 없어
설정 오류나 향후 코드 변경 시 injection 벡터가 열릴 수 있다.

## 대상 파일

| 파일 | 위치 | 위험 | 심각도 |
|------|------|------|--------|
| `pkg/experiment/metric.go` | L44-49 | Shell command injection via `sh -c` | P0 |
| `pkg/pipeline/worktree.go` | L82-107 | Git branch name injection | P1 |

## 요구사항

### R1: Metric Command Validation (P0)

WHEN a metric command string is passed to `RunMetric`,
THE SYSTEM SHALL validate the command against a set of disallowed shell metacharacters
(`; | && || $( ) \` { } < > \n`) before execution.

WHEN the command contains any disallowed metacharacter,
THE SYSTEM SHALL return an error without executing the command.

WHEN the caller explicitly opts into unsafe execution (e.g., via `AllowShellMeta` option),
THE SYSTEM SHALL bypass validation and execute the command as-is, logging a warning.

### R2: Branch Name Inline Validation (P1)

WHEN `addWorktreeWithRetry` receives a branch name,
THE SYSTEM SHALL validate inline that the branch name matches `^[a-zA-Z0-9/_.-]+$`
before constructing the git command arguments.

WHEN the branch name contains invalid characters,
THE SYSTEM SHALL return an error describing the invalid characters found.

### R3: Backward Compatibility

WHERE existing callers pass commands/branch names that are currently valid,
THE SYSTEM SHALL accept them without behavioral change.

### R4: Sanitize Function Enhancement

WHEN `sanitizeBranchName` is called,
THE SYSTEM SHALL additionally reject names containing `..`, names starting with `-`,
and names exceeding 255 characters, returning an error instead of silently replacing.

## 생성 파일 상세

| 파일 | 역할 |
|------|------|
| `pkg/experiment/cmdvalidate.go` | Shell metacharacter allowlist 검증 함수 |
| `pkg/experiment/cmdvalidate_test.go` | 검증 함수 단위 테스트 |
| `pkg/pipeline/branchvalidate.go` | Branch name regex 검증 함수 |
| `pkg/pipeline/branchvalidate_test.go` | Branch name 검증 단위 테스트 |

기존 파일 수정:
- `pkg/experiment/metric.go`: `RunMetric` 진입부에 검증 호출 추가
- `pkg/pipeline/worktree.go`: `addWorktreeWithRetry` 진입부에 검증 호출 추가, `sanitizeBranchName` 시그니처 변경 (error 반환)
