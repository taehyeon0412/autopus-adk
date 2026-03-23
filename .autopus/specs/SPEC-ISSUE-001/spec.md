# SPEC-ISSUE-001: Auto Issue Reporter

**Status**: done
**Created**: 2026-03-23
**Domain**: ISSUE

## 목적

Autopus 하네스 사용자가 CLI 에러, 파이프라인 실패, 설정 문제를 겪을 때 에러 컨텍스트를 자동 수집하고, 민감 정보를 제거한 후, `gh` CLI로 GitHub 이슈를 생성하는 기능을 제공한다. 수동 이슈 작성의 마찰을 제거하고, 재현에 필요한 환경 정보를 표준화된 형태로 포함시킨다.

## 요구사항

### P0 — Must Have

- **REQ-01**: WHEN `auto issue report` is executed, THE SYSTEM SHALL collect error context including error message, command, and exit code.
- **REQ-02**: WHEN collecting context, THE SYSTEM SHALL gather environment information (OS, Go version, auto version, platform).
- **REQ-03**: WHEN formatting the report, THE SYSTEM SHALL sanitize all file paths by replacing `$HOME` with the literal string `$HOME`.
- **REQ-04**: WHEN formatting the report, THE SYSTEM SHALL strip API keys, tokens, secrets, and git remote URLs from all collected data.
- **REQ-05**: WHEN the report is ready, THE SYSTEM SHALL display a markdown preview and require explicit user confirmation before submission.
- **REQ-06**: WHEN the user confirms, THE SYSTEM SHALL create a GitHub issue via `gh issue create` with the formatted report body.
- **REQ-07**: WHEN `gh` CLI is not installed or not authenticated, THE SYSTEM SHALL display an actionable error with setup instructions.

### P1 — Should Have

- **REQ-08**: WHEN submitting, THE SYSTEM SHALL compute an xxhash of (error_message + command) and search existing issues for duplicates.
- **REQ-09**: WHERE a matching open issue exists, THE SYSTEM SHALL add a comment to the existing issue instead of creating a new one.
- **REQ-10**: WHEN `--dry-run` flag is passed, THE SYSTEM SHALL display the full formatted report without submitting.
- **REQ-11**: WHEN `auto issue list` is executed, THE SYSTEM SHALL list previously submitted auto-reports from the target repository filtered by `auto-report` label.
- **REQ-12**: WHEN `auto issue search` is executed with a query, THE SYSTEM SHALL search existing issues in the target repository.

### P2 — Could Have

- **REQ-13**: WHEN `issue_report.auto_submit: true` is set in `autopus.yaml`, THE SYSTEM SHALL skip user confirmation and submit automatically.
- **REQ-14**: WHILE submitting, THE SYSTEM SHALL enforce a client-side rate limit of max 1 issue per 5 minutes per error hash.

## 생성 파일 상세

### pkg/issue/types.go
`IssueReport`, `IssueContext`, `SubmitResult`, `Config` 등 도메인 타입 정의. 다른 파일들이 의존하는 핵심 데이터 구조.

### pkg/issue/collector.go
`CollectContext()` 함수 — OS, Go 버전, auto 버전, 플랫폼 정보, autopus.yaml 설정, 최근 텔레메트리 데이터를 수집하여 `IssueContext` 를 반환.

### pkg/issue/sanitizer.go
`SanitizePath()`, `SanitizeConfig()`, `SanitizeEnv()` — 경로에서 `$HOME` 치환, API 키/토큰/시크릿 패턴 매칭 및 제거, git remote URL 제거.

### pkg/issue/formatter.go
`FormatMarkdown()` — Go 템플릿을 사용하여 `IssueReport`를 GitHub 이슈 본문 마크다운으로 변환. 65,536자 초과 시 truncate.

### pkg/issue/submitter.go
`Submit()` — `gh issue create` 실행, 중복 검색(`gh issue list --search`), 기존 이슈에 코멘트 추가. gh CLI 존재 여부 확인.

### internal/cli/issue.go
Cobra 커맨드 정의: `auto issue report`, `auto issue list`, `auto issue search`. `--dry-run`, `--auto-submit` 플래그 지원.

### pkg/config/schema.go (수정)
`HarnessConfig`에 `IssueReport IssueReportConfig` 필드 추가.

### templates/shared/issue-report.md.tmpl
이슈 본문 마크다운 템플릿. 에러 정보, 환경, 설정, 텔레메트리 섹션 포함.
