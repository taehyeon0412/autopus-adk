# SPEC-REVCONV-001 수락 기준

## 시나리오

### S1: Finding에 ID와 Status가 부여됨
- Given: 리뷰어가 SPEC에 대해 finding을 생성한다
- When: ParseVerdict가 provider 출력을 파싱한다
- Then: 각 ReviewFinding에 `F-{seq}` 형식의 revision-agnostic ID가 부여된다
- And: Status는 `open`으로 초기화된다
- And: Category와 ScopeRef가 파싱된 값으로 설정된다

### S2: Discover 모드 프롬프트 생성
- Given: SPEC 문서와 코드 컨텍스트가 존재한다
- When: BuildReviewPrompt가 mode=`discover`로 호출된다
- Then: 전수 스캔 지시가 포함된 프롬프트가 생성된다
- And: 요구사항, 인수 기준, 코드 컨텍스트가 모두 포함된다
- And: finding 출력 형식에 ID, Category, ScopeRef 필드가 지시된다

### S3: Verify 모드 프롬프트 생성
- Given: 이전 라운드의 open finding 목록이 존재한다
- When: BuildReviewPrompt가 mode=`verify`로 호출된다
- Then: prior findings가 체크리스트로 포함된 프롬프트가 생성된다
- And: "새 finding을 탐색하지 말고 기존 finding의 해결 여부만 검증하라"는 지시가 포함된다

### S4: Scope Lock — 새 finding 차단
- Given: verify 모드에서 prior findings가 [F-001, F-002]이다
- When: 리뷰어가 F-001, F-002, F-003(신규, minor/completeness)을 포함한 응답을 반환한다
- Then: F-003의 Status가 `out_of_scope`로 태깅된다
- And: F-003은 verdict 계산에서 제외된다

### S5: Circuit Breaker 작동
- Given: Revision 1에서 open finding이 3개이다
- When: Revision 2 완료 후 open finding이 4개로 증가한다
- Then: REVISE 루프가 즉시 중단된다
- And: 경고 메시지가 stderr에 출력된다
- And: 현재 상태가 최종 결과로 반환된다

### S6: REVISE 루프 정상 수렴
- Given: Revision 0(discover)에서 verdict가 REVISE이고 finding이 3개이다
- When: SPEC이 수정된 후 Revision 1(verify) 리뷰가 실행된다
- Then: verify 모드 프롬프트가 3개 finding의 체크리스트와 함께 생성된다
- And: finding이 모두 resolved되면 verdict가 PASS로 전환된다

### S7: REVISE 루프 최대 반복 제한
- Given: 최대 반복 횟수가 3으로 설정되어 있다
- When: 3회 반복 후에도 verdict가 REVISE이다
- Then: 루프가 중단되고 현재 상태가 최종 결과로 반환된다
- And: "최대 반복 횟수 도달" 경고가 출력된다

### S8: 정적 분석 Finding 사전 주입
- Given: config에 static_analysis가 활성화되어 있고 golangci-lint가 설치되어 있다
- When: discover 모드 리뷰가 시작된다
- Then: golangci-lint 실행 결과가 Category=`style`의 ReviewFinding으로 변환된다
- And: LLM 리뷰 프롬프트에 "이미 발견된 정적 분석 이슈" 섹션으로 주입된다
- And: LLM이 동일 이슈를 재발견하지 않도록 지시가 포함된다

### S9: 정적 분석 바이너리 미설치 시 graceful skip
- Given: config에 static_analysis가 활성화되어 있지만 바이너리가 없다
- When: discover 모드 리뷰가 시작된다
- Then: stderr에 경고가 출력된다
- And: 정적 분석 없이 LLM 리뷰만 진행된다

### S10: 하위 호환성
- Given: 기존 코드가 ReviewFinding을 ID/Status 없이 생성한다
- When: 해당 Finding이 포함된 ReviewResult가 처리된다
- Then: ID가 빈 문자열이어도 정상 동작한다
- And: PersistReview가 review.md에 기존 형식과 호환되는 출력을 생성한다
