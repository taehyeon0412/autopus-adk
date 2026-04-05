# SPEC-AXCONC-001 리서치

## 기존 코드 분석

### Issue 1: runParallel() — 공유 context 문제

**파일**: `pkg/orchestra/runner.go:122-138`

현재 `runParallel()`은 모든 고루틴에 동일한 `ctx`를 전달한다:

```go
go func(idx int, provider ProviderConfig) {
    defer wg.Done()
    resp, err := runProvider(ctx, provider, cfg.Prompt)  // 공유 ctx
```

이로 인해:
- 한 프로바이더가 hang되어도 개별 취소 불가 — 공유 ctx를 cancel하면 전체가 취소됨
- `runFastest()`는 이미 `context.WithCancel(ctx)`를 사용 중 (runner.go:189) — 동일 패턴 적용 가능

**호출자**: `RunOrchestra()` consensus 분기 (runner.go:64)에서만 호출. `timeoutCtx`가 전달됨.

**runProvider()** (runner.go:229): `exec.CommandContext`를 사용하므로 context 취소 시 프로세스가 kill됨. per-goroutine context는 이 메커니즘과 자연스럽게 호환된다.

### Issue 2: WaitForCompletion() — 무한 블로킹 문제

**파일**: `pkg/orchestra/completion_poll.go:29-92`

`WaitForCompletion()`은 `for { select { case <-ctx.Done(): ... } }` 루프로 동작한다.
caller가 deadline 없는 context를 전달하면 영원히 블로킹된다.

**호출 경로 분석**:
- `CompletionDetector` 인터페이스 (completion_detector.go:11)를 통해 호출
- 구현체: `ScreenPollDetector`, `SignalDetector`, `FileIPCDetector`
- `ScreenPollDetector`만 poll 루프를 가지므로 이 구현체가 주 대상
- `SignalDetector`는 signal channel 대기, `FileIPCDetector`는 file poll — 둘 다 동일한 ctx.Done() 의존

**caller 분석**: interactive pane 흐름에서 `waitForCompletion()` wrapper를 통해 호출됨.
보통 timeout이 설정된 context가 전달되지만, 방어적 코딩으로 safety deadline 추가가 바람직하다.

### 기존 테스트 패턴

- `runner_test.go`: `echoProvider()`, `sleepProvider()` 헬퍼 사용
- `completion_poll_test.go`: `countingScreenMock`, `newPlainMock()` mock 사용
- 테스트 프레임워크: `testify/assert`, `testify/require`
- 모든 테스트에 `t.Parallel()` 사용

### 관련 파일 목록

| 파일 | 역할 | 변경 여부 |
|------|------|----------|
| `pkg/orchestra/runner.go` | runParallel, runProvider | 수정 |
| `pkg/orchestra/completion_poll.go` | ScreenPollDetector.WaitForCompletion | 수정 |
| `pkg/orchestra/completion_detector.go` | CompletionDetector 인터페이스 | 변경 없음 |
| `pkg/orchestra/types.go` | ProviderConfig, OrchestraConfig | 변경 없음 |
| `pkg/orchestra/runner_test.go` | runParallel 테스트 | 추가 |
| `pkg/orchestra/completion_poll_test.go` | WaitForCompletion 테스트 | 추가 |

## 설계 결정

### Decision 1: context.WithTimeout vs context.WithCancel for per-goroutine

**선택**: `context.WithTimeout(ctx, perProviderTimeout)`

**이유**: per-goroutine cancel만 하면 timeout 누가 트리거할지 별도 goroutine이 필요하다. `WithTimeout`이면 Go runtime이 자동으로 처리하므로 코드가 단순하다. 부모 ctx 취소 시 자식도 자동 취소되므로 기존 동작과 호환된다.

**대안**: `WithCancel` + 별도 timer goroutine — 불필요한 복잡성.

### Decision 2: Safety deadline 값 10분

**선택**: 10분 (hardcoded constant)

**이유**: `RunOrchestra()`의 기본 timeout이 120초(2분)이고, interactive pane 모드는 더 길게 실행될 수 있다. 10분은 합리적 AI 모델 응답 시간의 상한이며, 실제 무한 블로킹만 방지하면 된다.

**대안 1**: configurable via OrchestraConfig — 인터페이스 변경 필요하여 R5 위반.
**대안 2**: 5분 — interactive debate 모드에서 너무 짧을 수 있음.

### Decision 3: Safety deadline을 ScreenPollDetector에만 적용

**선택**: `ScreenPollDetector.WaitForCompletion()`에만 적용

**이유**: `SignalDetector`는 signal channel 대기이므로 poll 루프와 성격이 다르다. `FileIPCDetector`도 file 존재 여부 poll이지만, hook session이 관리하므로 별도 safety가 있다. @AX:WARN이 `completion_poll.go`에만 달려있으므로 scope를 좁힌다.

**대안**: CompletionDetector 인터페이스 레벨에서 wrapper 적용 — overengineering.
