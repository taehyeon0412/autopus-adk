# SPEC-LEARNWIRE-002: Pipeline-Learn 자동 통합 와이어링

**Status**: completed
**Created**: 2026-04-05
**Domain**: LEARNWIRE
**Module**: autopus-adk
**Revision**: 2

## 목적

`pkg/learn/` 패키지에 학습 인프라(Store, Record*, Query, Prune, Summary)가 완전히 구현되어 있으나, `pkg/pipeline/`에서 `pkg/learn`을 한 번도 import하지 않아 파이프라인 gate 실패 시 자동 학습 기록이 동작하지 않는다.

SPEC-LEARN-001의 설계 의도를 완성하기 위해, SequentialRunner/ParallelRunner의 gate/phase 실패 지점에서 `learn.Record*()` 함수를 자동 호출하는 와이어링을 구현한다.

## 핵심 설계 결정

### D1: 와이어링 대상 — SequentialRunner/ParallelRunner (runner 레벨)
`SubprocessEngine.Run()` (engine.go:87)는 gate 평가/retry가 없는 단순 순차 루프. gate 평가(`EvaluateGate`)와 retry는 `SequentialRunner.runPhaseWithRetry()` (runner.go:64)에 존재. learn hook은 runner 레벨에 삽입.

- `SequentialRunner.runPhaseWithRetry()` line 78: `verdict := EvaluateGate(phase.Gate, resp.Output)` → VerdictFail 시 hook 호출
- `ParallelRunner.RunPhases()` line 123: `verdict := EvaluateGate(ph.Gate, resp.Output)` → VerdictFail 시 hook 호출
- `SubprocessEngine.Run()`은 변경하지 않음 (gate 없는 단순 루프)

### D2: learn.Store 동시성 안전 — AppendAtomic API
현재 `recordEntry()` (record.go:8)이 `NextID()` + `Append()`를 별도 호출하여 race condition 존재. 개별 메서드에 mutex를 걸어도 두 호출 사이 race는 해결 불가. 해결: `Store.AppendAtomic(entry LearningEntry) error` 신규 메서드 — 내부에서 mutex lock → NextID → entry.ID 할당 → Append → unlock을 단일 트랜잭션으로 수행. 기존 `recordEntry()`가 이 메서드를 사용하도록 수정.

### D3: nil-store 규칙 — hook 레이어에서만 nil guard
기존 `learn.Record*()` 함수는 nil store 시 error 반환 (record.go:10). 이 동작은 변경하지 않음 (기존 API 계약 유지). 대신 `learn_hook.go`의 wrapper 함수가 호출 전에 nil guard: `if store == nil { return }`. `pkg/learn` 패키지의 공개 API는 불변.

### D4: Store 초기화 — runPipeline()에서 조건부 생성
`runPipeline()` (pipeline_run.go:103)에서 `.autopus/learnings/` 디렉토리 존재 여부 확인:
- 존재 → `learn.NewStore(dir)` → RunConfig에 전달
- 미존재 → nil (학습 비활성화, `NewStore()`가 `MkdirAll`로 디렉토리를 자동 생성하므로 이를 피하기 위해 사전 확인 필요)

### D5: 기록 정책 — per-attempt + exhaustion
매 retry 시도마다 gate_fail 기록 (attempt 번호 포함). max retry 초과 시 severity=critical 별도 기록. 중복은 의도적 — retry 진행 추적용.

### D6: 출력 파싱 — PhaseResponse.Output 기반
`PhaseResponse.Output` (engine.go:38)은 단순 string. coverage 값과 review issue는 이 string에서 정규식 파싱:
- Coverage: `coverage: (\d+\.?\d*)% of statements` 정규식
- Review issues: `pkg/spec/reviewer.go`의 `findingRe` 재사용 (`FINDING: \[(\w+)\] (.+)`)
- 파싱 실패 시 graceful degradation — 전체 output을 단일 pattern으로 기록

## 요구사항

### R1: Learn Store 의존성 주입 (RunConfig)

WHEN `SequentialRunner.RunPhases()` or `ParallelRunner.RunPhases()` is called, THE SYSTEM SHALL accept an optional `LearnStore *learn.Store` field in `RunConfig` (runner.go:14). 초기화:
- `runPipeline()` (pipeline_run.go)에서 `os.Stat(".autopus/learnings")` → 존재 시 `learn.NewStore(dir)` → `RunConfig.LearnStore`
- 미존재 시 nil → 학습 비활성화

### R2: Gate Failure 자동 기록 (SequentialRunner)

