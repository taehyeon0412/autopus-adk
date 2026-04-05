# SPEC-HARN-DOCFETCH-001 리서치

## 기존 코드 분석

### Context7 HTTP 클라이언트 (`pkg/search/context7.go`)

기존 구현이 이미 Context7 REST API를 HTTP로 직접 호출하고 있다:
- `ResolveLibrary(name)` → `GET /api/v1/libraries?name={name}` → library ID 반환
- `GetDocs(libraryID, topic)` → `GET /api/v1/libraries/{id}/docs?topic={topic}` → content 반환
- Base URL: `https://context7.com/api/v1`
- 인증: 현재 구현에 API key 헤더 없음 (무인증 엔드포인트 사용 중)

이 코드는 `pkg/docs`에서 참조 패턴으로 재활용 가능하나, 직접 import하지 않는다. `pkg/docs/context7.go`에 문서 패치 전용 클라이언트를 구현하여 토큰 트리밍, 에러 분류, 캐시 키 생성 로직을 추가한다.

### CLI 커맨드 구조 (`internal/cli/`)

- `root.go:44` — `newDocsCmd()` 이미 등록됨 (현재 `search.go`에 정의)
- 현재 `auto docs <library>` — 단순 Context7 조회 + stdout 출력
- 확장 방향: `auto docs fetch [--lib] [--topic] [--format prompt]` + `auto docs cache [list|clear]`
- 기존 `newDocsCmd()`를 `internal/cli/docs_fetch.go`로 이전하고 서브커맨드 구조로 전환

### 기술 감지 (`pkg/setup/scanner.go`)

- `Scan()` → `ProjectInfo` 반환 (Languages, Frameworks, BuildFiles 포함)
- `detectLanguages()` — Go(`go.mod`), TypeScript/JS(`package.json`), Python(`pyproject.toml`, `requirements.txt`), Rust(`Cargo.toml`) 감지
- `detectFrameworks()` — React, Vue, Next.js, Express, NestJS, Angular 감지 (`package.json` deps 기반)
- 의존성 목록 추출 가능: `go.mod`의 require 블록, `package.json`의 dependencies/devDependencies

`pkg/docs/detect.go`에서 `setup.Scan()`을 호출하되, 추가로 다음을 구현:
1. `go.mod` require 블록 전체 의존성 파싱 (setup은 언어만 감지, 개별 의존성은 미추출)
2. SPEC/plan.md 파일에서 라이브러리 이름 추출 (정규식 기반)
3. 표준 라이브러리 필터링 목록

### 파이프라인 (`pkg/pipeline/`, `internal/cli/pipeline.go`)

- `pipeline.Checkpoint` — Phase별 상태 관리 (pending/in_progress/done/failed)
- Phase 1.8 통합은 SPEC-HARN-PIPE-001에 의존하므로, 이 SPEC에서는 `pkg/docs` 패키지의 API를 파이프라인에서 호출 가능한 형태로 설계하는 것이 핵심

## Research Questions 결과

### Q1: Context7 API는 MCP 없이 HTTP REST로 접근 가능한가?

**결론: 가능하다 (High confidence)**

- 기존 `pkg/search/context7.go`가 이미 HTTP 직접 호출로 동작 중
- Context7 공식 문서에 REST API Reference (`context7.com/docs/api-guide`) 존재
- `CONTEXT7_API_KEY` 헤더로 인증 가능 (무인증도 rate-limited로 동작)
- MCP 서버(`mcp.context7.com/mcp`)는 내부적으로 같은 REST API를 래핑

### Q2: pkg.go.dev API가 Go 패키지 문서를 구조화된 형태로 제공하는가?

**결론: 공식 REST API는 없다 (Medium confidence)**

- golang/go#36785에서 API 요청이 2019년부터 열려 있으나 공식 API 미제공
- `proxy.golang.org` — 모듈 버전/소스 다운로드용 (문서 아님)
- 대안: pkg.go.dev HTML 페이지를 HTTP GET + HTML 파싱으로 문서 추출
- `pkg.go.dev/{module}` 페이지의 `<div class="Documentation">` 섹션에 렌더된 godoc 존재
- 구조화된 API 대신 HTML 스크래핑이 현실적 접근법

### Q3: npm registry API에서 README/type definitions를 추출할 수 있는가?

**결론: README 추출 가능, type definitions는 별도 경로 (High confidence)**

- `GET https://registry.npmjs.org/{package}` → `readme` 필드에 마크다운 README 포함
- `GET https://registry.npmjs.org/{package}/{version}` → 특정 버전 정보
- TypeScript type definitions: `@types/{package}` 패키지 또는 패키지 내 `.d.ts` 파일 (npm tarball 다운로드 필요)
- README만으로도 대부분의 API 문서 커버 가능

## 설계 결정

### Decision 1: `pkg/docs` 신규 패키지 vs `pkg/search` 확장

**결정: `pkg/docs` 신규 패키지**

이유:
- `pkg/search`는 웹 검색(Exa) + 문서 조회(Context7) — 범용 검색 도메인
- `pkg/docs`는 파이프라인 통합, 토큰 예산, 캐시, 자동 감지 — 문서 주입 전용 도메인
- 관심사 분리가 명확하고, 파일 크기 제한(300줄) 준수에도 유리

대안:
- `pkg/search/context7.go` 확장 → 파일 크기 초과 위험, 검색/문서 혼재
- `pkg/search/docs/` 서브패키지 → Go 패키지 관습에 맞지 않음

### Decision 2: 스크래핑 범위

**결정: 최소 스크래핑 — README/godoc만 추출**

이유:
- pkg.go.dev: HTML `Documentation` 섹션에서 godoc 텍스트 추출
- npm: registry API의 `readme` 필드 (JSON, 스크래핑 불필요)
- PyPI: `GET https://pypi.org/pypi/{package}/json` → `info.description` 필드

풀 스크래핑(튜토리얼, 가이드 등)은 토큰 예산 대비 가치가 낮고 유지보수 부담이 크다.

### Decision 3: 캐시 전략

**결정: 파일 기반 JSON 캐시, TTL 24h**

구조:
```
.autopus/cache/docs/
  {library-hash}.json  # { "library": "...", "topic": "...", "content": "...", "fetched_at": "...", "source": "context7|scraper" }
```

이유:
- 외부 의존성(Redis, SQLite) 불필요
- Git에서 제외 (`.gitignore`)
- 파이프라인 실행 간 공유 가능
- 24h TTL은 문서 업데이트 빈도 대비 적절한 균형

### Decision 4: 기존 `newDocsCmd()` 처리

**결정: `search.go`에서 분리하여 `docs_fetch.go`로 이전 + 확장**

이유:
- 현재 `search.go`에 `newDocsCmd()`와 `newSearchCmd()`가 혼재
- `search.go`는 검색 전용으로 유지
- `docs_fetch.go`에서 서브커맨드 구조로 전환: `auto docs fetch`, `auto docs cache`
