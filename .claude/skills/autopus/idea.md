---
name: idea
description: 멀티 프로바이더 아이디어 토론 및 발산 스킬
triggers:
  - idea
  - 아이디어 토론
  - brainstorm idea
  - 아이디어 발산
category: workflow
level1_metadata: "멀티 프로바이더 아이디어 토론, SCAMPER/HMW/ICE, BS 파일 생성"
---

# Idea Skill

멀티 프로바이더 오케스트라를 활용해 아이디어를 구조화하고 발산 후 BS 파일로 저장합니다.

## 사용법

```
/auto idea "설명" [--strategy debate|consensus|pipeline|fastest] [--providers list] [--auto]
```

**플래그:**
- `--strategy` — 오케스트레이션 전략 지정 (기본값: `debate`)
- `--providers` — 사용할 프로바이더 목록 (기본값: orchestra 설정 전체)
- `--auto` — 완료 후 `/auto plan --from-idea BS-{ID}` 자동 체이닝

## 저장 위치 규칙

BS 파일은 **대상 모듈** 기준으로 저장합니다.

1. 아이디어 설명에서 관련 코드를 검색하여 대상 서브모듈을 자동 감지
2. **단일 모듈 대상**: `{target-module}/.autopus/brainstorms/`에 BS 파일 생성
3. **크로스-모듈 (2+ 모듈)**: 루트 `.autopus/brainstorms/`에 BS 파일 생성 (meta repo 커밋 대상)
4. 감지 실패 시 루트 `.autopus/brainstorms/`에 저장
5. BS ID는 프로젝트 전체에서 유일해야 함: `.autopus/brainstorms/BS-*` AND `*/.autopus/brainstorms/BS-*` 스캔

Ref: `.claude/rules/autopus/doc-storage.md` for full storage rules.

## 5단계 파이프라인

### Step 1: Parse Input and Flags

```
# Parse user input
input = args[0]            # required: idea description
strategy = flags.strategy  # default: "debate"
providers = flags.providers # default: all configured providers
auto_chain = flags.auto    # default: false
```

### Step 2: Structure Idea as What/Why/Who/When

입력을 아래 4개 축으로 구조화합니다:

- **What**: 무엇을 만드는가?
- **Why**: 왜 필요한가? (문제/기회)
- **Who**: 누구를 위한 것인가? (대상 사용자)
- **When**: 언제 필요한가? (타임라인/맥락)

#### Opportunity-Solution Tree (선택)

아이디어가 기존 제품 개선인 경우, OST 프레임워크로 구조화:

```
Outcome (목표)
  └─ Opportunity (기회/문제)
       ├─ Solution A
       │    └─ Experiment (검증 방법)
       ├─ Solution B
       │    └─ Experiment
       └─ Solution C
            └─ Experiment
```

- **Outcome**: 달성하려는 비즈니스/사용자 목표
- **Opportunity**: 사용자의 unmet need 또는 pain point
- **Solution**: 기회를 해결하는 구체적 방안
- **Experiment**: 솔루션의 가정을 검증하는 최소 실험

#### Assumption Identification

아이디어의 핵심 가정을 4축으로 식별:

| 축 | 질문 | 예시 |
|---|---|---|
| **Value** | 사용자가 이것을 원하는가? | "사용자가 자동 분석을 필요로 한다" |
| **Usability** | 사용자가 이것을 쓸 수 있는가? | "CLI 인터페이스로 충분하다" |
| **Feasibility** | 기술적으로 구현 가능한가? | "LLM API 지연이 허용 범위 내다" |
| **Viability** | 비즈니스적으로 지속 가능한가? | "API 비용이 수익 내에서 감당 가능하다" |

가장 위험한 가정(높은 Impact × 높은 Uncertainty)을 상위 3개 식별합니다.

### [REQUIRED] Step 3: Call Orchestra Brainstorm (MUST call Bash tool)

IMPORTANT: 이 단계는 반드시 Bash 툴로 CLI를 실행해야 합니다. Sequential Thinking이나 단일 모델 시뮬레이션으로 대체 금지.

#### Multi-Perspective Brainstorming

Orchestra 프롬프트에 3가지 관점을 포함하여 다각적 발산을 유도:

