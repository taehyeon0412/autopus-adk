# SPEC-ROUTE-001 리서치

## 기존 코드 분석

### adapter 패키지 구조
- `pkg/worker/adapter/interface.go` — TaskConfig 구조체 (L22-29): TaskID, SessionID, Prompt, MCPConfig, WorkDir, EnvVars 필드. **Model 필드 없음**.
- `pkg/worker/adapter/interface.go` — ProviderAdapter 인터페이스 (L10-19): BuildCommand(ctx, TaskConfig) 시그니처. 인터페이스 변경 불필요.
- `pkg/worker/adapter/registry.go` — Registry.Get(name): 이름 기반 어댑터 조회. 모델 전환과 무관.

### 각 어댑터의 BuildCommand
- `pkg/worker/adapter/claude.go:24-56` — `exec.CommandContext(ctx, "claude", args...)`. args에 `--model` 플래그 없음.
- `pkg/worker/adapter/codex.go:22-46` — `exec.CommandContext(ctx, "codex", args...)`. `--model` 플래그 없음.
- `pkg/worker/adapter/gemini.go:22-48` — `exec.CommandContext(ctx, "gemini", args...)`. `--model` 플래그 없음.

### TaskConfig 사용처 (통합 지점)
- `pkg/worker/loop_exec.go:17` — `WorkerLoop.executeSubprocess(ctx, taskCfg)`: TaskConfig를 생성하여 BuildCommand에 전달. **Router 삽입 지점**.
- `pkg/worker/pipeline.go:98-107` — `PipelineExecutor.runPhase()`: TaskConfig를 생성하여 BuildCommand에 전달. **Router 삽입 지점**.
- `pkg/worker/context.go` — ContextBuilder.Build(): 프롬프트 조립만 담당, TaskConfig와 직접 관련 없음.

### 기존 테스트
- `pkg/worker/adapter/claude_test.go` — BuildCommand 결과 검증. Model 플래그 테스트 케이스 추가 필요.
- `pkg/worker/adapter/codex_test.go` — 동일 패턴.
- `pkg/worker/adapter/gemini_test.go` — 동일 패턴.

## 설계 결정

### D1: TaskConfig.Model 필드 추가 (선행 작업)
**결정**: ProviderAdapter 인터페이스를 변경하지 않고, TaskConfig에 `Model string` 필드만 추가한다.

**이유**: 인터페이스 변경은 모든 어댑터 구현체에 영향을 주지만, TaskConfig 필드 추가는 하위 호환성을 완전히 유지한다. 빈 문자열이면 기존 동작.

**대안 검토**:
- (A) BuildCommand 시그니처에 model 파라미터 추가 → 인터페이스 breaking change, 모든 구현체 수정 필요. 기각.
- (B) 별도 ModelOverrideAdapter 래퍼 → 과도한 추상화. 기각.

### D2: 분류기는 룰 기반 (ML 아님)
**결정**: Hermes 패턴을 따라 문자 수, 코드 블록 유무, 키워드 매칭 기반의 룰 기반 분류기를 구현한다.

**이유**: Worker 메시지는 구조화된 프롬프트(ContextBuilder가 생성)이므로 패턴이 예측 가능하다. ML 분류기는 학습 데이터와 추론 비용이 필요하여 이 단계에서는 과도하다.

**대안 검토**:
- (A) LLM 기반 분류 (메시지를 소형 모델에 보내 복잡도 판단) → 추가 API 호출 비용 발생, 레이턴시 증가. Phase 4에서 재검토.
- (B) 고정 규칙 없이 항상 complex → 현재 동작과 동일, 최적화 효과 없음.

### D3: 기본 비활성화
**결정**: RoutingConfig.Enabled = false를 기본값으로 한다. 명시적 활성화 필요.

**이유**: BS-021에서 지적한 대로 Worker 태스크는 대부분 complex일 수 있다. 실측 없이 기본 활성화하면 품질 저하 리스크가 있다. 먼저 메트릭을 수집하고 효과를 검증한 후 기본 활성화를 검토한다.

### D4: 라우팅 패키지를 adapter 밖에 배치
**결정**: `pkg/worker/routing/`에 별도 패키지로 생성한다.

**이유**: adapter 패키지는 프로바이더 CLI 실행이라는 단일 책임을 가진다. 복잡도 분석과 모델 선택은 별도 관심사이므로 독립 패키지가 적절하다. adapter → routing 의존은 없고, routing → adapter (TaskConfig 타입 참조)만 단방향이다. 실제로는 routing 패키지가 model string만 반환하므로 adapter 의존도 불필요할 수 있다.

## 참고 자료

### Hermes smart_model_routing.py 패턴
- 문자 수 임계값: 160자 (simple/complex 경계)
- 단어 수 임계값: 28단어
- 코드 블록 감지: ``` 패턴
- URL 포함 여부
- 복잡 키워드 목록 매칭

### Autopus Backend ReasoningEffort
- `low/medium/high/xhigh/max` 5단계
- 라우팅 시스템의 복잡도 레벨(3단계)과 1:1 매핑은 아니지만, Backend가 ReasoningEffort를 전달하면 라우팅 힌트로 활용 가능 (향후 확장)

### 비용 비율 추정
| Tier | 비용 비율 | 예시 모델 |
|------|----------|----------|
| simple | 1x | haiku, gpt-4o-mini, gemini-flash |
| medium | 5x | sonnet, gpt-4o, gemini-pro |
| complex | 20x | opus, o3, gemini-pro-max |

BS-021 주의사항: Worker 태스크의 simple turn 비율이 낮을 수 있음. Backend 에이전트 통신(채널 메시지, 짧은 응답)에 적용하면 효과가 더 클 수 있음.
