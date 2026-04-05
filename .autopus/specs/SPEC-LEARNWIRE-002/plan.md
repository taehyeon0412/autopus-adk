# SPEC-LEARNWIRE-002 구현 계획

## 태스크 목록

### Phase 1: 인프라 (병렬 가능)

- [ ] T1: learn.Store AppendAtomic — sync.Mutex + AppendAtomic() 메서드, recordEntry() 수정 (store.go, record.go, R8)
- [ ] T2: DefaultPhases() gate 할당 — validate→GateValidate, review→GateReview (phase.go, R10)
- [ ] T3: RunConfig 확장 — LearnStore, CoverageThreshold 필드 추가 (runner.go, R1)

### Phase 2: Hook 구현 (T1 이후)

- [ ] T4: learn_hook.go 신규 — nil-safe wrapper 함수 4개 + coverage 파싱 + review issue 파싱 (R2-R7, D6)

### Phase 3: 와이어링 (T2+T3+T4 이후)

- [ ] T5: SequentialRunner 와이어링 — runPhaseWithRetry()에 learnHookGateFail/learnHookExecutorError 삽입 (runner.go, R2/R5/R6)
- [ ] T6: ParallelRunner 와이어링 — RunPhases() goroutine에 hook 삽입 (runner.go, R9)
- [ ] T7: runPipeline() Store 초기화 — .autopus/learnings/ 존재 확인, RunConfig 전달 (pipeline_run.go, R1/D4)

### Phase 4: 테스트

- [ ] T8: learn_hook_test.go — nil guard, coverage 파싱, review issue 파싱, 각 hook 호출 검증
- [ ] T9: store_test.go — AppendAtomic 동시성 테스트 (go test -race)
- [ ] T10: runner 통합 테스트 — gate fail → learn 기록 확인, nil store → 동작 불변

## 의존성

- T1, T2, T3: 독립, 병렬 가능 (파일 소유권 분리: store.go/record.go, phase.go, runner.go)
- T4: T1 완료 후 (AppendAtomic API 필요)
- T5, T6: T2 + T3 + T4 완료 후
- T7: T3 완료 후 (RunConfig 필드 필요)
- T8-T10: 대응 구현 완료 후

## 구현 전략

### 핵심 와이어링 포인트 (정확한 코드 위치)

**SequentialRunner.runPhaseWithRetry()** (runner.go:64):
```go
// line 73-86: retry loop
for attempt := 0; ; attempt++ {
    resp, err := r.backend.Execute(ctx, req)
    if err != nil {
        learnHookExecutorError(cfg.LearnStore, phase.ID, err)  // R5: 여기 삽입
        return PhaseResult{}, fmt.Errorf(...)
    }

    verdict := EvaluateGate(phase.Gate, resp.Output)
    if verdict == VerdictPass || phase.Gate == GateNone {
        return PhaseResult{...}, nil
    }

    learnHookGateFail(cfg.LearnStore, phase.ID, phase.Gate, resp.Output, attempt)  // R2: 여기 삽입

    if attempt >= maxRetries {
        learnHookGateFail(cfg.LearnStore, phase.ID, phase.Gate, resp.Output, attempt)  // R6: severity=critical
        return PhaseResult{}, fmt.Errorf(...)
    }
}
```

**ParallelRunner.RunPhases()** (runner.go:115):
```go
go func(idx int, ph Phase) {
    defer wg.Done()
    <-gate
    resp, err := r.backend.Execute(ctx, req)
    if err != nil {
        learnHookExecutorError(cfg.LearnStore, ph.ID, err)  // R5: 여기 삽입
        errs[idx] = ...
        return
    }
    verdict := EvaluateGate(ph.Gate, resp.Output)
    if verdict != VerdictPass && ph.Gate != GateNone {
        learnHookGateFail(cfg.LearnStore, ph.ID, ph.Gate, resp.Output, 0)  // R9: 여기 삽입
    }
    results[idx] = PhaseResult{...}
}(i, phase)
```

### 시그니처 변경

- `runPhaseWithRetry(ctx, phase, previousOutput)` → `runPhaseWithRetry(ctx, phase, previousOutput, cfg RunConfig)`
- `RunPhases` 시그니처는 불변 (이미 `cfg RunConfig` 받음)

### 파일 크기 제한

- runner.go 현재 164줄 → +25줄 (hook 호출 삽입) = ~189줄 ✓
- learn_hook.go 신규 ~130줄 ✓
- store.go 현재 ~90줄 → +20줄 (mutex + AppendAtomic) = ~110줄 ✓
- record.go 현재 60줄 → -10줄/+5줄 (recordEntry 수정) = ~55줄 ✓
- phase.go 현재 ~80줄 → +5줄 = ~85줄 ✓
