# SPEC-REVCONV-001: SPEC Review 수렴성 보장 — 2-Phase Scoped Review

**Status**: completed
**Created**: 2026-04-05
**Domain**: REVCONV
**Ref**: BS-026
**Revision**: 1

## 목적

SPEC review의 REVISE 루프가 수렴하지 않는 문제를 해결한다. 현재 `BuildReviewPrompt`(pkg/spec/prompt.go:11)는 매번 동일한 open-ended 리뷰 지시를 생성하고, `ReviewFinding`(pkg/spec/types.go:69-74)에는 ID/Status 필드가 없어 finding 추적이 불가능하다. 리뷰어가 매 라운드마다 새 코드 경로를 탐색하면서 finding이 줄지 않고 늘어나는 근본 원인을 2-Phase 구조(Discovery → Verification)와 Finding 추적 메커니즘으로 해결한다.

## 배경

- `ParseVerdict`(pkg/spec/reviewer.go:18)는 provider 출력에서 finding을 blind append하며, 이전 라운드 finding과의 관계를 추적하지 않음
- `spec_review.go`(internal/cli/spec_review.go:41)의 `runSpecReview`는 revision 번호를 항상 0으로 전달하며 REVISE 루프 자체가 미구현
- 정적 분석 도구(lint, staticcheck)의 결과를 LLM 리뷰와 통합하지 않아, LLM이 매번 동일한 스타일/린트 이슈를 재발견

## 핵심 설계 결정

### D1: Finding ID는 리비전 독립 (Revision-Agnostic)
Finding ID 형식: `F-{seq}` (예: `F-001`, `F-002`). 리비전 번호는 ID에 포함하지 않으며, `FirstSeenRev` 메타데이터로 별도 기록. 이렇게 해야 cross-revision 추적이 가능하고, verify 모드에서 prior findings list와의 ID 비교가 성립함.

### D2: Verify 모드의 Regression Escape Hatch
verify 모드에서 발견된 새 finding 중 severity가 `critical` 또는 category가 `security`인 것은 `out_of_scope` 면제 — open finding으로 즉시 등록. 나머지 새 finding은 `out_of_scope`로 태깅하여 review.md에 기록하되 verdict 계산에서 제외. 이로써 수렴성과 안전성을 동시에 보장.

### D3: Prior Findings 저장소
각 revision의 findings를 `review-findings.json`에 누적 저장 (SPEC 디렉토리 내). 이 파일이 revision 간 source of truth. `review.md`는 사람이 읽는 요약 보고서로, 판정에 사용하지 않음.

### D4: Multi-Provider Finding Merge
verify 모드에서 동일 finding에 대해 provider 간 상태가 다를 때: **most-conservative** 규칙 적용. 즉, 1개 provider라도 `open`이면 해당 finding은 `open` 유지. 수렴성을 희생하지 않으면서 false-negative를 방지.

### D5: Max Iteration 도달 시 최종 상태
REVISE 루프가 최대 횟수에 도달하고 여전히 open findings가 있으면, 최종 verdict는 `REVISE` (PASS가 아님). 사용자에게 남은 findings와 수동 개입 안내를 표시.

## 요구사항

### REQ-001 [EventDriven] Finding ID 및 Status 추적
WHEN a ReviewFinding is created, THE SYSTEM SHALL assign a revision-agnostic ID (format: `F-{seq}`, 시퀀셜 증분) and a Status field (`open`, `resolved`, `regressed`, `deferred`, `out_of_scope`). Finding은 최초 생성 시 `FirstSeenRev`를 기록하며, 이후 revision에서도 동일 ID를 유지한다.

Status 전이 규칙:
- `open → resolved`: verify에서 해결 확인됨
- `open → deferred`: 의도적 보류 (reviewer 판단)
- `resolved → regressed`: 이전에 resolved였으나 verify에서 다시 미해결 확인됨
- `out_of_scope`: verify 모드에서 새로 발견된 non-critical finding (verdict 계산 제외)

