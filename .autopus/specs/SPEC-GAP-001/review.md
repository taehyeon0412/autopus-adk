# Review: SPEC-GAP-001

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-03-24 22:26:14

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| claude | critical | ** REQ-001의 "SigMap" 개념이 정의되지 않았다. 코드베이스에 SigMap 관련 구현이 전혀 없으며, SPEC 본문에서도 SigMap이 무엇인지(함수 시그니처 맵? 시그널 맵?) 설명이 없다. 구현자가 해석할 수 없다. |
| claude | critical | ** REQ-002의 체크포인트/재개 메커니즘이 현재 아키텍처와 충돌한다. 현재 파이프라인은 Claude Code 세션 내에서 에이전트 오케스트레이션으로 실행되며, CLI 바이너리(`auto`)가 파이프라인 상태를 관리하는 구조가 아니다. `.autopus/pipeline-state/` 에 YAML을 저장하더라도 Claude Code 세션의 에이전트 컨텍스트(대화 히스토리, 에이전트 상태)는 복원할 수 없다. `--continue` 플래그의 실제 재개 범위와 한계를 명시해야 한다. |
| claude | major | ** REQ-003에서 `gates.mode: mandatory` 설정이 `autopus.yaml` 스키마에 정의되어 있지 않다. 또한 "파이프라인 스킬 내 Gate 로직에서 `auto check` CLI 명령을 호출"한다는 설계가 모호하다 — 스킬(.md 파일)이 CLI를 호출하는 메커니즘이 없다. hooks를 통해 구현할 것인지, 에이전트가 Bash로 실행할 것인지 명확히 해야 한다. |
| claude | major | ** REQ-005의 `auto react check`는 네이밍이 혼란스럽다. "react"가 React.js 프레임워크를 연상시키지만, 실제로는 CI 상태 조회 기능이다. `auto ci check` 또는 `auto ci status`가 의도를 더 명확히 전달한다. |
| claude | major | ** 모든 REQ가 `[Ubiquitous]` 우선순위로 되어 있다. 6개 요구사항을 모두 동일 우선순위로 두면 실질적인 우선순위 결정이 불가능하다. 최소한 REQ-002(체크포인트)와 REQ-006(deep-worker)은 복잡도가 크게 다르므로 차등이 필요하다. |
| claude | minor | ** REQ-004의 "기존 스킬/에이전트 패턴을 분석"하는 범위가 불명확하다. 현재 `.claude/agents/autopus/`와 `.claude/skills/autopus/` 내 .md 파일의 구조를 파싱하여 템플릿화한다는 의미인지, 아니면 LLM이 패턴을 추론한다는 의미인지 구분이 필요하다. |
| claude | minor | ** REQ-006의 `deep-worker` 에이전트가 "장시간 자율 작업에 최적화"라고만 되어 있고, 기존 에이전트(executor, explorer 등)와의 차별점이 명시되지 않았다. 프롬프트 길이 제한, 컨텍스트 관리 전략, 중간 체크포인트 등 구체적 최적화 방향이 필요하다. |
| claude | suggestion | ** Acceptance criteria가 누락되어 있다. 각 REQ에 대해 "무엇이 충족되면 완료인지" 검증 가능한 기준이 없으면 구현 완료 판단이 주관적이 된다. |
| claude | suggestion | ** 기존 코드 컨텍스트에 `content/embed.go`의 embed 패턴이 포함되어 있지만, 새로운 에이전트/스킬 파일(REQ-004, REQ-006)이 embed에 자동 포함되는 흐름에 대한 언급이 없다. `//go:embed` 패턴 업데이트 필요 여부를 명시해야 한다. |
| gemini | critical | REQ-001 (TypeScript SigMap) lacks a technical implementation strategy. Implementing robust TypeScript analysis in a Go-based codebase requires significant dependencies (e.g., Tree-sitter bindings or an embedded runtime) which are not addressed, posing a high feasibility risk for Phase 1. |
| gemini | major | REQ-002 (Pipeline State) refers to a `--continue` flag and a "Pipeline" concept that is not yet reflected in the CLI architecture. Existing commands are task-specific (init, check, doctor), and the SPEC fails to define which command will act as the pipeline orchestrator. |
| gemini | minor | REQ-005 (auto react check) is overly narrow in its naming. To maintain the ADK's general-purpose utility, this should be generalized (e.g., `auto ci check`) or implemented as a framework-specific plugin/flag. |
| gemini | suggestion | REQ-003 specifies calling the CLI `auto check`. While this ensures consistency, the SPEC should also mandate a machine-readable output format (e.g., `--json`) for `check` to allow the calling Gate logic to parse results reliably. |
| gemini | minor | REQ-004 (Builder Agent) does not specify the target directory for generated skills/agents. It should clarify if these are added to the local `.autopus/` project directory or the ADK's core `content/` library. |

## Provider Responses

### Response 1

