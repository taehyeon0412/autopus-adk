# SPEC-E2EFIX-001 구현 계획

## 태스크 목록

- [ ] T1: BuildEntry 구조체 및 파서 구현 (pkg/e2e/build.go 신규)
  - BuildEntry 구조체 정의 (Command, Label, SubmodulePath)
  - ParseBuildLine() 함수: 쉼표 구분 빌드 라인을 []BuildEntry로 파싱
  - ResolveBuildDir() 함수: 레이블 → 서브모듈 경로 매핑 (autopus.yaml 또는 규칙 기반)
  - MatchBuild() 함수: 시나리오 섹션/태그에 맞는 BuildEntry 선택

- [ ] T2: ScenarioSet 파서 확장 (pkg/e2e/scenario.go 수정)
  - ScenarioSet에 Builds []BuildEntry 필드 추가
  - ParseScenarios()에서 ## Build: 줄을 ParseBuildLine()으로 위임
  - 기존 Build string 필드를 deprecated 처리 (하위 호환)
  - RenderScenarios()에서 Builds → ## Build: 줄 직렬화
  - Scenario 구조체에 Section string 필드 추가 (섹션 헤더 파싱)

- [ ] T3: Runner 멀티 빌드 지원 (pkg/e2e/runner.go 수정)
  - RunnerOptions에 Builds []BuildEntry 추가
  - 빌드 엔트리별 sync.Once 맵으로 변경
  - Run()에서 MatchBuild()로 시나리오에 맞는 빌드 선택
  - 선택된 빌드의 SubmodulePath를 WorkDir로 사용

- [ ] T4: CLI 통합 (internal/cli/test.go 수정)
  - runAutoTest()에서 set.Builds를 RunnerOptions에 전달
  - 단일 BuildCommand 폴백 유지 (하위 호환)

- [ ] T5: 유닛 테스트 작성
  - ParseBuildLine() 테스트: 단일/멀티/빈 빌드 라인
  - ResolveBuildDir() 테스트: 알려진/미알려진 레이블
  - MatchBuild() 테스트: 섹션별 매칭
  - Runner 멀티 빌드 통합 테스트

- [ ] T6: 기존 테스트 업데이트
  - scenario_test.go, scenario_coverage_test.go: Builds 필드 검증 추가
  - runner_test.go: 멀티 빌드 시나리오 추가
  - edge_cases_test.go: 빌드 실패 시 동작 확인

## 구현 전략

### 접근 방법
1. **새 파일 분리**: BuildEntry 관련 로직을 `build.go`로 분리하여 300줄 제한 준수
2. **하위 호환 유지**: 기존 단일 Build 문자열도 동작하도록 폴백 로직 포함
3. **섹션 인식 파싱**: scenarios.md의 `## ADK CLI Scenarios` 등 섹션 헤더를 파싱하여 시나리오에 섹션 정보 부여
4. **레이블 기반 매핑**: `(ADK)` → `autopus-adk/`, `(Backend)` → `Autopus/` 등 괄호 안 레이블로 서브모듈 매핑

### 변경 범위
- 신규: 1개 파일 (build.go)
- 수정: 3개 파일 (scenario.go, runner.go, test.go)
- 테스트: 기존 테스트 파일 업데이트 + 신규 build_test.go

### 위험 요소
- scenarios.md 포맷 변경 시 기존 프로젝트와의 호환성 → Build string 폴백으로 완화
- 서브모듈 경로 매핑이 프로젝트마다 다를 수 있음 → autopus.yaml에 매핑 설정 추가 고려
