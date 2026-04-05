# SPEC-REVCONV-001 구현 계획

## 태스크 목록

### Phase 1: 타입 시스템 및 핵심 로직 (즉시)

- [ ] T1: ReviewFinding 타입 확장 — ID, Status, Category, ScopeRef 필드 추가 (types.go)
- [ ] T2: ReviewMode 타입 정의 — `discover` / `verify` 모드 상수 및 ReviewPromptOptions 구조체 (types.go)
- [ ] T3: BuildReviewPrompt 모드 분기 — mode 파라미터 추가, discover/verify 템플릿 분리 (prompt.go)
- [ ] T4: ParseVerdict 확장 — 구조화 Finding 파싱(ID 포함), priorFindings 기반 scope filtering (reviewer.go)
- [ ] T5: Circuit Breaker 로직 — finding 수 증가 감지 시 루프 중단 판단 함수 (reviewer.go)
- [ ] T6: 정적 분석 통합 모듈 — golangci-lint 실행, 출력 파싱, ReviewFinding 변환 (static_analysis.go, 신규)
- [ ] T6.5: Findings 영속화 모듈 — review-findings.json 읽기/쓰기, ScopeRef 정규화, finding dedup (findings.go, 신규)
- [ ] T7: REVISE 루프 구현 — spec_review.go에서 discover→verify 전환, 최대 반복 제한, PASS/REVISE verdict 분기 (spec_review.go)

### Phase 2: 테스트 및 검증

- [ ] T8: 단위 테스트 — T1~T5에 대한 테스트 케이스 (reviewer_test.go, types_test.go)
- [ ] T9: 정적 분석 통합 테스트 — mock 린터 출력 기반 Finding 변환 검증 (static_analysis_test.go)
- [ ] T10: 통합 테스트 — REVISE 루프 전체 플로우 검증 (spec_review_test.go)

## 구현 전략

### 접근 방법

1. **하위 호환성 유지**: ReviewFinding의 새 필드는 모두 선택적(zero value가 유효). 기존 코드가 ID/Status 없이 생성한 Finding도 정상 동작.
2. **BuildReviewPrompt 시그니처 변경**: `BuildReviewPrompt(doc, codeContext)` → `BuildReviewPrompt(doc, codeContext, opts ReviewPromptOptions)`. 기존 호출부(spec_review.go:76)만 수정하면 됨.
3. **ParseVerdict 시그니처 변경**: `ParseVerdict(specID, output, provider, revision)` → `ParseVerdict(specID, output, provider, revision, priorFindings)`. priorFindings가 nil이면 기존 동작(discover 모드).
4. **정적 분석은 옵트인**: config에 `review_gate.static_analysis` 키가 없으면 스킵. 정적 분석 바이너리 미설치 시에도 경고만 출력하고 계속 진행.

### 기존 코드 활용

- `spec.CollectContext`(prompt.go:51): 그대로 활용, verify 모드에서도 코드 컨텍스트 제공
- `spec.MergeVerdicts`(reviewer.go:54): 그대로 활용, 각 프로바이더 verdict 병합
- `orchestra.RunOrchestra`(spec_review.go:94): REVISE 루프의 각 iteration에서 재호출
- `detect.IsInstalled`(spec_review.go:138): 정적 분석 바이너리 존재 확인에 재활용

### 변경 범위

- **수정 파일 4개**: types.go, prompt.go, reviewer.go, spec_review.go
- **신규 파일 2개**: static_analysis.go, static_analysis_test.go
- **테스트 수정 2개**: reviewer_test.go, spec_review_test.go (또는 types_test.go)
- 각 파일 300줄 미만 유지 확인 필요 (현재 모두 150줄 미만이므로 여유 있음)

### 의존성

- T1, T2는 독립적으로 병렬 실행 가능
- T3는 T2(ReviewMode 타입) 완료 후
- T4, T5는 T1(Finding ID/Status) 완료 후
- T6는 T1 완료 후 독립 실행 가능
- T6.5는 T1 완료 후 독립 실행 가능 (T6과 병렬)
- T7은 T3, T4, T5, T6, T6.5 모두 완료 후
- T8~T10은 각각 대응하는 구현 태스크 완료 후
