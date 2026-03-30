# SPEC-SURFCOMP-001 구현 계획

## 태스크 목록

### Phase 1: Terminal 인터페이스 확장 (R5, R6)

- [ ] T1: `pkg/terminal/surface_health.go` — SurfaceStatus 타입 정의 및 SignalCapable 옵셔널 인터페이스 정의
  - SurfaceStatus struct: Valid bool, SurfaceRef string, InWindow bool
  - SignalCapable interface: SurfaceHealth, WaitForSignal, SendSignal 메서드
  - 기존 Terminal 인터페이스는 변경하지 않고, 옵셔널 인터페이스 패턴 사용 (type assertion)
  - 예상: ~60 lines

- [ ] T2: `pkg/terminal/cmux_signal.go` — CmuxAdapter의 SignalCapable 구현
  - SurfaceHealth: `cmux surface-health` 실행 + 출력 파싱
  - WaitForSignal: `cmux wait-for {name}` 실행 (타임아웃 포함)
  - SendSignal: `cmux wait-for -S {name}` 실행
  - 예상: ~80 lines

- [ ] T3: `pkg/terminal/tmux.go` 및 `pkg/terminal/plain.go`에 no-op SignalCapable 구현 (P2 tmux 지원 예약)
  - TmuxAdapter: ErrNotSupported 반환 (P2에서 tmux wait-for 구현)
  - PlainAdapter: ErrNotSupported 반환
  - 예상: 각 ~20 lines 추가

### Phase 2: CompletionDetector 추상화 (R2, R3, R4)

- [ ] T4: `pkg/orchestra/completion_detector.go` — CompletionDetector 인터페이스 + 팩토리
  - CompletionDetector interface: WaitForCompletion(ctx, paneInfo, patterns, baseline, round) (bool, error)
  - NewCompletionDetector(term Terminal) CompletionDetector — 터미널 타입에 따라 Signal/Poll 자동 선택
  - 예상: ~50 lines

- [ ] T5: `pkg/orchestra/completion_signal.go` — SignalDetector 구현
  - cmux wait-for "done-{provider}" 블로킹 대기
  - 타임아웃 시 ScreenPollDetector로 폴백
  - round 파라미터로 "done-{provider}-round{N}" 신호명 생성
  - 예상: ~80 lines

- [ ] T6: `pkg/orchestra/completion_poll.go` — ScreenPollDetector 구현
  - 기존 interactive_completion.go의 waitForCompletion() 로직 이동
  - 2-phase consecutive match + idle fallback 유지
  - CompletionDetector 인터페이스 준수
  - 예상: ~90 lines

- [ ] T7: `pkg/orchestra/interactive_completion.go` 리팩토링
  - waitForCompletion()을 CompletionDetector 디스패치 wrapper로 변환
  - 기존 호출자(waitAndCollectResults 등) 호환성 유지
  - 예상: ~30 lines (기존 74 lines에서 축소)

### Phase 3: SurfaceManager (R1, R7, R9)

- [ ] T8: `pkg/orchestra/surface_manager.go` — SurfaceManager 구현
  - Start(ctx, panes []paneInfo): 백그라운드 헬스 모니터링 goroutine 시작
  - Stop(): 모니터링 중지
  - IsHealthy(paneID) bool: 캐시된 헬스 상태 반환
  - SwapIfStale(ctx, pi paneInfo, cfg OrchestraConfig, round int) (paneInfo, error): warm pool에서 교체
  - 5초 간격 surface-health 폴링 (SurfaceHealth via SignalCapable)
  - 예상: ~150 lines

- [ ] T9: `pkg/orchestra/interactive_surface.go` 수정
  - validateSurface()를 SurfaceManager.IsHealthy()로 교체
  - recreatePane()에 warm pool swap 경로 추가 (P1 R9)
  - 복구 후 baseline 재캡처 보장 (R7)
  - 예상: ~100 lines (기존 107 lines에서 유사)

