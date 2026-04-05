# SPEC-LEARNWIRE-002 리서치

## 기존 코드 분석

### Pipeline 패키지 구조

- `pkg/pipeline/engine.go`: `SubprocessEngine` — `EngineConfig`로 설정, `Run()` 메서드가 5-phase 파이프라인 실행. 현재 gate 평가 없이 단순 순차 실행.
- `pkg/pipeline/runner.go`: `SequentialRunner` — `runPhaseWithRetry()`에서 `EvaluateGate()` 호출 후 VerdictFail 시 retry. `ParallelRunner` — gate 평가 후 결과 저장. 두 runner 모두 `PhaseBackend` 인터페이스 의존.
- `pkg/pipeline/phase_gate.go`: `EvaluateGate()` — GateValidation은 "PASS" 토큰 확인, GateReview는 "APPROVE" 토큰 확인.
- `pkg/pipeline/phase.go`: 5-phase 정의 (plan, test_scaffold, implement, validate, review). DefaultPhases()에서 Gate 필드가 있으나 현재 GateNone 기본값.

### Learn 패키지 구조

- `pkg/learn/store.go`: `NewStore(dir)` — `.autopus/learnings/` 디렉토리 생성, `pipeline.jsonl` 파일 관리. `Append()`, `Read()`, `NextID()` 메서드.
- `pkg/learn/record.go`: `RecordGateFail()`, `RecordCoverageGap()`, `RecordReviewIssue()`, `RecordExecutorError()`, `RecordFixPattern()` — 모두 `*Store`와 `RecordOpts`를 인자로 받음. store가 nil이면 에러 반환.
- `pkg/learn/types.go`: `LearningEntry` 구조체, `EntryType` 상수 5종, `Severity` 4단계, `RecordOpts` 구조체.

### 주요 삽입 지점

1. **SequentialRunner.runPhaseWithRetry()** (runner.go:64-87)
   - L78: `verdict := EvaluateGate(phase.Gate, resp.Output)` — VerdictFail 시 hook 삽입
   - L83-84: max retries 초과 시 critical severity hook 삽입

2. **ParallelRunner.RunPhases()** (runner.go:102-136)
   - L124: `verdict := EvaluateGate(ph.Gate, resp.Output)` — VerdictFail 시 hook 삽입
   - L119-120: Execute 에러 시 executor error hook 삽입

3. **SubprocessEngine.Run()** (engine.go:87-134)
   - L122: `resp, err := e.cfg.Backend.Execute(ctx, req)` — 에러 시 executor error hook

## 설계 결정

### D1: learn_hook.go 분리 파일 방식

**결정**: learn 호출 로직을 `learn_hook.go`로 분리

**근거**:
- runner.go (164줄)에 직접 추가하면 200줄 경계에 근접하여 file-size-limit 규칙 위반 위험
- learn 관련 로직을 한 곳에 모아 응집도 향상
- nil 체크를 hook 내부에 캡슐화하여 호출 지점 코드 단순화

**대안 검토**:
- runner.go에 직접 삽입: 단순하지만 파일 크기 증가 및 관심사 혼재
- 미들웨어/래퍼 패턴: PhaseBackend를 감싸는 래퍼로 투명하게 처리 가능하나, gate 평가 결과에 접근하려면 래퍼 계층이 복잡해짐

### D2: Runner 구조체에 store 주입

**결정**: `SequentialRunner`와 `ParallelRunner` 구조체에 `learnStore *learn.Store` 필드 추가, 생성자에서 옵셔널 주입

**근거**:
- 기존 `PhaseBackend` 주입 패턴과 일관성
- nil 기본값으로 opt-in 동작 자연스럽게 구현
- EngineConfig.LearnStore를 통해 SubprocessEngine에서도 전달 가능

**대안 검토**:
- context.Context에 store 삽입: Go context 안티패턴 (optional dependency를 context에 넣는 것은 비권장)
- 글로벌 싱글턴: 테스트 격리 불가, 동시성 문제

### D3: record.go의 nil store 에러 처리

**현재 상태**: `recordEntry()`는 store가 nil이면 에러를 반환한다.

**결정**: learn_hook.go의 헬퍼에서 store nil 체크를 먼저 수행하여 `record.go`를 호출하기 전에 조기 반환. record.go 자체는 수정하지 않는다.

**근거**: 기존 API 계약을 변경하지 않으면서, pipeline 통합 레이어에서 opt-in 동작을 보장

### D4: learn 기록 실패의 비치명적 처리

**결정**: learn.Record*() 호출이 에러를 반환해도, 파이프라인 실행에는 영향을 주지 않는다. 에러는 로깅만 하고 무시.

**근거**: SPEC-LEARN-001 원칙 — "학습은 보조 기능이며 핵심 파이프라인을 방해해서는 안 됨"
