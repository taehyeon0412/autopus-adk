# SPEC-REVCONV-001 리서치

## 기존 코드 분석

### ReviewFinding 타입 (pkg/spec/types.go:69-74)

```go
type ReviewFinding struct {
    Provider    string // provider that found the issue
    Severity    string // critical, major, minor, suggestion
    Description string // finding description
}
```

3개 필드만 존재. ID, Status, Category, ScopeRef 없음. finding 간 동일성 비교나 라운드 간 추적이 불가능.

### BuildReviewPrompt (pkg/spec/prompt.go:11-48)

```go
func BuildReviewPrompt(doc *SpecDocument, codeContext string) string
```

mode 파라미터 없음. 매 호출마다 동일한 open-ended 지시를 생성:
- "Review the SPEC and respond with: 1. VERDICT: PASS, REVISE, or REJECT"
- "For each issue found, write: FINDING: [severity] description"

discover/verify 구분 없이 항상 전수 스캔을 유도. 이것이 재리뷰 시 finding 수 증가의 직접 원인.

### ParseVerdict (pkg/spec/reviewer.go:18-50)

```go
func ParseVerdict(specID, output, provider string, revision int) ReviewResult
```

정규표현식 기반 파싱:
- `verdictRe`: `(?i)VERDICT:\s*(PASS|REVISE|REJECT)`
- `findingRe`: `(?i)FINDING:\s*\[(\w+)]\s*(.+)`

Finding 파싱 시 severity와 description만 추출. ID 없음. prior findings 참조 없음. 새 finding과 기존 finding을 구분할 방법이 없음.

### runSpecReview (internal/cli/spec_review.go:41-131)

- revision을 항상 `0`으로 하드코딩 (102줄: `spec.ParseVerdict(specID, resp.Output, resp.Provider, 0)`)
- REVISE 판정 후 재리뷰 루프 없음 — 판정 출력 후 즉시 종료
- 정적 분석 통합 없음

### MergeVerdicts (pkg/spec/reviewer.go:54-65)

REJECT > REVISE > PASS 우선순위. 단순하고 정확함. 변경 불필요.

### PersistReview (pkg/spec/reviewer.go:68-75)

review.md에 결과 기록. Finding 테이블에 ID/Status 컬럼 추가 필요.

## 설계 결정

### D1: Finding ID 형식 — `F-{seq}` (Revision-Agnostic)

**결정**: 시퀀셜 순번만으로 Finding을 식별한다 (`F-001`, `F-002`). revision 번호는 ID에 포함하지 않고 `FirstSeenRev` 메타데이터로 별도 기록.

**근거**: revision-agnostic ID로 cross-revision 추적이 가능하고, verify 모드에서 prior findings list와의 ID 비교가 성립. UUID보다 가독성이 높고, LLM이 프롬프트에서 참조하기 쉬움.

**대안 검토**:
- `F-{revision}-{seq}`: revision 컨텍스트가 내장되지만, cross-revision 추적 시 ID 비교가 복잡해짐 (revision 1에서 spec.md D1 결정으로 기각)
- UUID: 충돌 방지는 좋지만 LLM 프롬프트에 삽입 시 토큰 낭비, 사람이 읽기 어려움
- 해시 기반: description 해시로 동일 finding 자동 감지 가능하지만, 문구 변형 시 실패

### D2: Mode-Aware Prompt — 시그니처 확장 vs 별도 함수

**결정**: `BuildReviewPrompt(doc, codeContext, opts ReviewPromptOptions)` 형태로 opts 구조체에 mode, priorFindings, staticFindings를 포함한다.

**근거**: 단일 진입점 유지 + 향후 확장 용이. 옵션 구조체의 zero value가 기존 discover 동작과 동일하므로 하위 호환.

**대안 검토**:
- `BuildDiscoverPrompt` / `BuildVerifyPrompt` 분리: 명시적이지만 공통 로직(SPEC 섹션 렌더링) 중복
- mode를 string으로: 타입 안전성 부족, 오타 위험

### D3: Scope Lock 강도 — Hard Block vs Soft Tag

**결정**: out_of_scope 태깅 + verdict 제외 (soft). 단, review.md에는 기록하여 가시성 유지.

**근거**: hard block(완전 삭제)은 진짜 중요한 새 발견을 놓칠 수 있음. soft tag는 수렴성을 보장하면서도 사람이 review.md에서 out_of_scope finding을 확인하고 판단 가능.

**대안 검토**:
- Hard block (완전 삭제): 수렴성 최대화지만 중요 발견 유실 위험
- 경고만: 수렴성 개선 없음, 현재와 동일한 문제

### D4: 정적 분석 통합 — 사전 주입 vs 병렬 실행

**결정**: Phase A에서 정적 분석을 먼저 실행하고, 결과를 LLM 프롬프트에 "이미 발견된 이슈"로 주입한다.

**근거**: LLM이 린트 이슈를 재발견하는 것은 토큰 낭비이자 finding 증가의 원인. 사전 주입으로 LLM이 더 고수준 이슈(correctness, feasibility)에 집중하도록 유도.

**대안 검토**:
- 병렬 실행 후 중복 제거: 구현 복잡도 높음, 중복 판단 기준 모호
- 정적 분석만 별도 단계: 통합 안 하면 LLM이 여전히 재발견

### D5: Circuit Breaker 기준 — finding 수 단순 비교

**결정**: open finding 수가 이전 라운드보다 증가하면 중단.

**근거**: 수렴하는 리뷰는 finding 수가 단조 감소해야 함. 증가는 scope creep 또는 프롬프트 문제의 신호. 단순하고 예측 가능한 기준.

**대안 검토**:
- severity 가중치 기반: 복잡도 대비 이점 불분명
- 연속 2회 증가 시: 더 관대하지만 수렴 지연

## 리스크 분석

### HIGH: Phase A 완전성 과신

정적 분석이 모든 스타일 이슈를 잡는다고 가정하면 위험. golangci-lint가 잡지 못하는 코드 패턴도 있음.

**완화**: Phase A를 "best-effort discovery"로 정의. LLM에게 "정적 분석이 놓친 스타일 이슈도 보고 가능하되, 이미 보고된 것은 재보고 금지"로 지시.

### HIGH: Finding ID 안정성

LLM이 `F-{seq}` 형식을 정확히 따르지 않을 수 있음.

**완화**: ParseVerdict에서 정규표현식으로 ID 파싱. 매칭 실패 시 자동 ID 할당(`F-{auto-seq}`). 테스트에서 다양한 출력 형식 커버.

### MEDIUM: Scope lock 부작용

verify 모드에서 requirement 자체의 변경이 새 finding을 필요로 하는 경우, scope lock이 이를 차단할 수 있음.

**완화**: scope를 requirement + acceptance criteria까지 확장. requirement가 변경된 경우 해당 requirement의 ScopeRef를 가진 finding은 re-open 가능하도록 설계.

### LOW: REVISE 루프와 orchestra 비용

매 iteration마다 RunOrchestra를 재실행하므로 multi-provider 비용이 iteration 수에 비례.

**완화**: 기본 max iteration = 3. Circuit breaker로 불필요한 반복 조기 중단.
