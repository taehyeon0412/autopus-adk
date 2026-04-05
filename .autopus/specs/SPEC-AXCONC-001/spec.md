# SPEC-AXCONC-001: Orchestra Concurrency Improvements

**Status**: completed
**Created**: 2026-04-05
**Domain**: AXCONC

## 목적

`pkg/orchestra` 패키지의 동시성 패턴에서 @AX:WARN 어노테이션으로 식별된 두 가지 문제를 해결한다.

1. `runParallel()`에서 모든 고루틴이 단일 context를 공유하여 개별 프로바이더 취소가 불가능
2. `WaitForCompletion()`에서 caller가 deadline을 설정하지 않으면 무한 블로킹 발생

이 개선은 프로바이더 하나가 hang될 때 전체 오케스트레이션이 영향받지 않는 graceful degradation을 가능하게 한다.

## 요구사항

### R1 (P0): Per-Goroutine Context in runParallel

WHEN `runParallel()` launches goroutines for each provider,
THE SYSTEM SHALL derive a per-goroutine context via `context.WithCancel(ctx)` so that individual providers can be cancelled independently without affecting other goroutines.

### R2 (P0): Per-Provider Timeout Cancellation

WHEN a provider goroutine exceeds the per-provider timeout (derived from `OrchestraConfig.TimeoutSeconds / providers` or a configurable per-provider value),
THE SYSTEM SHALL cancel only that provider's derived context and record it as a failed provider with a timeout error, while allowing remaining providers to continue.

### R3 (P1): Safety Deadline in WaitForCompletion

WHEN `WaitForCompletion()` is called with a context that has no deadline set,
THE SYSTEM SHALL enforce an internal safety deadline of 10 minutes as a fallback to prevent indefinite blocking.

### R4 (P1): Safety Deadline Warning Log

WHEN the internal safety deadline fallback activates in `WaitForCompletion()`,
THE SYSTEM SHALL log a warning message indicating that the caller did not set a deadline and the fallback was used.

### R5 (P0): Backward Compatibility

WHERE the existing `runParallel()` and `WaitForCompletion()` public interfaces are used,
THE SYSTEM SHALL maintain the same function signatures and return types to avoid breaking callers.

## 생성 파일 상세

| 파일 | 역할 |
|------|------|
| `pkg/orchestra/runner.go` | `runParallel()` 수정: per-goroutine context 도입 |
| `pkg/orchestra/completion_poll.go` | `WaitForCompletion()` 수정: safety deadline 추가 |
| `pkg/orchestra/runner_test.go` 또는 신규 테스트 | R1, R2 검증 테스트 |
| `pkg/orchestra/completion_poll_test.go` 또는 신규 테스트 | R3, R4 검증 테스트 |