### REQ-002 [EventDriven] Finding Category 및 ScopeRef 분류
WHEN a ReviewFinding is created, THE SYSTEM SHALL assign a Category (`correctness`, `completeness`, `feasibility`, `style`, `security`) and a ScopeRef (requirement ID 또는 file path) to bound the finding's scope.

### REQ-003 [EventDriven] Mode-Aware Review Prompt 생성
WHEN `BuildReviewPrompt` is called with mode `discover`, THE SYSTEM SHALL generate an open-ended full-scan review prompt. WHEN called with mode `verify`, THE SYSTEM SHALL generate a scoped checklist prompt containing only prior unresolved findings and a structured response schema.

**Verify 전용 응답 스키마**:
```
FINDING_STATUS: F-{id} | {open|resolved|regressed} | {reason}
```
예시: `FINDING_STATUS: F-003 | resolved | error handling added at line 42`

discover 모드의 응답 스키마는 기존과 동일 (`FINDING: [severity] description`), 단 Category와 ScopeRef를 필수로 포함:
```
FINDING: [severity] [category] [scope_ref] description
```

### REQ-004 [EventDriven] Phase A — Discovery Review
WHEN a SPEC enters review for the first time (revision 0), THE SYSTEM SHALL execute a `discover` mode review that performs a full-scope scan of requirements, acceptance criteria, and code context.

### REQ-005 [EventDriven] Phase B — Verification Review with Selective Re-Discovery
WHEN a SPEC enters re-review (revision >= 1), THE SYSTEM SHALL execute a `verify` mode review that provides the prior unresolved findings (from `review-findings.json`) as a checklist and instructs the reviewer to report each finding's status using the verify response schema.

Additionally, the verify prompt SHALL instruct the reviewer to report any **regression** or **newly broken behavior** caused by fixes to prior findings, even if not in the original checklist. Such findings are subject to REQ-006 scope filtering (critical/security escape hatch). This ensures verify mode is not blind to fix-induced regressions while maintaining convergence.

### REQ-006 [EventDriven] Scope Lock — Out-of-Scope Filtering with Regression Escape Hatch
WHEN `ParseVerdict` processes a `verify` mode response, THE SYSTEM SHALL:
1. For findings matching a prior ID: update status per reviewer's response (`resolved`, `regressed`, or still `open`)
2. For NEW findings (ID not in prior list) with severity `critical` OR category `security`: register as new `open` findings (escape hatch)
3. For ALL OTHER new findings: tag as `out_of_scope`, record in review.md, but exclude from verdict calculation

### REQ-007 [Unwanted] Circuit Breaker — Non-Convergence Guard
IF the count of `open` + `regressed` findings (excluding `out_of_scope`, `deferred`, **and findings added via REQ-006 escape hatch in the current round**) after a verification round does NOT decrease compared to the previous round's **pre-escape-hatch count**, THE SYSTEM SHALL halt the REVISE loop, emit a warning with the stalled findings list, and return the current state as the final `REVISE` verdict. This detects fixes that introduce regressions at the same rate as resolving issues.

Escape hatch findings (critical/security new findings from REQ-006) are counted separately and reported in the circuit breaker warning, but do NOT trigger the halt. This prevents the paradox where discovering a real security issue would stop the loop prematurely.

### REQ-008 [EventDriven] 정적 분석 통합 (Phase A Deterministic Lane)
WHEN a `discover` mode review begins AND `review_gate.static_analysis` is configured, THE SYSTEM SHALL run configured static analysis tools (e.g., `golangci-lint`) and inject their findings as pre-seeded `ReviewFinding` entries with Category `style` before LLM review. This reduces (not prevents) LLM re-discovery of lint issues. WHEN the LLM reports a finding with the same **normalized** ScopeRef (see REQ-012) and Category `style` as a pre-seeded finding, THE SYSTEM SHALL deduplicate by keeping the pre-seeded finding and discarding the LLM duplicate.

**Config schema** for `review_gate.static_analysis`:
```yaml
review_gate:
  static_analysis:
    enabled: true
    tools:
      - name: golangci-lint
        command: "golangci-lint run --out-format json"
        category: style
```
When `static_analysis` key is absent or `enabled: false`, static analysis is skipped entirely.

