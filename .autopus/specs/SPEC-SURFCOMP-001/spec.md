# SPEC-SURFCOMP-001: cmux Surface Lifecycle & Completion Detection Overhaul

**Status**: completed
**Created**: 2026-03-30
**Domain**: SURFCOMP
**Module**: autopus-adk

## 목적

Round 2+ 다중 프로바이더 토론(debate) 오케스트레이션의 핵심 신뢰성 문제 3가지를 해결한다:

1. **Reactive Surface Recovery**: `validateSurface()`가 라운드 경계에서만 검사하며, `recreatePane()`이 전체 surface + CLI 세션을 파괴/재생성하여 15-20초 소요 + 2초 하드코딩 딜레이 해킹
2. **Polling-Based Completion Detection**: 2초 간격 `ReadScreen()` 폴링 + regex 매칭으로 false positive (중간 `>` 문자)와 false negative (ANSI escape 오염) 발생, 60초 idle fallback은 과도
3. **Baseline Fragility**: 베이스라인 캡처 타이밍이 surface 재생성 및 프롬프트 전송과 경합하여 Round 2+에서 false-positive 완료 탐지 유발

cmux v0.62.2 API 테스트 결과, `cmux surface-health` (경량 헬스체크)와 `cmux wait-for` (신호 기반 동기화)가 이 문제들의 근본 해결책임을 확인했다.

## 요구사항

### P0 — Must Have

**R1: SurfaceManager — 사전 헬스 모니터링**
WHILE interactive debate가 활성 상태인 동안, THE SYSTEM SHALL `cmux surface-health` 명령을 사용하여 5초 간격으로 모든 활성 surface의 건강 상태를 백그라운드에서 모니터링해야 한다. ReadScreen 호출 없이 surface 유효성을 판단하여, 라운드 경계 이전에 stale surface를 사전 감지한다.

**R2: CompletionDetector 인터페이스**
WHEN 프로바이더 완료를 감지해야 할 때, THE SYSTEM SHALL `CompletionDetector` 인터페이스를 제공하여 완료 탐지 전략을 교체 가능하게 해야 한다. 인터페이스는 `WaitForCompletion(ctx, paneInfo, round) (bool, error)` 시그니처를 포함한다.

**R3: SignalDetector — cmux wait-for 기반 완료 탐지 (PRIMARY)**
WHEN cmux 터미널에서 프로바이더 완료를 감지해야 할 때, THE SYSTEM SHALL `cmux wait-for "done-{provider}"` 신호를 사용하여 즉각적이고 정확한 완료 탐지를 수행해야 한다. 프로바이더 완료 후크가 `cmux wait-for -S "done-{provider}"` 신호를 전송하면, 오케스트레이터는 해당 신호를 블로킹 대기한다. 폴링 불필요.

**R4: ScreenPollDetector — 폴백 완료 탐지**
WHEN cmux가 아닌 터미널이거나 신호 기반 탐지가 실패할 때, THE SYSTEM SHALL 기존 2-phase 연속 매칭 + idle fallback 로직을 `ScreenPollDetector`로 캡슐화하여 폴백으로 사용해야 한다.

**R5: Terminal 인터페이스 확장**
WHEN cmux 기능을 활용해야 할 때, THE SYSTEM SHALL Terminal 인터페이스에 다음 메서드를 추가해야 한다:
- `SurfaceHealth(ctx, paneID) (SurfaceStatus, error)` — 경량 헬스체크
- `WaitForSignal(ctx, name string, timeout time.Duration) error` — 명명된 신호 블로킹 대기
- `SendSignal(ctx, name string) error` — 명명된 신호 전송

**R6: CmuxAdapter 구현**
WHEN Terminal 확장 메서드가 호출될 때, THE SYSTEM SHALL CmuxAdapter에서 `cmux surface-health`, `cmux wait-for`, `cmux wait-for -S` CLI 명령을 실행하여 구현해야 한다.

**R7: Baseline 재캡처**
WHEN surface가 복구된 후, THE SYSTEM SHALL 해당 프로바이더의 baseline을 즉시 재캡처하여 false-positive 완료 탐지를 방지해야 한다.

