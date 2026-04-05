# SPEC-AXQUAL-001: Code Quality — resolvePlatform Unit Tests and Template TODO Documentation

**Status**: completed
**Created**: 2026-04-05
**Domain**: AXQUAL
**Module**: autopus-adk

## 목적

코드베이스 AX 스캔에서 발견된 @AX:TODO 태그를 해결하고 코드 품질을 개선한다. 유일한 실제 코드 결함은 `resolvePlatform` 함수의 PATH 의존 동작에 대한 단위 테스트 부재이다. 나머지 TODO는 템플릿 콘텐츠로, 의도적 마커임을 문서화하여 AX 추적에서 제외한다.

## 요구사항

### P0 — Must Have

**REQ-1**: WHEN `resolvePlatform` is called with an explicit platform string, THE SYSTEM SHALL return the string as-is without PATH lookup.

**REQ-2**: WHEN `resolvePlatform` is called with an empty string and a recognized binary (`claude`, `codex`, or `gemini`) exists in PATH, THE SYSTEM SHALL return the first matching binary name in priority order.

**REQ-3**: WHEN `resolvePlatform` is called with an empty string and no recognized binary exists in PATH, THE SYSTEM SHALL return `"claude"` as the default fallback.

**REQ-4**: WHEN `resolvePlatform` is called with an empty string and multiple recognized binaries exist in PATH, THE SYSTEM SHALL return the highest-priority candidate (`claude` > `codex` > `gemini`).

**REQ-5**: WHEN the `resolvePlatform` unit tests execute, THE SYSTEM SHALL use a mocked or isolated PATH environment to avoid host-dependent test flakiness.

### P1 — Should Have

**REQ-6**: WHEN a `TODO` string appears inside `agentTemplate` or `skillTemplate` const literals, THE SYSTEM SHALL treat these as intentional user-facing template markers, not as code defects.

**REQ-7**: WHERE template TODO markers are intentional, THE SYSTEM SHALL document them in an `@AX:EXCLUDE` annotation or equivalent comment to prevent future AX scans from flagging them.

**REQ-8**: WHEN the @AX:TODO on `pipeline_run.go:101` is resolved by adding tests, THE SYSTEM SHALL remove the @AX:TODO tag from the source file.

## 생성 파일 상세

| File | Role |
|------|------|
| `internal/cli/pipeline_run_test.go` | `resolvePlatform` table-driven unit tests 추가 (기존 파일에 append) |
| `internal/cli/pipeline_run.go` | @AX:TODO 태그 제거 |
| `internal/cli/agent_create.go` | 템플릿 TODO에 @AX:EXCLUDE 주석 추가 |
| `internal/cli/skill_create.go` | 템플릿 TODO에 @AX:EXCLUDE 주석 추가 |
