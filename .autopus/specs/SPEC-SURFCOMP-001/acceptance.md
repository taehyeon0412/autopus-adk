# SPEC-SURFCOMP-001 수락 기준

## 시나리오

### S1: SignalCapable 인터페이스 감지 (R5, R6)

- Given: CmuxAdapter 인스턴스가 Terminal 인터페이스로 전달됨
- When: `term.(SignalCapable)` 타입 어설션을 수행
- Then: 어설션이 성공하고 SurfaceHealth, WaitForSignal, SendSignal 메서드가 사용 가능

### S2: PlainAdapter에서 SignalCapable 미지원 (R5)

- Given: PlainAdapter 인스턴스가 Terminal 인터페이스로 전달됨
- When: `term.(SignalCapable)` 타입 어설션을 수행
- Then: 어설션이 실패하여 ok=false 반환

### S3: CmuxAdapter SurfaceHealth 정상 동작 (R6)

- Given: CmuxAdapter와 유효한 paneID가 주어짐
- When: SurfaceHealth(ctx, paneID)를 호출
- Then: `cmux surface-health` 명령이 실행되고 SurfaceStatus{Valid: true, InWindow: true}가 반환됨

### S4: CmuxAdapter SurfaceHealth stale surface 감지 (R6)

- Given: CmuxAdapter와 이미 종료된 paneID가 주어짐
- When: SurfaceHealth(ctx, paneID)를 호출
- Then: SurfaceStatus{Valid: false}가 반환되거나 에러 반환

### S5: SignalDetector 즉시 완료 탐지 (R3)

- Given: SignalDetector가 cmux 터미널과 함께 초기화됨
- When: 프로바이더가 `cmux wait-for -S "done-claude"` 신호를 전송
- Then: WaitForCompletion이 즉시 (< 500ms) true를 반환

### S6: SignalDetector 타임아웃 시 ScreenPollDetector 폴백 (R3, R4)

- Given: SignalDetector가 초기화되었으나 프로바이더 후크가 신호를 전송하지 않음
- When: WaitForCompletion의 타임아웃이 만료됨
- Then: ScreenPollDetector로 자동 폴백하여 2-phase 매칭으로 완료 탐지 시도

### S7: ScreenPollDetector 2-phase 연속 매칭 (R4)

- Given: ScreenPollDetector가 비-cmux 터미널과 함께 초기화됨
- When: ReadScreen에서 프롬프트 패턴이 2회 연속 감지됨
- Then: WaitForCompletion이 true를 반환

### S8: ScreenPollDetector baseline 필터링 (R4)

- Given: ScreenPollDetector에 이전 라운드 baseline이 설정됨
- When: ReadScreen 결과가 baseline과 동일
- Then: 해당 읽기를 무시하고 candidateDetected를 리셋

### S9: SurfaceManager 백그라운드 헬스 모니터링 (R1)

- Given: SurfaceManager가 3개 프로바이더 pane과 함께 Start됨
- When: 5초 경과
- Then: 모든 pane에 대해 SurfaceHealth가 최소 1회 호출되고 결과가 캐시됨

### S10: SurfaceManager stale surface 사전 감지 (R1)

- Given: SurfaceManager가 실행 중이고 하나의 surface가 stale 상태
- When: executeRound()가 Round 2를 시작
- Then: SurfaceManager.IsHealthy(paneID)가 false를 반환하여 validateSurface ReadScreen 호출 전에 감지

### S11: Surface 복구 후 baseline 재캡처 (R7)

- Given: Round 2에서 surface가 stale로 감지되어 recreatePane()이 실행됨
- When: 새 pane이 생성되고 CLI 세션이 준비됨
- Then: 해당 프로바이더의 baselines map이 새 pane의 ReadScreen 결과로 즉시 갱신됨

### S12: CompletionDetector 자동 선택 (R2)

- Given: NewCompletionDetector 팩토리에 CmuxAdapter가 전달됨
- When: CmuxAdapter가 SignalCapable을 구현
- Then: SignalDetector가 반환됨

### S13: CompletionDetector 폴백 선택 (R2)

- Given: NewCompletionDetector 팩토리에 PlainAdapter가 전달됨
- When: PlainAdapter가 SignalCapable을 구현하지 않음
- Then: ScreenPollDetector가 반환됨

### S14: 완료 후크 템플릿 신호 전송 (R8)

- Given: claude 프로바이더 CLI 세션이 응답을 완료
- When: 완료 후크가 실행됨
- Then: `cmux wait-for -S "done-claude"` 명령이 실행되어 오케스트레이터에 신호 전달

### S15: Warm Pool swap 복구 (R9, P1)

- Given: SurfaceManager에 1개의 warm spare surface가 준비됨
- When: 활성 surface가 stale로 감지됨
- Then: warm spare와 즉시 교체(swap)되어 복구 시간이 2초 이내

### S16: 프로바이더별 Idle Threshold (R10, P1)

- Given: autopus.yaml에 claude의 idle_threshold가 45s로 설정됨
- When: ScreenPollDetector가 claude의 완료를 감지
- Then: idle fallback threshold로 45s가 사용됨 (기본 30s 대신)

### S17: 동적 타임아웃 재분배 (R11, P1)

- Given: 3라운드 debate, 총 타임아웃 180s, Round 1이 30s에 완료
- When: Round 2의 perRound 타임아웃을 계산
- Then: 남은 150s를 2라운드에 재분배하여 Round 2 타임아웃이 75s

### S18: 파일 크기 제한 준수

- Given: 모든 신규 및 수정 소스 코드 파일
- When: 코드 리뷰 수행
- Then: 모든 .go 파일이 300 lines 미만이고, 200 lines 미만을 목표
