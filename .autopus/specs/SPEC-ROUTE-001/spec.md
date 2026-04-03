# SPEC-ROUTE-001: Smart Model Routing

**Status**: completed
**Created**: 2026-04-02
**Domain**: ROUTE
**Phase**: Phase 3 (v0.15~v0.16) — Optimization
**ICE Score**: 1.20 (Impact 5, Confidence 6, Ease 4)

## 목적

Worker가 모든 메시지에 동일한 고비용 모델을 사용하면 비용 비효율이 발생한다. 메시지 수준 복잡도를 분석하여 단순 작업(상태 확인, 짧은 응답)은 저비용 모델로, 복잡 작업(아키텍처 분석, 리팩토링)은 고비용 모델로 자동 라우팅하면 품질 손실 없이 비용을 절감할 수 있다.

현재 `adapter.TaskConfig`에 Model 필드가 없고, 각 어댑터의 `BuildCommand`는 고정 바이너리만 호출하며 모델 선택 파라미터를 지원하지 않는다. Registry도 런타임 모델 전환을 지원하지 않는다.

## 요구사항

### REQ-ROUTE-01: TaskConfig 모델 오버라이드
WHEN a TaskConfig is created with a non-empty Model field, THE SYSTEM SHALL pass that model to the provider CLI via the appropriate flag (e.g., `--model`).

### REQ-ROUTE-02: 어댑터 모델 플래그 지원
WHEN BuildCommand receives a TaskConfig with a Model value, THE SYSTEM SHALL append the model flag to the CLI arguments for each provider:
- claude: `--model {Model}`
- codex: `--model {Model}`
- gemini: `--model {Model}`

WHEN the Model field is empty, THE SYSTEM SHALL use the provider's default model (current behavior).

### REQ-ROUTE-03: 메시지 복잡도 분류
WHEN a message is submitted for classification, THE SYSTEM SHALL analyze it using the following signals and assign a complexity level (simple, medium, complex):

| Signal | Simple | Medium | Complex |
|--------|--------|--------|---------|
| Character count | < 200 | 200-1000 | > 1000 |
| Code blocks | None | Optional | Present |
| Keywords | "확인", "상태", "목록" | "수정", "추가", "변경" | "리팩토링", "분석", "아키텍처", "설계" |
| Tool hints | Read-only | Mixed | File modification, build |

### REQ-ROUTE-04: 복잡도 기반 모델 매핑
WHEN a message complexity is determined, THE SYSTEM SHALL map it to a model tier:

| Complexity | Claude | Codex/OpenAI | Gemini |
|------------|--------|--------------|--------|
| simple | haiku | gpt-4o-mini | gemini-flash |
| medium | sonnet | gpt-4o | gemini-pro |
| complex | opus | o3 | gemini-pro (max) |

### REQ-ROUTE-05: 라우팅 설정 오버라이드
WHEN a routing configuration is provided (via config file or API), THE SYSTEM SHALL use those mappings instead of the defaults. THE SYSTEM SHALL support per-provider model mappings and complexity threshold overrides.

### REQ-ROUTE-06: 라우팅 비활성화
WHEN smart routing is disabled (default for initial release), THE SYSTEM SHALL pass through all tasks without modification, preserving current behavior.

### REQ-ROUTE-07: 라우팅 메트릭 로깅
WHEN a model routing decision is made, THE SYSTEM SHALL log the complexity classification, selected model, and the signals that contributed to the decision.

## 생성 파일 상세

### 선행 작업 (adapter 패키지 확장)
- `pkg/worker/adapter/interface.go` — TaskConfig에 `Model string` 필드 추가
- `pkg/worker/adapter/claude.go` — BuildCommand에 `--model` 플래그 조건부 추가
- `pkg/worker/adapter/codex.go` — BuildCommand에 `--model` 플래그 조건부 추가
- `pkg/worker/adapter/gemini.go` — BuildCommand에 `--model` 플래그 조건부 추가

### 라우팅 패키지 (신규)
- `pkg/worker/routing/classifier.go` — MessageComplexityClassifier: 메시지 분석 및 복잡도 레벨 반환
- `pkg/worker/routing/router.go` — ProviderRouter: 복잡도 + 프로바이더 → 모델명 결정
- `pkg/worker/routing/config.go` — RoutingConfig: 매핑 규칙, 임계값, 활성화 플래그

### 테스트 파일
- `pkg/worker/routing/classifier_test.go`
- `pkg/worker/routing/router_test.go`
- `pkg/worker/routing/config_test.go`
- `pkg/worker/adapter/claude_test.go` (모델 플래그 테스트 추가)
- `pkg/worker/adapter/codex_test.go` (모델 플래그 테스트 추가)
- `pkg/worker/adapter/gemini_test.go` (모델 플래그 테스트 추가)

## 통합 지점

라우팅은 `WorkerLoop.executeSubprocess()` (loop_exec.go) 또는 `PipelineExecutor.runPhase()` (pipeline.go)에서 TaskConfig를 생성하는 시점에 삽입된다. Router가 TaskConfig.Model을 설정하면, 어댑터의 BuildCommand가 해당 모델 플래그를 CLI에 전달한다.