WHEN `SequentialRunner.runPhaseWithRetry()` (runner.go:78)에서 `EvaluateGate()` returns VerdictFail, THE SYSTEM SHALL call `learnHookGateFail(store, phaseID, gateType, output, attempt)`. per-attempt 기록. 호출 지점: runner.go:82 (retry 루프 내, `attempt >= maxRetries` 체크 전).

### R3: Coverage Gap 자동 기록

WHEN R2의 gate fail 출력에서 `coverage: (\d+\.?\d*)% of statements` 패턴이 매칭되고 값이 `RunConfig.CoverageThreshold` (기본 85.0) 미만이면, THE SYSTEM SHALL `learnHookCoverageGap(store, coverage, threshold, phaseID)` 추가 호출. `CoverageThreshold` 필드를 RunConfig에 추가.

### R4: Review Issue 개별 기록

WHEN R2의 gate fail 출력에서 `REQUEST_CHANGES` 패턴이 매칭되면, THE SYSTEM SHALL `FINDING: \[(\w+)\] (.+)` 정규식으로 개별 issue를 추출하고 각각 `learnHookReviewIssue(store, issueDesc, severity, specID)` 호출. 정규식 매칭 실패 시 전체 output을 단일 issue로 기록 (graceful degradation).

### R5: Executor Error 자동 기록

WHEN `backend.Execute()` returns a Go error (runner.go:74, runner.go:119), THE SYSTEM SHALL `learnHookExecutorError(store, phaseID, err)` 호출 before propagating the error.

### R6: Max Retry 초과 기록

WHEN `SequentialRunner.runPhaseWithRetry()` (runner.go:83)에서 `attempt >= maxRetries`, THE SYSTEM SHALL `learnHookGateFail(store, phaseID, gateType, output, attempt)` with severity "critical" 호출. 이는 R2 per-attempt 기록과 별개의 최종 exhaustion 기록.

### R7: Opt-in 동작 보장 (Zero-Impact nil)

WHEN `RunConfig.LearnStore` is nil, THE SYSTEM SHALL NOT alter any existing pipeline behavior. `learn_hook.go`의 모든 함수 첫 줄: `if store == nil { return }`. 에러/경고/로그 없음. 기존 `pkg/learn` 패키지의 `Record*()` nil 에러 반환 동작은 변경하지 않음 (R7은 hook 레이어에만 적용).

### R8: learn.Store 동시성 안전 (AppendAtomic)

THE SYSTEM SHALL add `AppendAtomic(entryType EntryType, opts RecordOpts) error` method to `learn.Store`:
1. `sync.Mutex` lock 획득
2. `NextID()` → ID 할당
3. `LearningEntry` 생성 + `Append()`
4. unlock
기존 `recordEntry()` (record.go:8)가 `AppendAtomic()` 사용하도록 수정. 기존 `Record*()` 공개 API 시그니처는 불변.

### R9: ParallelRunner Gate Failure 기록

WHEN `ParallelRunner.RunPhases()` (runner.go:123)에서 `EvaluateGate()` returns VerdictFail in a goroutine, THE SYSTEM SHALL `learnHookGateFail(store, phaseID, gateType, output, 0)` 호출. R8의 `AppendAtomic`이 동시 goroutine에서의 안전을 보장.

### R10: Gate 할당 (전제 조건)

THE SYSTEM SHALL update `DefaultPhases()` in `pkg/pipeline/phase.go` to assign `GateValidate` to the validate phase and `GateReview` to the review phase. Unknown gate = pass 동작은 유지.

## 생성/수정 파일 상세

| 파일 | 역할 | 변경 유형 |
|------|------|-----------|
| `pkg/learn/store.go` | Store에 sync.Mutex + AppendAtomic 메서드 추가 (R8) | 수정 |
| `pkg/learn/record.go` | recordEntry()가 AppendAtomic() 사용하도록 수정 (R8) | 수정 |
| `pkg/pipeline/runner.go` | RunConfig에 LearnStore/CoverageThreshold 추가, runPhaseWithRetry/RunPhases에 hook 호출 삽입 (R1-R6, R9) | 수정 |
| `pkg/pipeline/learn_hook.go` (신규) | nil-safe wrapper: learnHookGateFail, learnHookCoverageGap, learnHookReviewIssue, learnHookExecutorError + 출력 파싱 함수 (R7, D6) | 신규 |
| `pkg/pipeline/phase.go` | DefaultPhases() gate 할당 (R10) | 수정 |
| `internal/cli/pipeline_run.go` | runPipeline()에서 learn.Store 조건부 생성 및 RunConfig 전달 (R1, D4) | 수정 |
| `pkg/pipeline/learn_hook_test.go` (신규) | hook nil guard, 출력 파싱, 각 Record* 연동 테스트 | 신규 |
| `pkg/learn/store_test.go` | AppendAtomic 동시성 테스트 추가 | 수정 |
