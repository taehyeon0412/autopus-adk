# SPEC-E2EFIX-001: E2E 시나리오 러너 모노레포 빌드 경로 해석 수정

**Status**: completed
**Created**: 2026-03-26
**Domain**: E2EFIX

## 목적

모노레포(autopus-co) 루트에서 `auto test run --project-dir .`을 실행할 때, E2E 시나리오 러너가
서브모듈별 빌드 커맨드와 작업 디렉토리를 올바르게 해석하지 못하는 문제를 수정한다.

현재 `scenarios.md`의 `## Build:` 줄은 복수의 빌드 커맨드를 포함하지만(ADK, Backend, Frontend),
파서(`ParseScenarios`)는 이를 단일 문자열로 저장하고, 러너(`Runner`)는 해당 문자열을
`ProjectDir`(모노레포 루트)에서 그대로 실행한다. `go build ./cmd/auto/`는 `autopus-adk/` 안에서만
유효하므로 빌드가 실패한다.

## 요구사항

### R1: 멀티 빌드 커맨드 파싱
WHEN scenarios.md의 `## Build:` 줄에 쉼표로 구분된 복수의 빌드 항목이 존재할 때,
THE SYSTEM SHALL 각 항목을 독립된 BuildEntry(command + label)로 파싱하여 ScenarioSet.Builds 슬라이스에 저장한다.

### R2: 서브모듈 경로 매핑
WHEN BuildEntry에 서브모듈 레이블(예: "ADK", "Backend", "Frontend")이 포함되어 있을 때,
THE SYSTEM SHALL 모노레포 루트 기준으로 해당 서브모듈 디렉토리를 자동 매핑하여 빌드 작업 디렉토리(WorkDir)를 결정한다.

### R3: 시나리오별 빌드 선택
WHEN 시나리오가 실행될 때,
THE SYSTEM SHALL 시나리오의 섹션 헤더(예: "ADK CLI Scenarios", "Bridge CLI Scenarios") 또는 명시적 태그를 기반으로 해당 시나리오에 적합한 빌드 커맨드를 선택하여 실행한다.

### R4: 하위 호환성
WHEN scenarios.md의 `## Build:` 줄에 단일 빌드 커맨드만 존재할 때,
THE SYSTEM SHALL 기존 동작과 동일하게 ProjectDir에서 해당 커맨드를 실행한다.

### R5: 빌드 스킵
WHERE 시나리오가 빌드가 불필요한 유형(예: curl 기반 API 테스트, 프론트엔드 네비게이션 테스트)일 때,
THE SYSTEM SHALL 해당 시나리오의 빌드 단계를 건너뛴다.

## 생성 파일 상세

### pkg/e2e/scenario.go (수정)
- `ScenarioSet` 구조체에 `Builds []BuildEntry` 필드 추가
- `BuildEntry` 구조체 정의: `Command`, `Label`, `SubmodulePath` 필드
- `ParseScenarios()`에서 `## Build:` 줄의 멀티 빌드 파싱 로직 추가
- `RenderScenarios()`에서 멀티 빌드 직렬화 유지

### pkg/e2e/runner.go (수정)
- `RunnerOptions`에 `Builds []BuildEntry` 필드 추가 (단일 BuildCommand 대체)
- `Run()`에서 시나리오 섹션에 맞는 빌드 커맨드 선택 및 올바른 WorkDir에서 실행
- `buildOnce` 대신 빌드 엔트리별 once 관리

### pkg/e2e/build.go (신규)
- `BuildEntry` 구조체 정의
- `ParseBuildLine()`: `## Build:` 줄 문자열을 `[]BuildEntry`로 파싱
- `ResolveBuildDir()`: 레이블 기반 서브모듈 디렉토리 매핑
- `MatchBuild()`: 시나리오에 적합한 BuildEntry 선택

### internal/cli/test.go (수정)
- `runAutoTest()`에서 `set.Builds`를 `RunnerOptions.Builds`에 전달
- 단일 `BuildCommand` 대신 복수 빌드 지원