### REQ-009 [Ubiquitous] REVISE 루프 구현
THE SYSTEM SHALL implement a REVISE loop in `runSpecReview` that:
1. Executes Phase A (discover, revision 0) on first run
2. On REVISE verdict: persists findings to `review-findings.json`, increments revision, executes Phase B (verify)
3. Repeats up to a configurable maximum iteration count (config key: `review_gate.max_revisions`, code fallback default: `3`)
4. On max iteration reached with open findings remaining: returns final verdict `REVISE` with remaining findings list and manual intervention guidance
5. **PASS condition**: WHEN all findings have status `resolved` or `deferred` (no `open` or `regressed` remaining, `out_of_scope` excluded), THE SYSTEM SHALL return verdict `PASS` and update review.md and SPEC status accordingly
6. **REJECT**: Unchanged from existing behavior — only returned when a provider explicitly returns REJECT

### REQ-010 [EventDriven] Prior Findings 저장 및 조회
WHEN a review round completes, THE SYSTEM SHALL persist the current findings state to `{SPEC_DIR}/review-findings.json` with the following schema:
```json
{
  "spec_id": "SPEC-XXX-001",
  "revision": 1,
  "findings": [
    {"id": "F-001", "status": "open", "severity": "major", "category": "correctness", "scope_ref": "REQ-003", "description": "...", "first_seen_rev": 0, "last_seen_rev": 1}
  ]
}
```
- `last_seen_rev`: the most recent revision in which this finding was evaluated. Used to detect stale findings and track deferred item age.
- WHEN `review-findings.json` is missing or corrupted at revision >= 1, THE SYSTEM SHALL fall back to `discover` mode for that round and log a warning.
- WHEN a verify mode review begins, THE SYSTEM SHALL load prior findings from this file as the checklist source of truth.

### REQ-011 [EventDriven] Multi-Provider Finding Status Merge
WHEN multiple providers return verify results for the same finding ID, THE SYSTEM SHALL apply **supermajority** merge (configurable, default: 2/3 threshold):
- `resolved`: requires ≥ supermajority of providers to agree on `resolved`. Otherwise remains `open`.
- `regressed` > `open` priority: if any provider reports `regressed`, the finding is `regressed` (takes precedence over `open`).
- `open`: if < supermajority agree on `resolved` and no provider reports `regressed`, status is `open`.
- WHEN a provider fails to return a result for a finding (timeout, error), that provider is excluded from the vote count (does not count as a veto).

This balances false-negative prevention with convergence — unanimous agreement is too strict given LLM behavioral variance.

### REQ-012 [Ubiquitous] ScopeRef Normalization
THE SYSTEM SHALL normalize ScopeRef values before comparison (dedup, matching). Normalization rules:
1. File paths: strip leading `./`, normalize to `{package}/{file}:{line}` format (e.g., `pkg/spec/types.go:42`)
2. Line numbers: optional — if present, match at file level when line differs by ≤ 5
3. Requirement refs: exact match (e.g., `REQ-003`)
4. Case-insensitive comparison for file paths

## 생성 파일 상세

| 파일 | 역할 | 변경 유형 |
|------|------|-----------|
| `pkg/spec/types.go` | ReviewFinding에 ID, Status, Category, ScopeRef, FirstSeenRev 필드 추가; FindingStatus 타입; ReviewMode 타입 | 수정 |
| `pkg/spec/prompt.go` | BuildReviewPrompt에 ReviewPromptOptions 파라미터 추가, discover/verify 프롬프트 분기, verify 응답 스키마 포함 | 수정 |
| `pkg/spec/reviewer.go` | ParseVerdict에 mode/priorFindings 파라미터, scope filtering, regression escape hatch, circuit breaker, multi-provider merge | 수정 |
| `pkg/spec/findings.go` | review-findings.json 읽기/쓰기, finding dedup 로직, ScopeRef 정규화 (신규) | 신규 |
| `pkg/spec/static_analysis.go` | 정적 분석 도구 실행 및 Finding 변환 (신규) | 신규 |
| `internal/cli/spec_review.go` | REVISE 루프, 모드 전환, 정적 분석 통합, revision 관리 | 수정 |
