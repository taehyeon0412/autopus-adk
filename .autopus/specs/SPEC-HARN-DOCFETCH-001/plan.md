# SPEC-HARN-DOCFETCH-001 구현 계획

## 태스크 목록

### Phase 1: Core Library (`pkg/docs/`)

- [ ] T1: `pkg/docs/types.go` — 공통 타입 정의 (DocResult, FetchOptions, CacheEntry, TokenBudget)
- [ ] T2: `pkg/docs/context7.go` — Context7 REST API HTTP 클라이언트 (MCP 없이 직접 호출)
- [ ] T3: `pkg/docs/scraper.go` — 공식 문서 소스 스크래핑 (pkg.go.dev, npmjs.com, pypi.org)
- [ ] T4: `pkg/docs/cache.go` — `.autopus/cache/docs/` 파일 기반 TTL(24h) 캐시
- [ ] T5: `pkg/docs/detect.go` — 프로젝트 의존성 자동 감지 (`pkg/setup.Scan()` 재활용 + SPEC/plan.md 파싱)
- [ ] T6: `pkg/docs/format.go` — `## Reference Documentation` 프롬프트 주입 포맷 렌더러
- [ ] T7: `pkg/docs/fetcher.go` — 패치 오케스트레이션 (Context7 → scraper → cache fallback chain, 토큰 예산 관리)

### Phase 2: CLI Integration

- [ ] T8: `internal/cli/docs_fetch.go` — `auto docs fetch` 커맨드 (기존 `newDocsCmd()` 확장 또는 교체)
- [ ] T9: `internal/cli/docs_cache.go` — `auto docs cache [list|clear]` 서브커맨드

### Phase 3: Pipeline Integration (SPEC-HARN-PIPE-001 의존)

- [ ] T10: Pipeline Phase 1.8 에서 `pkg/docs/fetcher.Fetch()` 호출 통합
- [ ] T11: 환경 감지 로직 (MCP 가용 여부에 따른 경로 분기)

### Phase 4: Tests

- [ ] T12: `pkg/docs/` 유닛 테스트 (각 파일별)
- [ ] T13: `internal/cli/docs_*` CLI 테스트
- [ ] T14: `.autopus/cache/docs/` 를 `.gitignore`에 추가

## 구현 전략

### Context7 REST API 활용

기존 `pkg/search/context7.go`에 Context7 HTTP 클라이언트가 이미 존재한다. 이를 참고하되, `pkg/docs/context7.go`에 문서 패치 전용 클라이언트를 별도 구현한다. 이유:
- `pkg/search`는 검색 도메인, `pkg/docs`는 문서 패치 도메인 — 관심사 분리
- 문서 패치에는 토큰 트리밍, 캐시 키 생성 등 추가 로직이 필요
- `pkg/search/context7.go`의 API 엔드포인트(`/api/v1/libraries`, `/api/v1/libraries/{id}/docs`)를 그대로 사용

### 기술 감지 재활용

`pkg/setup/scanner.go`의 `detectLanguages()`, `detectFrameworks()` 결과를 재활용한다. `Scan()` 함수가 `ProjectInfo`를 반환하므로, `pkg/docs/detect.go`에서 `setup.Scan()`을 호출하여 `Languages`, `Frameworks` 필드에서 의존성 목록을 추출한다.

### Fallback Chain

```
Context7 REST API → 공식 문서 스크래핑 → 로컬 캐시 → skip (non-fatal)
```

각 단계 실패 시 다음 단계로 넘어가며, 전체 프로세스는 절대 파이프라인을 블로킹하지 않는다.

### 파일 크기 관리

모든 소스 파일 200줄 이내 목표 (300줄 하드 리밋). 책임별 분리:
- 타입 정의, API 호출, 스크래핑, 캐시, 감지, 포맷팅, 오케스트레이션을 각각 독립 파일로 분리

### 기존 코드 영향

- `internal/cli/search.go`의 `newDocsCmd()`를 `internal/cli/docs_fetch.go`로 이전 및 확장
- `search.go`에서 `newDocsCmd()` 제거 후 `newSearchCmd()`만 유지
- `root.go`에서 `newDocsCmd()` 호출을 새 구현으로 교체
