# SPEC-SETUP-002: Multi-Repo Workspace Detection and Cross-Repo Dependency Mapping

**Status**: approved
**Created**: 2026-03-25
**Domain**: SETUP

## 목적

현재 `auto setup generate` 및 `auto arch generate` 명령은 단일 git 레포 또는 go.work/npm workspace 기반 모노레포만 인식한다. autopus-co처럼 루트 디렉토리에 `.git`이 없고 각 서브디렉토리가 독립 git 레포인 "멀티레포 워크스페이스" 구조를 감지하지 못한다.

이 SPEC은 setup 바이너리에 멀티레포 워크스페이스 감지, 크로스 레포 의존성 매핑, 워크스페이스 수준 문서 생성 기능을 추가하여 AI 에이전트가 레포 경계와 컴포넌트 관계를 정확히 이해할 수 있게 한다.

## 요구사항

### R1: Multi-Repo Workspace Detection (P0)
WHEN the project root directory does NOT contain a `.git` directory, THE SYSTEM SHALL scan immediate subdirectories for `.git` directories to identify independent git repositories. Each detected repository SHALL be recorded as a `RepoComponent` with name, path, git remote URL, primary language, and module path.

### R2: Backward Compatibility (P0)
WHEN the project root directory contains a `.git` directory (single-repo), THE SYSTEM SHALL behave exactly as before, skipping multi-repo detection entirely.

### R3: Cross-Repo Dependency Mapping (P0)
WHEN a multi-repo workspace is detected, THE SYSTEM SHALL parse each repository's `go.mod` for `replace` directives and `require` statements that reference sibling repositories. The result SHALL be a directed dependency graph stored as `[]RepoDependency` with source repo, target repo, dependency type (replace/require), and module version.

### R4: NPM/Package Cross-Reference Detection (P1)
WHEN a multi-repo workspace is detected, THE SYSTEM SHALL also parse `package.json` files for cross-references between sibling repositories using package names or file: protocol links.

### R5: Workspace Section in Architecture Document (P0)
WHEN a multi-repo workspace is detected, THE SYSTEM SHALL generate a "Workspace" section in `architecture.md` containing: workspace type (multi-repo), repository list with roles, dependency graph in text format, and deploy target mapping.

### R6: Development Workflow Section (P0)
WHEN a multi-repo workspace is detected, THE SYSTEM SHALL generate a "Development Workflow" section in `architecture.md` documenting: which repository handles which concern, cross-repo change coordination strategy, and local development setup using replace directives.

### R7: Repository Boundaries in Structure Document (P0)
WHEN a multi-repo workspace is detected, THE SYSTEM SHALL add repository boundary indicators and git remote information to `structure.md`, marking each top-level directory as `[git repo]` with its remote URL.

### R8: Cross-Component Scenarios (P1)
WHEN a multi-repo workspace is detected, THE SYSTEM SHALL generate cross-component end-to-end scenarios in `scenarios.md` that span multiple repositories, based on the detected dependency graph.

### R9: ProjectInfo Extension (P0)
THE SYSTEM SHALL extend the `ProjectInfo` struct with a `MultiRepo *MultiRepoInfo` field. `MultiRepoInfo` SHALL contain: `IsMultiRepo bool`, `Components []RepoComponent`, `Dependencies []RepoDependency`, and `WorkspaceRoot string`.

### R10: Scan Aggregation (P0)
WHEN a multi-repo workspace is detected, THE SYSTEM SHALL scan each component repository independently and aggregate results. The aggregated `ProjectInfo` SHALL include languages, frameworks, build files, and entry points from all component repositories.

## 생성 파일 상세

### `pkg/setup/multirepo.go` (신규)
멀티레포 감지 핵심 로직: `DetectMultiRepo(dir string) *MultiRepoInfo`, `ScanRepoComponent(dir string) (*RepoComponent, error)`, `MapCrossRepoDeps(components []RepoComponent) []RepoDependency`.

### `pkg/setup/multirepo_types.go` (신규)
멀티레포 관련 타입 정의: `MultiRepoInfo`, `RepoComponent`, `RepoDependency`.

### `pkg/setup/multirepo_render.go` (신규)
문서 렌더링 헬퍼: `renderWorkspaceSection(info *MultiRepoInfo) string`, `renderDevWorkflow(info *MultiRepoInfo) string`, `renderRepoBoundaries(info *MultiRepoInfo) string`.

### 기존 파일 수정
- `pkg/setup/types.go` — `ProjectInfo`에 `MultiRepo *MultiRepoInfo` 필드 추가
- `pkg/setup/scanner.go` — `Scan()`에서 멀티레포 감지 호출, 컴포넌트별 스캔 집계
- `pkg/setup/renderer.go` — `renderArchitecture()`와 `renderStructure()`에서 멀티레포 섹션 삽입
- `pkg/setup/scenarios.go` — 크로스 컴포넌트 시나리오 생성 로직 추가
