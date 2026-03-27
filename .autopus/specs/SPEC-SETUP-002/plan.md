# SPEC-SETUP-002 구현 계획

## 태스크 목록

- [ ] T1: `multirepo_types.go` — MultiRepoInfo, RepoComponent, RepoDependency 타입 정의
- [ ] T2: `multirepo.go` — DetectMultiRepo, ScanRepoComponent, MapCrossRepoDeps 구현
- [ ] T3: `types.go` 수정 — ProjectInfo에 MultiRepo 필드 추가
- [ ] T4: `scanner.go` 수정 — Scan()에서 멀티레포 감지 및 컴포넌트별 집계 로직 추가
- [ ] T5: `multirepo_render.go` — 워크스페이스/워크플로우/바운더리 렌더링 헬퍼 구현
- [ ] T6: `renderer.go` 수정 — renderArchitecture, renderStructure에서 멀티레포 섹션 삽입
- [ ] T7: `scenarios.go` 수정 — 크로스 컴포넌트 시나리오 생성 로직 추가
- [ ] T8: 단위 테스트 — multirepo_test.go, multirepo_render_test.go
- [ ] T9: 통합 테스트 — 실제 autopus-co 구조로 end-to-end 검증

## 구현 전략

### 접근 방법

기존 `workspace.go`의 DetectWorkspaces 패턴을 참고하되, 근본적으로 다른 감지 메커니즘을 구현한다. 기존 워크스페이스 감지는 go.work, package.json 등 매니페스트 기반이지만, 멀티레포 감지는 `.git` 디렉토리 존재 여부 기반이다.

### 기존 코드 활용

- `workspace.go`의 `DetectWorkspaces()` — 패턴 참고 (호출 구조, 타입 네이밍)
- `scanner.go`의 `detectLanguages()`, `detectBuildFiles()` — 컴포넌트별 재활용
- `renderer.go`의 `renderArchitecture()`, `renderStructure()` — 삽입 지점
- `types.go`의 `Workspace` — 기존 타입과 공존, 충돌 없이 확장

### 변경 범위

- 신규 파일 3개: `multirepo.go`, `multirepo_types.go`, `multirepo_render.go`
- 수정 파일 4개: `types.go` (1줄 추가), `scanner.go` (약 30줄), `renderer.go` (약 20줄), `scenarios.go` (약 20줄)
- 테스트 파일 2개 (신규)

### 위험 요소

- `scanner.go`가 현재 573줄로 이미 300줄 제한을 초과하고 있음 — 추가 코드는 최소한으로 유지하고, 멀티레포 스캔 로직은 `multirepo.go`에 위치
- `renderer.go`가 467줄 — 렌더링 로직을 `multirepo_render.go`로 분리하여 제한 준수
- 단일 레포 사용자에게 영향 없도록 모든 멀티레포 코드 경로는 `MultiRepo != nil` 가드로 보호
