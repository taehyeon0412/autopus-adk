# SPEC-LEARNWIRE-002 수락 기준

## 시나리오

### S1: Gate 실패 시 자동 기록

- Given: LearnStore가 설정된 SequentialRunner와 GateValidation이 적용된 Phase
- When: PhaseBackend가 "FAIL" 출력을 반환하여 gate가 VerdictFail을 반환
- Then: learn store의 pipeline.jsonl에 type "gate_fail"인 엔트리가 추가됨

### S2: nil Store일 때 동작 불변

- Given: LearnStore가 nil인 SequentialRunner
- When: Gate가 VerdictFail을 반환
- Then: 기존과 동일하게 retry가 진행되며, 어떤 에러도 발생하지 않음

### S3: Max Retry 초과 시 critical 기록

- Given: LearnStore가 설정되고 MaxRetries=1인 Phase
- When: 모든 retry가 실패하여 max retries 초과 에러 반환
- Then: severity "critical"인 gate_fail 엔트리가 기록된 후 에러가 반환됨

### S4: Executor Error 자동 기록

- Given: LearnStore가 설정된 Runner
- When: PhaseBackend.Execute()가 에러를 반환
- Then: type "executor_error"인 엔트리가 기록된 후 에러가 전파됨

### S5: Coverage Gap 파싱 및 기록

- Given: LearnStore가 설정된 Runner와 PhaseValidate
- When: 출력에 "coverage: 62.5%" 패턴이 포함되고 gate가 실패
- Then: type "coverage_gap"인 엔트리가 커버리지 수치와 함께 기록됨

### S6: Review Issue 파싱 및 기록

- Given: LearnStore가 설정된 Runner와 PhaseReview
- When: 출력에 "REQUEST_CHANGES" 마커가 포함되고 gate가 실패
- Then: type "review_issue"인 엔트리가 리뷰 피드백과 함께 기록됨

### S7: ParallelRunner 동시 실패 독립 기록

- Given: LearnStore가 설정된 ParallelRunner와 복수의 실패 Phase
- When: 여러 Phase가 동시에 gate fail
- Then: 각 실패가 독립적으로 기록되며, 기록 간 데이터 경쟁이 없음

### S8: 기존 테스트 통과

- Given: 현재 pkg/pipeline/ 패키지의 모든 테스트
- When: SPEC-LEARNWIRE-002 변경 적용 후 go test 실행
- Then: 기존 테스트가 모두 통과 (regression 없음)
