# SPEC-HARN-DOCFETCH-001 수락 기준

## 시나리오

### S1: Context7 기반 단일 라이브러리 문서 패치

- Given: Context7 REST API가 정상 응답하는 환경
- When: `auto docs fetch cobra` 실행
- Then: cobra 라이브러리 문서가 stdout에 출력되고, `.autopus/cache/docs/` 에 캐시 파일이 생성된다

### S2: Context7 실패 시 스크래핑 Fallback

- Given: Context7 REST API가 503 오류를 반환하는 환경
- When: `auto docs fetch github.com/spf13/cobra` 실행
- Then: pkg.go.dev에서 문서를 스크래핑하여 출력하고, `[DOCFETCH→WEB]` 로그가 stderr에 출력된다

### S3: 프롬프트 주입 포맷 출력

- Given: cobra 라이브러리 문서가 패치 가능한 환경
- When: `auto docs fetch cobra --format prompt --topic "command registration"` 실행
- Then: `## Reference Documentation` 헤더 + `### cobra (via Context7)` 섹션 포맷으로 출력된다

### S4: 캐시 히트

- Given: 이전에 cobra 문서를 패치하여 `.autopus/cache/docs/cobra_*.json` 캐시가 존재하고 TTL(24h) 이내인 환경
- When: `auto docs fetch cobra` 실행
- Then: 외부 API 호출 없이 캐시된 결과가 반환된다

### S5: 캐시 만료

- Given: cobra 캐시 파일이 존재하지만 TTL(24h)을 초과한 환경
- When: `auto docs fetch cobra` 실행
- Then: Context7 API를 다시 호출하고, 캐시 파일을 갱신한다

### S6: 자동 의존성 감지

- Given: `go.mod`에 `github.com/spf13/cobra v1.9.1`과 `gopkg.in/yaml.v3 v3.0.1`이 선언된 프로젝트
- When: `auto docs fetch --format prompt` (라이브러리 미지정) 실행
- Then: cobra, yaml.v3 두 라이브러리의 문서가 각각 ~3000 토큰 예산으로 패치되어 프롬프트 포맷으로 출력된다

### S7: 토큰 예산 적응형 관리

- Given: 5개 라이브러리가 감지된 프로젝트
- When: `auto docs fetch --format prompt` 실행
- Then: 각 라이브러리당 ~2000 토큰으로 트리밍되고, 총 출력이 10000 토큰을 초과하지 않는다

### S8: 표준 라이브러리 필터링

- Given: `go.mod`에 선언된 의존성과 import에 `fmt`, `os`, `net/http`가 사용된 프로젝트
- When: 자동 감지 실행
- Then: `fmt`, `os`, `net/http`는 표준 라이브러리로 필터링되어 패치 대상에서 제외된다

### S9: 캐시 관리 CLI

- Given: `.autopus/cache/docs/`에 3개 라이브러리 캐시가 존재하는 환경
- When: `auto docs cache list` 실행
- Then: 각 캐시된 라이브러리명, 크기, TTL 잔여 시간이 표시된다
- When: `auto docs cache clear` 실행
- Then: 모든 캐시 파일이 삭제되고, 삭제 건수가 출력된다

### S10: 파이프라인 Phase 1.8 통합 (SPEC-HARN-PIPE-001 의존)

- Given: `auto pipeline run`이 SPEC을 대상으로 실행 중인 환경
- When: Phase 1.8 Doc Fetch 단계 도달
- Then: `pkg/docs/fetcher.Fetch()`가 호출되어 감지된 라이브러리 문서가 자동 주입된다

### S11: MCP 환경 분기

- Given: Claude Code 환경에서 MCP Context7 도구가 사용 가능한 환경
- When: Phase 1.8 실행
- Then: MCP를 primary로 사용하고, `auto docs fetch`는 fallback으로만 대기한다

### S12: Non-MCP 환경 동작

- Given: Codex CLI 또는 Gemini CLI 환경 (MCP 없음)
- When: Phase 1.8 실행
- Then: `auto docs fetch`가 유일한 문서 소스로 사용되어 문서 주입이 정상 수행된다
