# SPEC-E2EFIX-001 수락 기준

## 시나리오

### S1: 모노레포 루트에서 ADK CLI 시나리오 빌드 성공
- Given: autopus-co 모노레포 루트에 scenarios.md가 존재하고, `## Build:` 줄에 `go build ./cmd/auto/ (ADK), go build ./cmd/server/ (Backend), npm run build (Frontend)`가 있다
- When: `auto test run --project-dir . --scenario version`을 모노레포 루트에서 실행한다
- Then: `go build ./cmd/auto/`가 `autopus-adk/` 디렉토리 내에서 실행되고, 빌드가 성공한다

### S2: 멀티 빌드 라인 파싱
- Given: `## Build: go build ./cmd/auto/ (ADK), go build ./cmd/server/ (Backend), npm run build (Frontend)` 문자열
- When: ParseBuildLine()을 호출한다
- Then: 3개의 BuildEntry가 반환되며, 각각 Label이 "ADK", "Backend", "Frontend"이고 Command가 올바르게 분리된다

### S3: 단일 빌드 커맨드 하위 호환
- Given: `## Build: go build -o auto ./cmd/auto` (단일 커맨드, 레이블 없음)
- When: ParseBuildLine()을 호출한다
- Then: 1개의 BuildEntry가 반환되며, 기존과 동일하게 ProjectDir에서 실행 가능하다

### S4: 시나리오-빌드 매칭
- Given: 시나리오 S1이 "ADK CLI Scenarios" 섹션에 속하고, BuildEntry에 Label "ADK"가 있다
- When: MatchBuild()로 시나리오에 맞는 빌드를 선택한다
- Then: Label "ADK"인 BuildEntry가 반환된다

### S5: 빌드 불필요 시나리오 스킵
- Given: 시나리오 S18 (curl 기반 API 테스트)이 "Backend API Scenarios" 섹션에 속한다
- When: 해당 시나리오를 실행한다
- Then: 빌드 단계가 스킵되고, 시나리오 커맨드만 직접 실행된다

### S6: 서브모듈 경로 해석
- Given: Label "ADK"가 BuildEntry에 있고, 모노레포 루트가 `/path/to/autopus-co`이다
- When: ResolveBuildDir()를 호출한다
- Then: `/path/to/autopus-co/autopus-adk`가 빌드 작업 디렉토리로 반환된다

### S7: 빌드 엔트리별 독립 실행
- Given: ADK 빌드는 성공하지만 Frontend 빌드는 npm이 없어 실패하는 환경이다
- When: ADK CLI 시나리오만 필터하여 실행한다
- Then: ADK 빌드만 실행되고 시나리오가 성공하며, Frontend 빌드는 트리거되지 않는다

### S8: 빈 Build 줄 처리
- Given: `## Build:` 줄이 비어있거나 없는 scenarios.md
- When: ParseScenarios()를 호출한다
- Then: Builds 슬라이스가 비어있고, 빌드 단계가 전체적으로 스킵된다
