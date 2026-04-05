# SPEC-AXCONC-001 수락 기준

## 시나리오

### S1: Per-goroutine context isolation

- Given: 3개의 프로바이더가 설정된 consensus 오케스트레이션
- When: 1개 프로바이더가 timeout으로 cancel되고 나머지 2개가 정상 응답
- Then: 결과에 2개의 성공 응답과 1개의 FailedProvider(timeout)가 반환된다
- Then: 정상 프로바이더의 context는 cancel되지 않았어야 한다

### S2: Per-provider timeout triggers individual cancellation

- Given: runParallel()에 timeout 10초인 config와 sleep 20초인 프로바이더 1개, echo 프로바이더 2개
- When: 오케스트레이션 실행
- Then: sleep 프로바이더만 timeout FailedProvider로 기록
- Then: echo 프로바이더 2개는 정상 ProviderResponse 반환
- Then: 전체 실행 시간이 20초보다 현저히 짧아야 한다 (개별 취소 증거)

### S3: Safety deadline activates on no-deadline context

- Given: deadline이 없는 `context.Background()`로 WaitForCompletion() 호출
- When: completion이 감지되지 않고 10분 경과
- Then: safety deadline에 의해 false 반환 (무한 블로킹 아님)
- Then: 경고 로그가 출력됨

### S4: Safety deadline does NOT activate when caller sets deadline

- Given: 5분 deadline이 설정된 context로 WaitForCompletion() 호출
- When: completion이 감지되지 않음
- Then: caller의 5분 deadline이 적용됨 (10분 safety가 아님)
- Then: safety deadline 경고 로그 미출력

### S5: Backward compatibility of runParallel

- Given: 기존 consensus 전략 호출 코드
- When: 동일한 OrchestraConfig로 RunOrchestra() 호출
- Then: 함수 시그니처 변경 없음
- Then: 기존 테스트 전체 통과

### S6: Backward compatibility of WaitForCompletion

- Given: 기존 CompletionDetector 인터페이스 구현
- When: WaitForCompletion() 호출
- Then: 인터페이스 시그니처 변경 없음
- Then: 기존 completion_poll_test.go 테스트 전체 통과
