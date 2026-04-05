# SPEC-HARN-DOCFETCH-001: 플랫폼-무관 문서 자동 주입 시스템

**Status**: completed
**Created**: 2026-04-05
**Domain**: HARN-DOCFETCH
**Priority**: Must-Have
**Depends on**: SPEC-HARN-PIPE-001

## 목적

현재 Autopus-ADK의 문서 자동 주입(Phase 1.8 Doc Fetch)은 Claude Code의 Context7 MCP 도구에 의존한다. Codex CLI에는 MCP 지원이 없고 Gemini CLI는 제한적이어서, 해당 파이프라인에서 외부 라이브러리 문서가 주입되지 않아 구현 품질이 저하된다.

Go 바이너리(`auto`) 내부에 문서 패치 기능을 내장하여, 어떤 AI 플랫폼에서든 동일한 문서 주입 품질을 보장한다.

## 요구사항

### Domain 1: Go 바이너리 기반 문서 패치 (Must-Have)

- **REQ-DF-001**: WHEN a user runs `auto docs fetch <library>`, THE SYSTEM SHALL resolve the library ID via Context7 REST API (HTTP direct, no MCP) and return formatted documentation to stdout.
- **REQ-DF-002**: WHEN Context7 REST API is unavailable or returns no result, THE SYSTEM SHALL fall back to scraping official documentation sources (pkg.go.dev, npmjs.com, pypi.org) in priority order.
- **REQ-DF-003**: WHEN a cached result exists for the requested library+topic and is within TTL (24h), THE SYSTEM SHALL return the cached result without making external API calls.
- **REQ-DF-004**: WHEN `--format prompt` flag is provided, THE SYSTEM SHALL output documentation in `## Reference Documentation` injection format compatible with executor prompts.
- **REQ-DF-005**: WHEN `--topic <topic>` flag is provided, THE SYSTEM SHALL pass the topic parameter to narrow documentation scope.
- **REQ-DF-006**: THE SYSTEM SHALL enforce adaptive token budgets: 1 lib ~5000 tokens, 2 libs ~3000 each, 3 libs ~2500 each, 4-5 libs ~2000 each, hard cap 10000 total.

### Domain 2: 기술 감지 확장 (Must-Have)

- **REQ-DF-010**: WHEN `auto docs fetch` is invoked without `--lib` flag, THE SYSTEM SHALL auto-detect dependencies from `go.mod`, `package.json`, `pyproject.toml`, or `requirements.txt` in the current project.
- **REQ-DF-011**: WHEN scanning dependencies, THE SYSTEM SHALL reuse `pkg/setup` scanner's language/framework detection logic rather than reimplementing.
- **REQ-DF-012**: THE SYSTEM SHALL skip standard library modules (Go: `fmt`, `os`, `net/http`; Node.js: `fs`, `path`, `http`; Python: `os`, `sys`, `json`) during auto-detection.
- **REQ-DF-013**: WHEN SPEC or plan.md files reference library names, THE SYSTEM SHALL extract those names as additional detection sources.

### Domain 3: 파이프라인 통합 (Should-Have)

- **REQ-DF-020**: WHEN `auto pipeline run` executes Phase 1.8, THE SYSTEM SHALL invoke `auto docs fetch` internally and inject results into executor/tester prompts.
- **REQ-DF-021**: WHILE in a Claude Code environment with MCP available, THE SYSTEM SHALL use MCP as primary and `auto docs fetch` as fallback (not replacement).
- **REQ-DF-022**: WHILE in a non-MCP environment (Codex, Gemini), THE SYSTEM SHALL use `auto docs fetch` as the sole documentation source.
- **REQ-DF-023**: THE SYSTEM SHALL store cache files in `.autopus/cache/docs/` with library-specific filenames and TTL metadata.

### Domain 4: 캐시 관리 (Could-Have)

- **REQ-DF-030**: WHEN `auto docs cache clear` is invoked, THE SYSTEM SHALL remove all cached documentation files.
- **REQ-DF-031**: WHEN `auto docs cache list` is invoked, THE SYSTEM SHALL display cached libraries with TTL remaining.

## 생성 파일 상세

| 패키지/파일 | 역할 |
|---|---|
| `pkg/docs/fetcher.go` | 문서 패치 오케스트레이션 (소스 우선순위, 토큰 예산) |
| `pkg/docs/context7.go` | Context7 REST API 직접 호출 클라이언트 |
| `pkg/docs/scraper.go` | pkg.go.dev / npmjs / pypi 공식 문서 스크래핑 |
| `pkg/docs/cache.go` | `.autopus/cache/docs/` 기반 TTL 캐시 |
| `pkg/docs/detect.go` | 프로젝트 의존성 스캔 (pkg/setup 재활용) |
| `pkg/docs/format.go` | 프롬프트 주입 포맷 렌더링 |
| `pkg/docs/types.go` | 공통 타입 정의 |
| `internal/cli/docs_fetch.go` | `auto docs fetch` CLI 커맨드 |
| `internal/cli/docs_cache.go` | `auto docs cache` CLI 서브커맨드 |
