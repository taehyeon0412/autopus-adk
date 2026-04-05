# Review: SPEC-LEARNWIRE-002

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-04-05 23:22:56

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| codex | critical | REQ-002, REQ-005, REQ-006, and effectively REQ-009 are attached to the wrong execution layer. Current [`SubprocessEngine.Run()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/engine.go):87 only loops phases, calls `Backend.Execute()`, and returns raw output; it does not evaluate gates or run retries. Gate evaluation and retry live in [`SequentialRunner.runPhaseWithRetry()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/runner.go):64. As written, the spec cannot be implemented without either refactoring `SubprocessEngine.Run()` to delegate to the runner or moving the learn hooks into the runner. |
| codex | major | REQ-001 is incomplete for an “automatic integration wiring” spec. Adding `LearnStore *learn.Store` to [`EngineConfig`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/engine.go):42 is not enough, because the only production caller currently building that config, [`runPipeline()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/internal/cli/pipeline_run.go):103, never creates or injects a store. The spec needs an explicit requirement for where the store is created and who owns that initialization. |
| codex | major | REQ-003 assumes coverage is available as structured phase result data, but [`PhaseResponse`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/engine.go):36 only contains `Output string`. There is no defined parser for `go test -cover` output, no declared source of the threshold, and no contract for deriving `packages`. The requirement is underspecified and not directly implementable. |
| codex | major | REQ-004 is underspecified for issue extraction. The current review gate only reduces output to pass/fail via [`EvaluateGate()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/phase_gate.go):30. Parsing “individual issue”, `severity`, and `specID` from arbitrary `REQUEST_CHANGES` text has no defined format. If structured parsing is intended, the spec should mandate an output grammar or explicitly reuse the existing `FINDING:` parser in [`pkg/spec/reviewer.go`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/spec/reviewer.go):16. |
| codex | major | REQ-006 is internally inconsistent. It says to call `learnHookGateFail(store, phaseID, gateType, output, attempt)` “with severity critical”, but severity is not part of that signature. The spec needs to define whether severity is an extra argument, derived internally by the hook, or handled by a separate final-failure hook. |
| codex | major | REQ-008 cannot be satisfied by only adding a mutex to `learn.Store`. The race exists across the two-call sequence in [`recordEntry()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/learn/record.go):8, which does `NextID()` and then `Append()` separately. Locking each store method independently does not make the pair atomic; the spec needs a combined API such as `AppendWithNextID(...)` or it must explicitly move the lock to the caller path. |
| codex | minor | REQ-007 should clarify scope. Current [`recordEntry()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/learn/record.go):8 returns an error on nil store, and tests currently assert that behavior. Saying “all functions in learn_hook.go must nil-guard and emit no error” is fine for new wrapper hooks, but the spec should state that this does not silently change the general `pkg/learn` API contract. |

## Provider Responses

### Response 1

VERDICT: REVISE

FINDING: [critical] REQ-002, REQ-005, REQ-006, and effectively REQ-009 are attached to the wrong execution layer. Current [`SubprocessEngine.Run()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/engine.go):87 only loops phases, calls `Backend.Execute()`, and returns raw output; it does not evaluate gates or run retries. Gate evaluation and retry live in [`SequentialRunner.runPhaseWithRetry()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/runner.go):64. As written, the spec cannot be implemented without either refactoring `SubprocessEngine.Run()` to delegate to the runner or moving the learn hooks into the runner.

FINDING: [major] REQ-001 is incomplete for an “automatic integration wiring” spec. Adding `LearnStore *learn.Store` to [`EngineConfig`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/engine.go):42 is not enough, because the only production caller currently building that config, [`runPipeline()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/internal/cli/pipeline_run.go):103, never creates or injects a store. The spec needs an explicit requirement for where the store is created and who owns that initialization.

FINDING: [major] REQ-003 assumes coverage is available as structured phase result data, but [`PhaseResponse`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/engine.go):36 only contains `Output string`. There is no defined parser for `go test -cover` output, no declared source of the threshold, and no contract for deriving `packages`. The requirement is underspecified and not directly implementable.

FINDING: [major] REQ-004 is underspecified for issue extraction. The current review gate only reduces output to pass/fail via [`EvaluateGate()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/pipeline/phase_gate.go):30. Parsing “individual issue”, `severity`, and `specID` from arbitrary `REQUEST_CHANGES` text has no defined format. If structured parsing is intended, the spec should mandate an output grammar or explicitly reuse the existing `FINDING:` parser in [`pkg/spec/reviewer.go`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/spec/reviewer.go):16.

FINDING: [major] REQ-006 is internally inconsistent. It says to call `learnHookGateFail(store, phaseID, gateType, output, attempt)` “with severity critical”, but severity is not part of that signature. The spec needs to define whether severity is an extra argument, derived internally by the hook, or handled by a separate final-failure hook.

FINDING: [major] REQ-008 cannot be satisfied by only adding a mutex to `learn.Store`. The race exists across the two-call sequence in [`recordEntry()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/learn/record.go):8, which does `NextID()` and then `Append()` separately. Locking each store method independently does not make the pair atomic; the spec needs a combined API such as `AppendWithNextID(...)` or it must explicitly move the lock to the caller path.

FINDING: [minor] REQ-007 should clarify scope. Current [`recordEntry()`](/Users/bitgapnam/Documents/github/autopus-co/autopus-adk/pkg/learn/record.go):8 returns an error on nil store, and tests currently assert that behavior. Saying “all functions in learn_hook.go must nil-guard and emit no error” is fine for new wrapper hooks, but the spec should state that this does not silently change the general `pkg/learn` API contract.

The concept is feasible, but the spec is not implementation-ready yet. The main problem is architectural mismatch: the document places learn-hook behavior in `SubprocessEngine.Run()` even though gate and retry semantics currently live elsewhere. After that, the missing contracts for store initialization, coverage parsing, structured review parsing, and atomic ID allocation need to be tightened before implementation can be done correctly without guesswork.