# SPEC-GAP-001 리뷰

## VERDICT: REVISE

---

## Findings

**FINDING: [critical]** REQ-001의 "SigMap" 개념이 정의되지 않았다. 코드베이스에 SigMap 관련 구현이 전혀 없으며, SPEC 본문에서도 SigMap이 무엇인지(함수 시그니처 맵? 시그널 맵?) 설명이 없다. 구현자가 해석할 수 없다.

**FINDING: [critical]** REQ-002의 체크포인트/재개 메커니즘이 현재 아키텍처와 충돌한다. 현재 파이프라인은 Claude Code 세션 내에서 에이전트 오케스트레이션으로 실행되며, CLI 바이너리(`auto`)가 파이프라인 상태를 관리하는 구조가 아니다. `.autopus/pipeline-state/` 에 YAML을 저장하더라도 Claude Code 세션의 에이전트 컨텍스트(대화 히스토리, 에이전트 상태)는 복원할 수 없다. `--continue` 플래그의 실제 재개 범위와 한계를 명시해야 한다.

**FINDING: [major]** REQ-003에서 `gates.mode: mandatory` 설정이 `autopus.yaml` 스키마에 정의되어 있지 않다. 또한 "파이프라인 스킬 내 Gate 로직에서 `auto check` CLI 명령을 호출"한다는 설계가 모호하다 — 스킬(.md 파일)이 CLI를 호출하는 메커니즘이 없다. hooks를 통해 구현할 것인지, 에이전트가 Bash로 실행할 것인지 명확히 해야 한다.

**FINDING: [major]** REQ-005의 `auto react check`는 네이밍이 혼란스럽다. "react"가 React.js 프레임워크를 연상시키지만, 실제로는 CI 상태 조회 기능이다. `auto ci check` 또는 `auto ci status`가 의도를 더 명확히 전달한다.

**FINDING: [major]** 모든 REQ가 `[Ubiquitous]` 우선순위로 되어 있다. 6개 요구사항을 모두 동일 우선순위로 두면 실질적인 우선순위 결정이 불가능하다. 최소한 REQ-002(체크포인트)와 REQ-006(deep-worker)은 복잡도가 크게 다르므로 차등이 필요하다.

**FINDING: [minor]** REQ-004의 "기존 스킬/에이전트 패턴을 분석"하는 범위가 불명확하다. 현재 `.claude/agents/autopus/`와 `.claude/skills/autopus/` 내 .md 파일의 구조를 파싱하여 템플릿화한다는 의미인지, 아니면 LLM이 패턴을 추론한다는 의미인지 구분이 필요하다.

**FINDING: [minor]** REQ-006의 `deep-worker` 에이전트가 "장시간 자율 작업에 최적화"라고만 되어 있고, 기존 에이전트(executor, explorer 등)와의 차별점이 명시되지 않았다. 프롬프트 길이 제한, 컨텍스트 관리 전략, 중간 체크포인트 등 구체적 최적화 방향이 필요하다.

**FINDING: [suggestion]** Acceptance criteria가 누락되어 있다. 각 REQ에 대해 "무엇이 충족되면 완료인지" 검증 가능한 기준이 없으면 구현 완료 판단이 주관적이 된다.

**FINDING: [suggestion]** 기존 코드 컨텍스트에 `content/embed.go`의 embed 패턴이 포함되어 있지만, 새로운 에이전트/스킬 파일(REQ-004, REQ-006)이 embed에 자동 포함되는 흐름에 대한 언급이 없다. `//go:embed` 패턴 업데이트 필요 여부를 명시해야 한다.

---

## 근거

SPEC의 방향성(다국어 지원, 파이프라인 견고성, 스캐폴딩)은 ADK 경쟁력 강화에 적절하다. 그러나 핵심 요구사항(REQ-001 SigMap 미정의, REQ-002 아키텍처 충돌)에 구현 불가능하거나 해석이 분분한 부분이 있어 현 상태로는 구현에 착수할 수 없다. Critical 2건과 major 3건을 해소한 후 재리뷰가 필요하다.


### Response 2

MCP issues detected. Run /mcp list for status.1. VERDICT: REVISE

2. FINDINGS:
- FINDING: [critical] REQ-001 (TypeScript SigMap) lacks a technical implementation strategy. Implementing robust TypeScript analysis in a Go-based codebase requires significant dependencies (e.g., Tree-sitter bindings or an embedded runtime) which are not addressed, posing a high feasibility risk for Phase 1.
- FINDING: [major] REQ-002 (Pipeline State) refers to a `--continue` flag and a "Pipeline" concept that is not yet reflected in the CLI architecture. Existing commands are task-specific (init, check, doctor), and the SPEC fails to define which command will act as the pipeline orchestrator.
- FINDING: [minor] REQ-005 (auto react check) is overly narrow in its naming. To maintain the ADK's general-purpose utility, this should be generalized (e.g., `auto ci check`) or implemented as a framework-specific plugin/flag.
- FINDING: [suggestion] REQ-003 specifies calling the CLI `auto check`. While this ensures consistency, the SPEC should also mandate a machine-readable output format (e.g., `--json`) for `check` to allow the calling Gate logic to parse results reliably.
- FINDING: [minor] REQ-004 (Builder Agent) does not specify the target directory for generated skills/agents. It should clarify if these are added to the local `.autopus/` project directory or the ADK's core `content/` library.

