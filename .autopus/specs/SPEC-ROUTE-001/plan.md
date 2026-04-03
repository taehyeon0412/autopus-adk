# SPEC-ROUTE-001 구현 계획

## 태스크 목록

### Phase A: 선행 작업 — Adapter 확장
- [ ] T1: TaskConfig에 Model 필드 추가 (`pkg/worker/adapter/interface.go`)
- [ ] T2: ClaudeAdapter.BuildCommand에 `--model` 조건부 플래그 추가 및 테스트
- [ ] T3: CodexAdapter.BuildCommand에 `--model` 조건부 플래그 추가 및 테스트
- [ ] T4: GeminiAdapter.BuildCommand에 `--model` 조건부 플래그 추가 및 테스트

### Phase B: 라우팅 패키지 구현
- [ ] T5: `pkg/worker/routing/config.go` — RoutingConfig 타입 및 기본값 정의
- [ ] T6: `pkg/worker/routing/classifier.go` — MessageComplexityClassifier 구현
- [ ] T7: `pkg/worker/routing/router.go` — ProviderRouter 구현 (classifier + config → model)
- [ ] T8: 라우팅 패키지 단위 테스트 (classifier_test, router_test, config_test)

### Phase C: 통합
- [ ] T9: WorkerLoop / PipelineExecutor에 Router 옵션 주입 및 TaskConfig.Model 설정
- [ ] T10: 통합 테스트 — 라우팅 활성화/비활성화 시나리오 검증

## 구현 전략

### 접근 방법
1. **최소 침습적 확장**: 기존 adapter 인터페이스를 변경하지 않고 TaskConfig에 필드만 추가한다. BuildCommand의 시그니처는 그대로 유지되므로 기존 코드에 영향이 없다.
2. **기본 비활성화**: RoutingConfig.Enabled = false가 기본값이므로 기존 동작이 보존된다. 명시적으로 활성화해야 라우팅이 적용된다.
3. **Hermes 패턴 참고**: 복잡도 분류 로직은 Hermes의 `smart_model_routing.py` 패턴(문자 수, 단어 수, 코드 블록, 키워드)을 Go로 포팅한다.

### 기존 코드 활용
- `adapter.TaskConfig` — Model 필드 추가 (1줄)
- 각 어댑터 `BuildCommand` — 4~5줄 조건문 추가
- `WorkerLoop.executeSubprocess()` — Router 호출 삽입 (3~5줄)
- `PipelineExecutor.runPhase()` — Router 호출 삽입 (3~5줄)

### 변경 범위
- 기존 파일 수정: 5개 (interface.go, claude.go, codex.go, gemini.go, loop_exec.go 또는 pipeline.go)
- 신규 파일 생성: 3개 소스 + 3개 테스트 = 6개
- 총 신규 코드 예상: ~400줄 (소스 ~200줄, 테스트 ~200줄)

### 의존성
- T1은 T2~T4의 선행 조건
- T5는 T6~T7의 선행 조건
- T9는 T1~T7 모두 완료 후 진행
- T2, T3, T4는 병렬 실행 가능
- T6, T7은 T5 이후 병렬 실행 가능
