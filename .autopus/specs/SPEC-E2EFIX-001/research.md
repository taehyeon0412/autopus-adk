# SPEC-E2EFIX-001 리서치

## 기존 코드 분석

### 시나리오 파서: `autopus-adk/pkg/e2e/scenario.go`
- `ScenarioSet.Build` (string): `## Build:` 줄 전체를 단일 문자열로 저장 (L31)
- `reBuild` 정규식: `^## Build: (.+)$` — 줄 전체를 하나의 캡처 그룹으로 추출 (L40)
- `ParseScenarios()`: Build 매칭 시 `set.Build = m[1]`로 단순 대입 (L65-66)
- `RenderScenarios()`: `fmt.Fprintf(&buf, "## Build: %s\n", set.Build)`로 직렬화 (L123)
- 섹션 헤더(`## ADK CLI Scenarios` 등)는 현재 파싱되지 않음 — 시나리오에 섹션 정보가 없음

### 시나리오 러너: `autopus-adk/pkg/e2e/runner.go`
- `RunnerOptions.BuildCommand` (string): 단일 빌드 커맨드 (L18)
- `Runner.buildOnce` (sync.Once): 러너 인스턴스당 빌드를 한 번만 실행 (L48)
- `Run()` 내 빌드 실행: `exec.Command("sh", "-c", r.opts.BuildCommand)` → `cmd.Dir = r.opts.ProjectDir` (L70-71)
  - **핵심 문제**: 빌드 커맨드의 Dir가 항상 ProjectDir(모노레포 루트)로 고정됨
  - 서브모듈 안의 상대 경로(`./cmd/auto/`)를 루트에서 해석하므로 실패

### CLI 통합: `autopus-adk/internal/cli/test.go`
- `runAutoTest()`: `set.Build`를 그대로 `RunnerOptions.BuildCommand`에 전달 (L99, L105)
- 현재 값: `"go build ./cmd/auto/ (ADK), go build ./cmd/server/ (Backend), npm run build (Frontend)"` 전체가 하나의 셸 커맨드로 실행됨

### scenarios.md: `.autopus/project/scenarios.md`
- Build 줄: `go build ./cmd/auto/ (ADK), go build ./cmd/server/ (Backend), npm run build (Frontend)`
- 섹션 구분:
  - `## ADK CLI Scenarios (auto)` → S1-S15, S24-S27
  - `## Bridge CLI Scenarios (autopus-bridge)` → S16-S17
  - `## Backend API Scenarios (server)` → S18-S20
  - `## Frontend Scenarios` → S21-S23

### 모노레포 서브모듈 구조
```
autopus-co/
├── Autopus/              ← Backend (Go, cmd/server/)
├── autopus-adk/          ← ADK CLI (Go, cmd/auto/)
├── autopus-bridge/       ← Bridge CLI
├── autopus-agent-protocol/
├── autopus-codex-rpc/
└── homebrew-tap/
```

## 설계 결정

### Build 줄 포맷 규약
**결정**: 괄호 안 레이블 `(Label)`을 파싱 키로 사용한다.
- 입력: `go build ./cmd/auto/ (ADK), go build ./cmd/server/ (Backend)`
- 파싱: `[{Command: "go build ./cmd/auto/", Label: "ADK"}, {Command: "go build ./cmd/server/", Label: "Backend"}]`
- 레이블이 없으면 전체를 단일 커맨드로 취급 (하위 호환)

**이유**: 기존 scenarios.md 포맷을 변경하지 않고, 이미 존재하는 괄호 표기를 활용할 수 있다.

### 레이블 → 서브모듈 매핑 전략
**결정**: 하드코딩 대신, 프로젝트 루트의 디렉토리 스캔 + 레이블 퍼지 매칭을 사용한다.
- "ADK" → `autopus-adk/` (contains "adk", case-insensitive)
- "Backend" → `Autopus/` (contains go.mod with cmd/server/)
- "Frontend" → 프론트엔드 디렉토리 (contains package.json)
- 매칭 실패 시 ProjectDir 폴백

**대안 검토**:
1. autopus.yaml에 명시적 매핑 추가 → 설정 부담 증가, 기존 프로젝트 마이그레이션 필요
2. Build 줄에 디렉토리 경로 명시 (`go build ./cmd/auto/ @autopus-adk/`) → scenarios.md 포맷 변경 필요
3. 시나리오별 Build 필드 추가 → scenarios.md 스키마 대폭 변경

### 빌드 once 전략
**결정**: `map[string]*sync.Once`로 빌드 엔트리별 독립 once 관리
- 키: BuildEntry.Label (또는 Label이 없으면 Command 해시)
- ADK 시나리오 실행 시 ADK 빌드만 한 번, Backend 시나리오 실행 시 Backend 빌드만 한 번

**이유**: 기존 `sync.Once`는 러너당 하나였으나, 멀티 빌드에서는 불필요한 빌드를 방지해야 함.

### Scenario.Section 파싱
**결정**: `## {SectionName} Scenarios` 형태의 줄을 감지하여, 이후 시나리오에 Section 필드를 부여한다.
- `## ADK CLI Scenarios (auto)` → Section: "ADK CLI"
- `## Bridge CLI Scenarios (autopus-bridge)` → Section: "Bridge CLI"

**이유**: 시나리오에 명시적 태그를 추가하면 기존 scenarios.md를 수정해야 하지만, 섹션 헤더는 이미 존재하므로 파싱만 추가하면 된다.