- [ ] T10: `pkg/orchestra/interactive_debate.go` 수정
  - executeRound()에서 SurfaceManager + CompletionDetector 사용
  - 라운드간 동적 타임아웃 재분배 (P1 R11)
  - 예상: 변경 범위 ~50 lines

### Phase 4: 프로바이더 후크 & 설정 (R8, R10)

- [ ] T11: `templates/hooks/completion-hook.sh.tmpl` — 완료 후크 템플릿
  - `cmux wait-for -S "done-{provider}"` 신호 전송
  - claude-hook, gemini 설정 예시 포함
  - 예상: ~30 lines

- [ ] T12: `pkg/orchestra/types.go` 수정
  - OrchestraConfig에 CompletionDetector, SurfaceManager 필드 추가
  - ProviderConfig에 IdleThreshold time.Duration 필드 추가 (R10)
  - 예상: ~10 lines 추가

### Phase 5: 테스트 (전 Phase 병행)

- [ ] T13: `pkg/terminal/cmux_signal_test.go` — CmuxAdapter SignalCapable 단위 테스트
- [ ] T14: `pkg/orchestra/completion_detector_test.go` — 팩토리 및 인터페이스 테스트
- [ ] T15: `pkg/orchestra/completion_signal_test.go` — SignalDetector 테스트 (mock terminal)
- [ ] T16: `pkg/orchestra/completion_poll_test.go` — ScreenPollDetector 테스트 (기존 테스트 마이그레이션)
- [ ] T17: `pkg/orchestra/surface_manager_test.go` — SurfaceManager 헬스 모니터링 + swap 테스트
- [ ] T18: `pkg/orchestra/interactive_surface_test.go` 기존 테스트 업데이트

## 구현 전략

### 접근 방법

1. **옵셔널 인터페이스 패턴**: Terminal 인터페이스를 변경하지 않고, `SignalCapable` 인터페이스를 별도 정의. 호출부에서 `if sc, ok := term.(SignalCapable); ok { ... }` 패턴으로 기능 감지. 기존 Terminal 구현체(plain, tmux)의 호환성 완전 보장.

2. **전략 패턴**: CompletionDetector 인터페이스로 완료 탐지 전략을 교체 가능하게 구성. 팩토리 함수가 터미널 타입에 따라 최적 전략 자동 선택.

3. **점진적 마이그레이션**: 기존 waitForCompletion() 로직을 ScreenPollDetector로 이동 후, 원래 함수는 thin wrapper로 유지. 호출자 변경 최소화.

4. **Warm Pool은 P1**: Phase 3에서 SurfaceManager 구조는 설계하되, warm pool 로직은 P1로 분리. P0에서는 IsHealthy() + recreatePane() fallback으로 동작.

### 기존 코드 활용

- `interactive_completion.go`의 waitForCompletion() → ScreenPollDetector로 이동
- `interactive_surface.go`의 validateSurface()/recreatePane() → SurfaceManager 연동
- `hook_signal.go`의 RoundSignalName() → SignalDetector에서 재사용
- `interactive_detect.go`의 isPromptVisible(), isProviderWorking() → ScreenPollDetector에서 그대로 사용

### 변경 범위

- 신규 파일 7개 (소스 코드 5 + 테스트 6 + 템플릿 1)
- 수정 파일 5개 (terminal.go 제외 — 옵셔널 인터페이스로 변경 불필요)
- 총 신규 코드량: ~540 lines (테스트 제외)
- 모든 신규 파일은 200 lines 미만 목표, 300 lines 하드 리밋 준수

### 의존 관계

```
T1 → T2, T3 (Terminal 인터페이스 확장 먼저)
T4 → T5, T6, T7 (인터페이스 정의 후 구현)
T8 → T9, T10 (SurfaceManager 먼저)
T12 → T10 (types 변경 후 debate 수정)
T11은 독립 실행 가능
```

### 병렬 실행 가능 태스크

- Phase 1 (T1-T3) 과 Phase 4의 T11, T12는 병렬 가능
- Phase 2 (T4-T7) 내에서 T5와 T6은 T4 완료 후 병렬 가능
- Phase 5 테스트는 각 Phase 완료 후 즉시 병행
