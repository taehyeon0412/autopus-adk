# SPEC-ISSUE-001 구현 계획

## 태스크 목록

- [x] T1: `pkg/issue/types.go` — IssueReport, IssueContext, SubmitResult, Config 타입 정의
- [x] T2: `pkg/issue/sanitizer.go` + `pkg/issue/sanitizer_test.go` — 경로/키/토큰/시크릿/URL 삭제 로직 및 테스트
- [x] T3: `pkg/issue/collector.go` + `pkg/issue/collector_test.go` — 환경/설정/텔레메트리 수집 로직 및 테스트
- [x] T4: `templates/shared/issue-report.md.tmpl` — 이슈 본문 마크다운 템플릿
- [x] T5: `pkg/issue/formatter.go` + `pkg/issue/formatter_test.go` — 템플릿 렌더링 및 truncation 로직
- [x] T6: `pkg/issue/submitter.go` + `pkg/issue/submitter_test.go` — gh CLI 실행, 중복 검색, 이슈 생성/코멘트
- [x] T7: `pkg/config/schema.go` 수정 — IssueReportConfig 타입 추가 및 HarnessConfig에 필드 추가
- [x] T8: `internal/cli/issue.go` + `internal/cli/issue_test.go` — Cobra 커맨드 (report, list, search)
- [x] T9: `templates/embed.go` 수정 — issue-report.md.tmpl embed glob 추가
- [x] T10: `internal/cli/root.go` 수정 — `newIssueCmd()` 등록

## 구현 전략

### 접근 방법
TDD 기반으로 타입 정의(T1) 후 sanitizer(T2), collector(T3), formatter(T5), submitter(T6) 순서로 하위 패키지를 먼저 구현한다. 각 단계마다 테스트를 작성한 뒤 구현을 진행한다. CLI 통합(T8-T10)은 패키지 레이어가 완성된 후 진행한다.

### 기존 코드 활용
- **xxhash**: `github.com/cespare/xxhash/v2` — 이미 go.mod에 의존성 존재. `pkg/search/hash.go`, `internal/cli/hash.go`에서 사용 패턴 참조.
- **Cobra CLI 패턴**: `internal/cli/telemetry.go`의 서브커맨드 그룹 패턴 (`cmd.AddCommand()`) 그대로 적용.
- **Config 패턴**: `pkg/config/schema.go`의 기존 `*Conf` 타입 네이밍 컨벤션 (`IssueReportConf`) 사용.
- **Template embed**: `templates/embed.go`의 `//go:embed` glob에 `shared/*.tmpl` 이미 포함.
- **Telemetry 타입**: `pkg/telemetry/types.go`의 `PipelineRun`, `AgentRun` 타입을 컨텍스트 수집 시 참조.
- **version 패키지**: 바이너리 버전 정보 수집에 활용.

### 변경 범위
- 신규 파일: 8개 소스 + 4개 테스트 + 1개 템플릿 = 13개
- 수정 파일: 3개 (`schema.go`, `embed.go`, `root.go`)
- 의존성 추가: 없음 (기존 모듈만 사용)

### 병렬 실행 가능 태스크
- T1 (types) 완료 후: T2, T3, T4 병렬 가능
- T2+T3+T4 완료 후: T5, T6 병렬 가능
- T5+T6 완료 후: T7, T8, T9, T10 병렬 가능
