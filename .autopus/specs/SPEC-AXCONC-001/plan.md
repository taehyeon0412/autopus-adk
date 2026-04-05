# SPEC-AXCONC-001 구현 계획

## 태스크 목록

- [ ] T1: `runParallel()` per-goroutine context 도입 (R1)
- [ ] T2: Per-provider timeout 취소 로직 추가 (R2)
- [ ] T3: `WaitForCompletion()` safety deadline 추가 (R3, R4)
- [ ] T4: `runParallel()` 변경에 대한 단위 테스트 작성 (R1, R2, R5)
- [ ] T5: `WaitForCompletion()` 변경에 대한 단위 테스트 작성 (R3, R4)
- [ ] T6: @AX:WARN 어노테이션 업데이트/제거

## 구현 전략

### T1: Per-Goroutine Context (runner.go)

현재 `runParallel()`의 for 루프 내부에서 `context.WithCancel(ctx)`로 파생 context를 생성한다.

```go
for i, p := range cfg.Providers {
    wg.Add(1)
    childCtx, childCancel := context.WithCancel(ctx)
    go func(idx int, provider ProviderConfig, cancel context.CancelFunc) {
        defer wg.Done()
        defer cancel()
        resp, err := runProvider(childCtx, provider, cfg.Prompt)
        // ...
    }(i, p, childCancel)
}
```

변경 범위: `runner.go` 122-138행, 약 10줄 변경.

### T2: Per-Provider Timeout (runner.go)

`ProviderConfig`에 이미 `StartupTimeout`이 있으나 실행 전체 timeout은 없다.
`runParallel()` 내에서 per-provider context에 timeout을 추가한다.
기본값: `OrchestraConfig.TimeoutSeconds`를 사용하되, `ProviderConfig`에 개별 timeout 필드가 있으면 그것을 우선한다.

per-provider context 체인: `ctx` → `WithTimeout(perProviderTimeout)` → `WithCancel()` (cancel은 부모 취소 시 자동 전파).

실제로는 `context.WithTimeout(ctx, timeout)`이 cancel도 포함하므로 단일 호출로 충분하다.

### T3: Safety Deadline (completion_poll.go)

`WaitForCompletion()` 진입 시 `ctx.Deadline()` 체크. deadline이 없으면:

```go
const defaultSafetyDeadline = 10 * time.Minute

func (d *ScreenPollDetector) WaitForCompletion(ctx context.Context, ...) (bool, error) {
    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, defaultSafetyDeadline)
        defer cancel()
        log.Printf("[WARN] WaitForCompletion called without deadline; using %v safety fallback", defaultSafetyDeadline)
    }
    // ... existing logic
}
```

변경 범위: `completion_poll.go` 30행 부근, 약 8줄 추가.

### T4-T5: TDD 접근

테스트를 먼저 작성하고 구현을 맞추는 TDD 방식:

1. T4 테스트: 한 프로바이더가 느릴 때 나머지가 정상 완료되는지 검증
2. T4 테스트: 개별 cancel 후 다른 고루틴 영향 없는지 검증
3. T5 테스트: deadline 없는 context로 호출 시 safety deadline 적용 확인
4. T5 테스트: deadline 있는 context로 호출 시 safety deadline 미적용 확인

### T6: AX 어노테이션 정리

구현 완료 후 해당 @AX:WARN 어노테이션을 @AX:RESOLVED로 변경하거나 제거한다.

## 의존성

- T1 → T2 (per-goroutine context가 있어야 per-provider timeout 가능)
- T4는 T1+T2 완료 후
- T3은 T1과 독립적으로 병렬 진행 가능
- T5는 T3 완료 후
- T6은 T1-T5 모두 완료 후
