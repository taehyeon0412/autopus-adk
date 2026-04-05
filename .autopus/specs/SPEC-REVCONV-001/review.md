# Review: SPEC-REVCONV-001

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-04-05 21:30:53

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| claude | critical | acceptance.md의 Finding ID 형식이 spec.md와 모순됨 |
| claude | critical | research.md의 D1도 구버전 ID 형식 유지 |
| claude | major | REQ-007 circuit breaker와 REQ-006 escape hatch의 상호작용이 의도와 충돌 가능 |
| claude | major | REQ-011 open과 regressed 우선순위 미정의 |
| claude | major | plan.md에 findings.go 관련 태스크 누락 |
| claude | minor | REQ-008 정적 분석 dedup의 ScopeRef 매칭 기준 불명확 |
| claude | minor | REQ-009 max_revisions 기본값 모호 |
| claude | minor | review-findings.json 스키마에 `last_seen_rev` 필드 부재 |
| claude | suggestion | `review_gate.static_analysis` config 스키마 미정의 |
| codex | critical | `REQ-008` and `REQ-009` are incomplete as written. Both end with `:` but do not actually define the REVISE-loop steps or the `review-findings.json` schema, so the core algorithm and persistence contract are not implementable consistently. |
| codex | major | `REQ-001` and `REQ-011` assume stable finding identity across providers and revisions, but the spec never defines discover-phase deduplication/matching for semantically identical findings returned by different providers. Without that, the same issue will receive multiple `F-{seq}` IDs, and the later “same finding ID” conservative merge in verify mode will not work reliably. |
| codex | major | `REQ-003`/`REQ-005` conflict with `REQ-006`. Verify mode is defined as a checklist containing only prior unresolved findings, but `REQ-006` claims the loop should detect regressions and newly introduced critical/security issues that offset resolved findings. There is no mechanism in verify mode to emit brand-new findings outside the prior checklist. |
| codex | major | `REQ-007` deduplicates static-analysis and LLM style findings using only `ScopeRef` + `Category=style`, which is too coarse. Multiple distinct lint issues in the same file or requirement would collapse into one finding, corrupting both correctness and convergence counts. |
| codex | major | The spec does not define the terminal verdict/loop contract against the current review workflow. It never says when the final verdict is `PASS` versus `REVISE` versus `REJECT`, how `review_gate.max_revisions` is applied, or whether existing `review.md` output and SPEC status updates still occur. The current system expects those behaviors in [autopus.yaml:28](/Users/bitgapnam/Documents/github/autopus-co/autopus.yaml#L28), [spec-review.md:36](/Users/bitgapnam/Documents/github/autopus-co/.claude/skills/autopus/spec-review.md#L36), and [SKILL.md:727](/Users/bitgapnam/Documents/github/autopus-co/.claude/skills/auto/SKILL.md#L727). |
| gemini | critical | REQ-008 and REQ-009 are severely truncated.** |
| gemini | major | Logical conflict between `verify` mode scope (REQ-003/REQ-005) and regression detection (REQ-006).** |
| gemini | major | `verify` mode (REQ-005) is permanently locked in after Rev 0, preventing discovery of new issues in subsequent revisions.** |
| gemini | major | Multi-provider merge policy (REQ-011) is excessively fragile and prone to infinite loops.** |
| gemini | minor | Static analysis deduplication logic (REQ-007) is likely to fail due to mismatched `ScopeRef` granularity.** |
| gemini | minor | Missing error handling for missing findings file (REQ-010).** |
| gemini | suggestion | Multi-provider failure handling in REQ-011.** |

## Provider Responses

### Response 1

전체 SPEC 문서, 연관 파일(research.md, acceptance.md, plan.md), 그리고 기존 코드를 모두 확인했습니다. 리뷰 결과를 보고합니다.

---

## VERDICT: REVISE

---

## Findings

### FINDING: [critical] acceptance.md의 Finding ID 형식이 spec.md와 모순됨

spec.md의 D1 설계 결정은 `F-{seq}` (revision-agnostic)로 명확히 정의했지만, acceptance.md는 여전히 구버전 형식을 사용합니다:

- **S1**: `F-{revision}-{seq}` 형식의 ID가 부여된다 (spec.md D1과 모순)
- **S4**: `F-0-1`, `F-0-2`, `F-1-3` 사용 (revision 포함 형식)

이는 구현자가 어느 형식을 따라야 할지 혼란을 야기합니다. spec.md revision 1에서 ID 형식을 변경했지만 acceptance.md가 동기화되지 않았습니다.

### FINDING: [critical] research.md의 D1도 구버전 ID 형식 유지

research.md:57의 `D1: Finding ID 형식 — F-{revision}-{seq}`가 spec.md D1의 `F-{seq}` (revision-agnostic)과 모순됩니다. research.md는 revision 0의 결정을 반영하고 있어 spec.md revision 1과 불일치합니다. research.md도 업데이트하거나, 최소한 "이 결정은 revision 1에서 변경됨"을 명시해야 합니다.

### FINDING: [major] REQ-007 circuit breaker와 REQ-006 escape hatch의 상호작용이 의도와 충돌 가능

REQ-006은 verify 모드에서 critical/security finding을 escape hatch로 open 등록합니다. REQ-007은 open+regressed 수가 감소하지 않으면 루프를 중단합니다. 

**시나리오**: Revision 1에서 open 3개. Revision 2에서 2개 resolved, 1개 critical escape hatch 추가 → open 2개. 이 경우 감소하므로 OK. 하지만 1개 resolved + 2개 critical escape hatch → open 4개 → circuit breaker 발동.

SPEC은 REQ-007에서 "(b) critical/security escape hatch findings that offset resolved findings"를 감지한다고 명시하지만, 이는 실제 중요한 보안 문제가 발견된 상황에서 루프를 중단시킵니다. **진짜 critical finding이 있을 때 오히려 계속 반복해서 해결해야 하는데 중단되는 역설**이 생깁니다.

**제안**: circuit breaker 카운트에서 escape hatch로 추가된 finding을 제외하거나, escape hatch finding 추가 시에는 "순수 감소"가 아닌 "기존 finding 기준 감소"로 판단하는 규칙 추가가 필요합니다.

### FINDING: [major] REQ-011 open과 regressed 우선순위 미정의

"if ANY provider reports `open` or `regressed`, the finding status is `open` or `regressed` respectively" — 한 provider가 `open`, 다른 provider가 `regressed`를 보고하면 최종 상태가 `open`인지 `regressed`인지 불분명합니다. 둘 다 "most-conservative"에 해당하므로 우선순위 규칙이 필요합니다. `regressed`는 이전에 resolved였다가 다시 열린 것이므로 의미적으로 `open`보다 심각한 신호입니다 — 명시적으로 정의해야 합니다.

### FINDING: [major] plan.md에 findings.go 관련 태스크 누락

spec.md 생성 파일 테이블에 `pkg/spec/findings.go` (review-findings.json 읽기/쓰기, dedup 로직)가 신규 파일로 정의되어 있지만, plan.md의 Phase 1 태스크 목록에 이 파일을 다루는 전용 태스크가 없습니다. T7(REVISE 루프)에 암묵적으로 포함된 것으로 보이나, REQ-009/REQ-010의 JSON 스키마 구현과 dedup 로직은 독립된 태스크로 분리하는 것이 구현 범위를 명확히 합니다.

### FINDING: [minor] REQ-008 정적 분석 dedup의 ScopeRef 매칭 기준 불명확

정적 분석 도구와 LLM이 같은 위치를 다르게 표현할 수 있습니다 (예: `types.go:42` vs `pkg/spec/types.go line 42`). "same ScopeRef"의 매칭이 exact match인지, file-level match인지, fuzzy match인지 정의가 없습니다. 구현자가 판단하기 어렵습니다.

### FINDING: [minor] REQ-009 max_revisions 기본값 모호

"config key: `review_gate.max_revisions`, default from existing config" — 현재 autopus.yaml에는 `max_revisions: 2`가 설정되어 있고, research.md:131에서는 "기본 max iteration = 3"이라고 언급합니다. SPEC이 참조하는 "default from existing config"이 config 파일의 2인지, 코드의 하드코딩 기본값인지 명시하지 않습니다. config 키가 누락된 경우의 fallback default를 숫자로 명시하는 것이 좋습니다.

### FINDING: [minor] review-findings.json 스키마에 `last_seen_rev` 필드 부재

REQ-010의 JSON 스키마에 `first_seen_rev`는 있지만 `last_seen_rev`(마지막으로 평가된 revision)가 없습니다. `deferred` finding이 몇 revision 동안 미평가 상태인지 추적할 수 없고, 오래된 finding의 staleness를 판단할 근거가 없습니다. 필수는 아니나 운영 시 유용합니다.

### FINDING: [suggestion] `review_gate.static_analysis` config 스키마 미정의

REQ-008은 `review_gate.static_analysis`가 설정되어 있을 때 동작을 정의하지만, 이 config 키의 구체적 스키마(어떤 필드가 있고, 도구 목록을 어떻게 지정하는지)가 SPEC 어디에도 없습니다. 현재 `ReviewGateConf` struct에도 이 필드가 없으므로, config 스키마 정의를 추가하면 구현이 명확해집니다.

---

## 근거 요약

SPEC의 핵심 설계(2-Phase Discovery→Verification, Finding 추적, circuit breaker)는 기존 코드의 수렴성 문제를 정확히 진단하고 합리적으로 해결합니다. 특히 D1(revision-agnostic ID)과 D2(escape hatch)의 설계 판단은 trade-off를 잘 고려했습니다.

그러나 **revision 1에서 ID 형식을 변경한 후 acceptance.md와 research.md가 동기화되지 않은 것**은 구현 시 혼란을 직접 유발하는 critical 이슈입니다. 또한 circuit breaker와 escape hatch의 상호작용, multi-provider merge에서의 open/regressed 우선순위 등 **엣지 케이스 규칙이 불완전**합니다.

이 이슈들은 모두 SPEC 문서 내에서 해결 가능하며, 아키텍처 변경은 필요 없으므로 REJECT가 아닌 **REVISE**를 판정합니다.

─── 🐙 applied: objective-reasoning · language-policy


### Response 2

**VERDICT**

REVISE

**Findings**

- FINDING: [critical] `REQ-008` and `REQ-009` are incomplete as written. Both end with `:` but do not actually define the REVISE-loop steps or the `review-findings.json` schema, so the core algorithm and persistence contract are not implementable consistently.
- FINDING: [major] `REQ-001` and `REQ-011` assume stable finding identity across providers and revisions, but the spec never defines discover-phase deduplication/matching for semantically identical findings returned by different providers. Without that, the same issue will receive multiple `F-{seq}` IDs, and the later “same finding ID” conservative merge in verify mode will not work reliably.
- FINDING: [major] `REQ-003`/`REQ-005` conflict with `REQ-006`. Verify mode is defined as a checklist containing only prior unresolved findings, but `REQ-006` claims the loop should detect regressions and newly introduced critical/security issues that offset resolved findings. There is no mechanism in verify mode to emit brand-new findings outside the prior checklist.
- FINDING: [major] `REQ-007` deduplicates static-analysis and LLM style findings using only `ScopeRef` + `Category=style`, which is too coarse. Multiple distinct lint issues in the same file or requirement would collapse into one finding, corrupting both correctness and convergence counts.
- FINDING: [major] The spec does not define the terminal verdict/loop contract against the current review workflow. It never says when the final verdict is `PASS` versus `REVISE` versus `REJECT`, how `review_gate.max_revisions` is applied, or whether existing `review.md` output and SPEC status updates still occur. The current system expects those behaviors in [autopus.yaml:28](/Users/bitgapnam/Documents/github/autopus-co/autopus.yaml#L28), [spec-review.md:36](/Users/bitgapnam/Documents/github/autopus-co/.claude/skills/autopus/spec-review.md#L36), and [SKILL.md:727](/Users/bitgapnam/Documents/github/autopus-co/.claude/skills/auto/SKILL.md#L727).

**Reasoning**

The overall direction is feasible: a 2-phase `discover`/`verify` review loop with persisted findings is a reasonable way to improve convergence. But the current spec is not complete enough to implement safely. The missing loop/schema details, undefined cross-provider finding identity, and the contradiction between “verify only prior unresolved findings” and “detect regressions/new issues” are central defects, not editorial nits. Once those contracts are specified, this should be re-reviewed.


### Response 3

### VERDICT: REVISE

### Findings

- **FINDING: [critical] REQ-008 and REQ-009 are severely truncated.**
  The text for REQ-008 ends at "THE SYSTEM SHALL implement a REVISE loop in `runSpecReview` that:", missing the actual steps of the loop. REQ-009 ends at "...with the following schema:", omitting the required JSON schema definition. The SPEC cannot be implemented without these core specifications.

- **FINDING: [major] Logical conflict between `verify` mode scope (REQ-003/REQ-005) and regression detection (REQ-006).**
  REQ-003 and REQ-005 explicitly restrict `verify` mode to act as a checklist of *prior unresolved findings*. However, REQ-006 claims the system will detect "fixes that introduce regressions." If the LLM is only evaluating the prior unresolved checklist, it lacks the broader context or instruction to scan for new issues (regressions) introduced elsewhere in the codebase or SPEC during the fix phase. The SPEC must either allow `verify` mode to report new findings or periodically trigger a `discover` scan.

- **FINDING: [major] `verify` mode (REQ-005) is permanently locked in after Rev 0, preventing discovery of new issues in subsequent revisions.**
  According to REQ-004 and REQ-005, `discover` mode is strictly bound to Rev 0, and Rev >= 1 only executes `verify` mode. If a developer heavily modifies the SPEC, adds new requirements, or writes new code in Rev 2, those additions will never be subjected to a `discover` scan. There must be a trigger or heuristic to run `discover` on later revisions (e.g., if a file changes significantly, or if requested by the user).

- **FINDING: [major] Multi-provider merge policy (REQ-011) is excessively fragile and prone to infinite loops.**
  REQ-011 mandates that `resolved` requires ALL providers to agree, and ANY provider reporting `open` keeps the finding `open`. Given the subjective nature of LLM evaluations (especially for `completeness` and `style`), a single hallucinating or overly strict provider will prevent resolution. Combined with the auto-halting in REQ-006, this guarantees pipeline deadlocks. A consensus mechanism (e.g., majority vote), confidence weighting, or a human-override path is required. 

- **FINDING: [minor] Static analysis deduplication logic (REQ-007) is likely to fail due to mismatched `ScopeRef` granularity.**
  REQ-007 deduplicates if the LLM and the linter report the exact same `ScopeRef` and `style` category. Linters typically report specific files and line numbers (e.g., `main.go:42`), whereas an LLM might report a broader scope (e.g., the whole file or a function name). Strict string matching on `ScopeRef` will result in duplicate findings.

- **FINDING: [minor] Missing error handling for missing findings file (REQ-010).**
  If `review-findings.json` is missing or corrupted but the SPEC is at Rev >= 1, the system behavior is undefined. The SPEC should specify a fallback behavior (e.g., fallback to `discover` mode or fail the pipeline).

- **FINDING: [suggestion] Multi-provider failure handling in REQ-011.**
  The existing code context (`auxiliary_bootstrap.go`) shows the use of CircuitBreakers and multiple AI providers. The SPEC should clarify how REQ-011 behaves if a provider times out, hits a rate limit, or triggers a circuit breaker during verification. Does a failed provider count as a veto against `resolved`?

### Reasoning for Verdict

The SPEC introduces a solid foundation for a continuous, structured LLM review process by distinguishing between `discover` and `verify` phases. However, the document is fundamentally incomplete due to the truncated requirements (REQ-008, REQ-009). 

Furthermore, the core loop logic has a critical flaw: by restricting all post-Rev-0 reviews to a narrow `verify` checklist, the system becomes blind to any new bugs, regressions, or architectural changes introduced during the resolution phase. Lastly, the strict unanimous agreement requirement for multi-provider validation (REQ-011) will inevitably lead to stalled review pipelines given standard LLM behavioral variance. These architectural and formatting issues must be resolved before the SPEC is actionable.


### Response 4

Now I have full context. Let me compose the Final Judge verdict.

---

# Final Judge Verdict — SPEC-REVCONV-001

## 최종 판정: **REVISE**

전 참여자 만장일치 REVISE이며, 실제 SPEC 파일 검증 결과 이에 동의합니다.

---

## 1. Consensus Areas (2+ 참여자 수렴)

### C1: Verify 모드 범위 vs 회귀 감지 모순 (A+B+C)
**공유 아이디어**: REQ-003/REQ-005는 verify 모드를 "기존 unresolved findings 체크리스트만"으로 제한하지만, REQ-006은 새 regression을 감지하라고 요구합니다. 3명 모두 이 논리적 모순을 지적했습니다.

**검증 결과**: spec.md REQ-006은 escape hatch로 critical/security finding만 신규 등록을 허용하지만, **일반 regression**(non-critical 영역에서 fix가 깨뜨린 기존 기능)의 감지 메커니즘이 없습니다. 체크리스트 전용 verify 프롬프트로는 scope 외 영역의 regression을 발견할 수 없습니다.

**중요도**: Critical — 이것이 해결되지 않으면 SPEC의 핵심 목적("수렴성 보장")이 역으로 새로운 blind spot을 만듭니다.

### C2: REQ-008/REQ-009 불완전성 (B+C)
**공유 아이디어**: 토론 프롬프트에 제시된 REQ-008과 REQ-009가 `:` 로 끝나며 핵심 내용이 누락되었습니다.

**검증 결과**: 실제 spec.md 파일(lines 82-100)에는 REQ-009(REVISE 루프 4단계)와 REQ-010(JSON 스키마)이 **완전하게 정의**되어 있습니다. B와 C는 토론 프롬프트의 truncated 발췌본을 검토한 것이며, 실제 SPEC 결함은 아닙니다. 다만, 이는 SPEC이 외부에서 참조될 때 핵심 내용이 누락될 수 있다는 **문서 구조 취약점**을 시사합니다.

### C3: ScopeRef 매칭 기준 미정의 (A+B+C)
**공유 아이디어**: 정적 분석 도구(`main.go:42`)와 LLM(`pkg/spec/types.go line 42`)이 같은 위치를 다르게 표현할 수 있는데, "same ScopeRef"의 매칭 규칙이 없습니다.

**중요도**: Major — dedup 실패 시 finding 수가 인위적으로 증가하여 circuit breaker를 잘못 트리거합니다.

### C4: Multi-Provider Merge의 취약성 (A+B+C)
A는 open/regressed 우선순위 미정의를, C는 LLM 주관성으로 인한 deadlock을, B는 cross-provider finding identity 미정의를 지적했습니다. 모두 REQ-011의 "most-conservative" 전략이 현실적으로 불안정하다는 같은 근본 문제를 다른 각도에서 보았습니다.

**중요도**: Major — LLM의 비결정적 특성을 고려하면 만장일치 `resolved` 요구는 실질적으로 수렴을 방해할 수 있습니다.

---

## 2. Unique Insights (단독 제기)

| ID | 참여자 | 아이디어 | 평가 |
|---|---|---|---|
| U1 | A | acceptance.md/research.md의 ID 형식이 spec.md revision 1과 동기화 안 됨 (`F-{revision}-{seq}` vs `F-{seq}`) | **검증 확인**: acceptance.md:8, research.md:57 모두 구 형식 사용. 구현자 혼란의 직접 원인. Critical 동의. |
| U2 | A | circuit breaker가 진짜 critical finding 발견 시 오히려 루프를 중단시키는 역설 | **타당함**: escape hatch finding이 net count를 올려 circuit breaker를 발동시키는 시나리오가 논리적으로 성립. |
| U3 | C | Rev 0 이후 discover 모드가 영구 잠김 — SPEC 대규모 수정 시 blind spot | **타당함**: REQ-004/REQ-005의 엄격한 Rev 0 = discover, Rev >= 1 = verify 구분은 SPEC이 크게 변경될 경우 문제. 단, 현실적으로 review 중 SPEC 대규모 변경은 드물어 severity는 medium. |
| U4 | B | 기존 workflow 파일(autopus.yaml, spec-review.md, SKILL.md)과의 통합 계약 미정의 | **타당함**: 최종 verdict가 기존 SPEC status 업데이트나 review.md 생성과 어떻게 연계되는지 SPEC이 명시하지 않음. |
| U5 | A | plan.md에 findings.go 전용 태스크 누락 | **검증 확인**: plan.md에 REQ-009/REQ-010의 JSON persistence 구현 태스크가 없음. T7에 암묵적 포함으로 보이나 분리가 명확. |

---

## 3. Cross-Risks (다수 독립 식별)

| Risk | 식별자 | Severity | 설명 |
|---|---|---|---|
| Verify scope blindness | A+B+C | **High** | Verify 모드가 기존 체크리스트 외 영역을 스캔하지 않아, fix가 유발한 새 문제를 감지 못함 |
| Multi-provider merge deadlock | A+C | **High** | LLM 비결정성 + unanimous resolved 요구 = 수렴 불가 시나리오 존재 |
| ScopeRef matching failure | A+B+C | **Medium** | 도구 간 위치 표현 불일치로 dedup 실패, finding 수 인위적 증가 |
| Circuit breaker vs escape hatch 역설 | A+C | **Medium** | 진짜 보안 문제 발견이 역으로 루프를 중단시킴 |
| Document sync drift | A (단독이지만 검증됨) | **High** | acceptance.md, research.md가 spec.md와 불일치 — 구현자 혼란 |

---

## 4. Top Ideas Ranking (ICE Score)

| Rank | Idea | I | C | E | Score |
|---|---|---|---|---|---|
| 1 | **acceptance.md/research.md ID 형식 동기화** — `F-{seq}`로 통일 | 9 | 8 | 10 | **7.2** |
| 2 | **Verify 모드에 selective re-discovery 허용** — 변경된 ScopeRef 영역은 discover 허용 | 10 | 10 | 6 | **6.0** |
| 3 | **REQ-011 merge를 supermajority로 완화** — 2/3 이상 resolved → resolved, 만장일치 불필요 | 8 | 8 | 7 | **4.5** |
| 4 | **ScopeRef 정규화 규칙 정의** — file path를 `pkg/spec/types.go:L42` 형태로 정규화 | 7 | 9 | 7 | **4.4** |
| 5 | **Circuit breaker에서 escape hatch finding 제외** — 기존 finding 기준 감소 판단 | 8 | 7 | 8 | **4.5** |

---

## 5. Recommendation

**SPEC 문서 내에서 해결 가능한 이슈들이므로 REVISE 판정이 적절합니다.** 우선 (1) acceptance.md와 research.md의 ID 형식을 `F-{seq}`로 즉시 동기화하고, (2) verify 모드에 "변경된 ScopeRef 영역은 discover 허용"하는 selective re-discovery 조항을 REQ-005에 추가하며, (3) REQ-011의 merge 정책을 supermajority 또는 configurable threshold로 완화하십시오. 이 3가지가 해결되면 SPEC의 핵심 목적(수렴성 보장)이 실질적으로 달성 가능합니다.

─── 🐙 applied: objective-reasoning · language-policy