- **PM 관점**: 사용자 가치, 비즈니스 임팩트, 우선순위
- **Designer 관점**: UX, 접근성, 사용자 여정, 인터랙션 패턴
- **Engineer 관점**: 기술적 실현 가능성, 아키텍처, 성능, 보안

```bash
auto orchestra brainstorm "{structured idea}" \
  --strategy {strategy} --no-judge --yield-rounds
```

- `--no-judge`: subprocess judge를 건너뛰고 주 세션이 직접 판정
- `--yield-rounds`: Round 1 후 JSON을 stdout으로 출력하고 pane을 유지한 채 종료
- 출력은 JSON (`YieldOutput` 구조)으로, `round_history[0].responses[]`에 프로바이더별 결과 포함
- 프로바이더 실패 시 graceful degradation — 나머지 프로바이더 결과로 계속 진행
- Bash 호출이 에러를 반환한 경우에만 사용자에게 fallback 여부를 확인

> **⏭ POST-STEP**: JSON 결과 수신 후 Step 4로 진행. Step 5로 건너뛰지 말 것.

### Step 4: Main-Session Judge — ICE Scoring and Top N Selection

IMPORTANT: 주 세션이 직접 judge 역할을 수행합니다. Step 3의 JSON 출력에서 각 프로바이더의 `output`을 파싱하여 아이디어를 통합하고 ICE 점수로 수렴합니다.

#### 4-1. 프로바이더 응답 파싱

Step 3 JSON의 `round_history[0].responses[]`에서 각 프로바이더의 SCAMPER 아이디어와 HMW 질문을 추출합니다.

#### 4-2. 아이디어 통합 (AUGMENT, not FILTER)

모든 프로바이더의 아이디어를 병합합니다. 중복은 통합하되 유사한 아이디어도 보존합니다. 발산적 아이디어는 부록에 포함합니다.

#### 4-3. ICE 스코어링

| 항목 | 설명 | 범위 |
|------|------|------|
| Impact | 영향력 | 1-10 |
| Confidence | 확신도 | 1-10 |
| Ease | 실행 용이성 | 1-10 |

`Score = (Impact × Confidence × Ease) / 100`

상위 N개 아이디어를 선별하여 BS 파일에 기록합니다.

#### 4-4. Assumption Risk Overlay

ICE Top N 아이디어 각각에 대해 Step 2에서 식별한 가정의 위험도를 오버레이:

| Rank | Idea | ICE Score | Top Risk Assumption | Risk Level |
|------|------|-----------|---------------------|------------|
| 1 | ... | 7.2 | "사용자가 X를 원한다" (Value) | HIGH |
| 2 | ... | 6.8 | "API 지연 < 500ms" (Feasibility) | MEDIUM |

HIGH 위험 가정이 있는 아이디어는 `/auto plan` 전에 검증 실험을 권장합니다.

#### 4-5. Pane Cleanup

판정 완료 후 Step 3에서 유지된 pane을 정리합니다:

```bash
auto orchestra cleanup --session {session_id}
```

### Step 5: Save and Guide Next Steps

BS-{ID} 파일 저장 후 Workflow Lifecycle 바 표시 및 다음 단계 안내.

**ID 자동 증분**: `{target-module}/.autopus/brainstorms/BS-{ID}.md` 파일이 이미 존재하면 ID를 증분합니다. 전체 프로젝트 스캔으로 ID 유일성을 보장합니다.

## BS 파일 형식

`{target-module}/.autopus/brainstorms/BS-{ID}.md`:

```markdown
# BS-{ID}: {title}

**Created**: {date}
**Strategy**: {strategy}
**Providers**: {provider list}
**Status**: active

## 원본 아이디어
- What: {description}
- Why: {motivation}
- Who: {target users}
- When: {timeline}

## 프로바이더별 발산 결과
{raw brainstorm output}

## ICE 스코어링 — Top N
| Rank | Idea | Impact | Confidence | Ease | Score |
|------|------|--------|------------|------|-------|

## 추천 방향
{judge's recommendation}

## 다음 단계
`/auto plan --from-idea BS-{ID} "feature description"`
```

## 완료 후 출력

```
🐙 Workflow: BS-{ID}
  ● idea  →  ○ plan  →  ○ go  →  ○ sync
```

`--auto` 플래그가 있으면 자동으로 `/auto plan --from-idea BS-{ID}`로 체이닝합니다.
