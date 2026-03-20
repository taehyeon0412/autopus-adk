package content

// GenerateAXInstruction은 @AX 코드 어노테이션 규칙 지침을 생성한다.
func GenerateAXInstruction() string {
	return `# @AX 코드 어노테이션 규칙

@AX 태그는 AI 에이전트가 코드 컨텍스트, 불변 계약, 위험 구역을 세션 간에 전달하는 어노테이션 시스템입니다.

## 태그 타입

### @AX:NOTE
컨텍스트와 의도를 전달합니다.

**추가 시점:**
- 매직 상수 발견 시
- 내보낸 함수에 godoc이 없고 100줄을 초과할 때
- 비즈니스 규칙이 설명되지 않을 때

예시:
` + "```" + `go
// @AX:NOTE: 이 상수는 결제 서비스 SLA에서 정의된 값입니다.
const paymentTimeout = 30 * time.Second
` + "```" + `

### @AX:WARN
위험 구역을 표시합니다. @AX:REASON 필수.

**추가 시점:**
- context.Context 없는 goroutine/channel
- 순환 복잡도 >= 15
- 전역 상태 변경 감지
- if 분기 >= 8

예시:
` + "```" + `go
// @AX:WARN: 고루틴 누수 위험 — 컨텍스트 취소 미처리
// @AX:REASON: 레거시 코드, SPEC-REFACTOR-001에서 수정 예정
go func() { ... }()
` + "```" + `

### @AX:ANCHOR
불변 계약을 표시합니다. @AX:REASON 필수. 자동 삭제 금지.

**추가 시점:**
- 함수 fan_in >= 3 호출자
- 공개 API 경계 식별
- 외부 시스템 통합 지점

예시:
` + "```" + `go
// @AX:ANCHOR: 이 함수 시그니처 변경 금지 — 3+ 컨슈머 의존
// @AX:REASON: PaymentService, BillingHandler, WebhookProcessor가 사용
func ProcessPayment(ctx context.Context, req PaymentRequest) (*PaymentResult, error) {
` + "```" + `

### @AX:TODO
미완성 작업을 표시합니다.

**추가 시점:**
- 공개 함수에 테스트 파일 없을 때
- SPEC 요구사항 미구현 시
- 에러가 처리 없이 반환될 때

예시:
` + "```" + `go
// @AX:TODO: 입력 검증 추가 필요 — SPEC-AUTH-001 참조
func ValidateToken(token string) bool {
` + "```" + `

## 라이프사이클 규칙

### ANCHOR
- fan_in >= 3일 때 생성
- 호출자 수 또는 SPEC 변경 시 업데이트
- fan_in < 3으로 감소 시 NOTE로 다운그레이드 (보고서 필요)
- **절대 자동 삭제 금지**

### WARN
- 위험 구조 발견 시 생성
- 위험이 개선되면 삭제 가능
- 구조적 위험 (예: 고루틴 라이프사이클)은 영속적

### TODO
- RED/ANALYZE 단계에서 생성
- GREEN/IMPROVE 단계에서 삭제
- 3회 이상 미해결 시 WARN으로 에스컬레이션

### NOTE
- 컨텍스트 필요 시 생성
- 함수 시그니처 변경 후 재검토
- 코드 삭제 시 함께 삭제

## 파일당 제한

- @AX:ANCHOR: 파일당 최대 3개
- @AX:WARN: 파일당 최대 5개

초과 시:
- ANCHOR: 가장 낮은 fan_in 초과분 다운그레이드
- WARN: P1-P5 최고 우선순위만 유지

## 필수 필드

- **@AX:REASON**: WARN, ANCHOR 태그에 **필수**
- **[AUTO] 접두사**: 에이전트가 생성한 태그에 **필수**
- **@AX:SPEC**: SPEC 존재 시 포함 (선택)

## 언어별 주석 문법

| 언어 | 접두사 |
|------|--------|
| Go, Java, TS, Rust | ` + "`//`" + ` |
| Python, Ruby | ` + "`#`" + ` |
| Haskell | ` + "`--`" + ` |
`
}