**R8: 프로바이더 완료 후크 템플릿**
WHEN 프로바이더 CLI 세션이 응답을 완료했을 때, THE SYSTEM SHALL `cmux wait-for -S "done-{provider}"` 신호를 전송하는 후크 템플릿을 제공해야 한다. 후크는 claude-hook, gemini 등 프로바이더별로 설정 가능해야 한다.

### P1 — Should Have

**R9: Warm Surface Pool**
WHEN surface 장애가 감지되었을 때, THE SYSTEM SHALL 사전 초기화된 여유 surface (최대 1개)와 교체(swap)하여 복구 시간을 2초 이내로 단축해야 한다. 전체 recreatePane() 재생성 대신 warm pool에서 즉시 교체한다.

**R10: 프로바이더별 Idle Threshold 설정**
WHERE autopus.yaml에 프로바이더별 idle threshold가 설정된 경우, THE SYSTEM SHALL 해당 값을 ScreenPollDetector의 idle fallback에 적용해야 한다.

**R11: 라운드간 동적 타임아웃 재분배**
WHEN 이전 라운드가 예상보다 빨리 완료되었을 때, THE SYSTEM SHALL 남은 시간 예산을 후속 라운드에 균등하게 재분배해야 한다.

**R12: FileIPCDetector — fsnotify 기반**
WHEN 후크 지원 프로바이더에서 cmux 없이 운영할 때, THE SYSTEM SHALL fsnotify를 사용하여 신호 파일 생성을 감시하는 FileIPCDetector를 제공해야 한다.

### P2 — Could Have

**R13: TmuxAdapter WaitForSignal 구현**
WHERE tmux 터미널이 사용되는 경우, THE SYSTEM SHALL `tmux wait-for` 명령을 사용하여 WaitForSignal을 구현해야 한다.

**R14: 완료 탐지 메트릭/로깅**
WHEN 완료 탐지가 수행될 때, THE SYSTEM SHALL 사용된 탐지 방법, 소요 시간, 폴백 발생 여부를 구조화된 로그로 기록해야 한다.

**R15: 프로바이더 후크 자동 감지**
WHEN debate가 시작될 때, THE SYSTEM SHALL 각 프로바이더의 후크 지원 여부를 자동 감지하여 최적의 CompletionDetector를 선택해야 한다.

## 생성 파일 상세

### 신규 파일

| 파일 | 역할 |
|------|------|
| `pkg/terminal/surface_health.go` | SurfaceStatus 타입, Terminal 확장 인터페이스 (SurfaceHealth, WaitForSignal, SendSignal) |
| `pkg/terminal/cmux_signal.go` | CmuxAdapter의 SurfaceHealth, WaitForSignal, SendSignal 구현 |
| `pkg/orchestra/surface_manager.go` | SurfaceManager — 백그라운드 헬스 모니터링, warm pool 관리 |
| `pkg/orchestra/completion_detector.go` | CompletionDetector 인터페이스, SignalDetector, ScreenPollDetector |
| `pkg/orchestra/completion_signal.go` | SignalDetector 구현 (cmux wait-for 기반) |
| `pkg/orchestra/completion_poll.go` | ScreenPollDetector 구현 (기존 waitForCompletion 리팩토링) |
| `templates/hooks/completion-hook.sh.tmpl` | 프로바이더 완료 후크 템플릿 |

### 수정 파일

| 파일 | 변경 |
|------|------|
| `pkg/terminal/terminal.go` | SignalCapable 인터페이스 임베딩 또는 옵셔널 인터페이스 패턴 |
| `pkg/terminal/cmux.go` | (변경 없음 — 새 메서드는 cmux_signal.go에 분리) |
| `pkg/terminal/tmux.go` | PlainAdapter용 no-op 또는 P2에서 tmux wait-for 구현 |
| `pkg/orchestra/interactive_surface.go` | SurfaceManager 연동, recreatePane을 swap 기반으로 전환 |
| `pkg/orchestra/interactive_completion.go` | waitForCompletion을 CompletionDetector 디스패치로 교체 |
| `pkg/orchestra/interactive_debate.go` | executeRound에서 CompletionDetector 사용, 동적 타임아웃 |
| `pkg/orchestra/types.go` | OrchestraConfig에 CompletionDetector, SurfaceManager 필드 추가 |
