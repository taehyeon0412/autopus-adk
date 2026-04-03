# SPEC-ROUTE-001 수락 기준

## 시나리오

### S1: TaskConfig 모델 오버라이드 전달
- Given: ClaudeAdapter가 등록되어 있고
- When: TaskConfig.Model이 "claude-3-5-haiku-20241022"로 설정된 상태에서 BuildCommand가 호출되면
- Then: 생성된 exec.Cmd의 Args에 "--model", "claude-3-5-haiku-20241022"가 포함되어야 한다

### S2: 빈 Model 필드는 기본 동작 유지
- Given: ClaudeAdapter가 등록되어 있고
- When: TaskConfig.Model이 빈 문자열인 상태에서 BuildCommand가 호출되면
- Then: 생성된 exec.Cmd의 Args에 "--model" 플래그가 포함되지 않아야 한다

### S3: 단순 메시지 분류
- Given: MessageComplexityClassifier가 기본 설정으로 초기화되어 있고
- When: "현재 상태 확인해줘" (18자, 코드 블록 없음, 단순 키워드)가 입력되면
- Then: 복잡도 레벨은 "simple"이어야 한다

### S4: 복잡 메시지 분류
- Given: MessageComplexityClassifier가 기본 설정으로 초기화되어 있고
- When: 1200자 이상의 코드 블록 포함 메시지에 "리팩토링", "아키텍처" 키워드가 있으면
- Then: 복잡도 레벨은 "complex"이어야 한다

### S5: 복잡도-모델 매핑
- Given: ProviderRouter가 기본 매핑으로 초기화되어 있고
- When: provider="claude", complexity="simple"로 라우팅 요청하면
- Then: 반환 모델은 기본 매핑의 claude simple 모델 (예: haiku)이어야 한다

### S6: 커스텀 매핑 오버라이드
- Given: RoutingConfig에 claude/simple 매핑이 "custom-model-v1"으로 오버라이드되어 있고
- When: provider="claude", complexity="simple"로 라우팅 요청하면
- Then: 반환 모델은 "custom-model-v1"이어야 한다

### S7: 라우팅 비활성화 시 패스스루
- Given: RoutingConfig.Enabled가 false이고
- When: 어떤 메시지든 라우팅을 시도하면
- Then: Model 필드가 빈 문자열로 반환되어 기존 동작이 유지되어야 한다

### S8: 라우팅 결정 로깅
- Given: RoutingConfig.Enabled가 true이고
- When: 메시지가 "complex"로 분류되어 모델이 선택되면
- Then: 로그에 복잡도 레벨, 선택된 모델, 분류 신호가 포함되어야 한다

### S9: Codex 어댑터 모델 플래그
- Given: CodexAdapter가 등록되어 있고
- When: TaskConfig.Model이 "o3"로 설정된 상태에서 BuildCommand가 호출되면
- Then: 생성된 exec.Cmd의 Args에 "--model", "o3"가 포함되어야 한다

### S10: Gemini 어댑터 모델 플래그
- Given: GeminiAdapter가 등록되어 있고
- When: TaskConfig.Model이 "gemini-2.0-flash"로 설정된 상태에서 BuildCommand가 호출되면
- Then: 생성된 exec.Cmd의 Args에 "--model", "gemini-2.0-flash"가 포함되어야 한다

### S11: 중간 복잡도 메시지 분류
- Given: MessageComplexityClassifier가 기본 설정으로 초기화되어 있고
- When: 500자 길이의 "이 함수를 수정해서 에러 처리를 추가해줘" 유형 메시지가 입력되면
- Then: 복잡도 레벨은 "medium"이어야 한다