3. REASONING:
The SPEC provides a strong vision for competitive parity, but it skips the necessary architectural groundwork. The transition from a Go-centric tool to a multi-language analysis engine (REQ-001) and the introduction of stateful long-running pipelines (REQ-002) require more detailed technical definitions to be considered feasible and complete. Specifically, the "Pipeline" runner's relationship with existing commands like `init` needs to be clarified to avoid feature fragmentation.


### Response 3

# SPEC-GAP-001 최종 판정

## VERDICT: REVISE

---

## 리뷰어 평가

두 리뷰 모두 핵심 문제를 정확히 짚었으며, 상호 보완적이다.

| 영역 | Claude | Gemini | 판정 |
|------|--------|--------|------|
| REQ-001 SigMap 정의 부재 | critical — 개념 자체 미정의 | critical — Go에서 TS 분석 구현 전략 부재 | **양쪽 모두 채택**. SigMap 정의 + 기술 구현 전략 둘 다 필요 |
| REQ-002 아키텍처 충돌 | critical — Claude Code 세션 컨텍스트 복원 불가 | major — 파이프라인 오케스트레이터 미정의 | **Claude 쪽이 더 정확**. 에이전트 컨텍스트 복원 한계는 critical |
| REQ-003 Gate 메커니즘 | major — 스킬이 CLI 호출하는 경로 모호 | suggestion — machine-readable 출력 필요 | **Claude 쪽 채택** + Gemini의 `--json` 출력 제안도 유효 |
| REQ-005 네이밍 | major — `react`가 React.js 연상 | minor — 프레임워크 종속 네이밍 | **합의**. `auto ci check`로 변경 권고 |
| 우선순위 일률 적용 | major — 차등 필요 | 미언급 | **Claude 쪽 채택**. 6개 전부 Ubiquitous는 비현실적 |
| embed 패턴 영향 | suggestion | 미언급 | **채택**. `content/embed.go` 업데이트 경로 명시 필요 |
| Acceptance criteria 부재 | suggestion | 미언급 | **채택**. 검증 기준 없으면 완료 판단 불가 |
| REQ-004 생성 위치 | 미언급 | minor — 타겟 디렉토리 미명시 | **Gemini 쪽 채택** |

---

## 최종 필수 수정 사항

### Critical (반드시 해소 후 재리뷰)

1. **REQ-001**: "SigMap" 개념을 정의하고, Go 바이너리에서 TypeScript AST를 분석하는 기술 전략(Tree-sitter, `go/ast` 대응체 등)을 명시할 것
2. **REQ-002**: Claude Code 세션 기반 아키텍처에서 체크포인트/재개의 실제 범위를 한정할 것. 에이전트 대화 컨텍스트는 복원 불가능하므로, "재개"가 의미하는 바를 정확히 정의할 것 (예: Phase 단위 재시작만 가능, 에이전트 중간 상태 복원은 범위 밖)

### Major (수정 필요)

3. **REQ-003**: 스킬(.md)이 `auto check`를 호출하는 구체적 메커니즘 명시 (hooks vs 에이전트 Bash 실행). `--json` 출력 포맷 지원 포함
4. **REQ-005**: `auto react check` → `auto ci check`로 변경
5. **우선순위 차등화**: 최소 2단계로 분리. REQ-006(deep-worker .md 파일 추가)은 복잡도가 낮고, REQ-002(체크포인트)는 아키텍처 변경이 필요

### Minor (권고)

6. **REQ-004**: 생성된 스킬/에이전트의 타겟 디렉토리 명시 (프로젝트 `.claude/` vs ADK `content/`)
7. **REQ-006**: 기존 executor/explorer 대비 deep-worker의 구체적 차별점 명시
8. 각 REQ에 검증 가능한 acceptance criteria 추가
9. `content/embed.go`의 `//go:embed` 패턴 업데이트 필요 여부 명시

---

## 근거

두 리뷰어가 독립적으로 REQ-001과 REQ-002를 최우선 문제로 지목했다. 이는 SPEC의 가장 야심찬 두 요구사항이 동시에 가장 불명확하다는 것을 의미한다. 방향성은 타당하나, critical 2건이 해소되지 않으면 구현자가 잘못된 가정 위에 작업을 시작할 위험이 크다. Major 수정까지 반영한 후 재리뷰를 권고한다.

─── 🐙 applied: language-policy


