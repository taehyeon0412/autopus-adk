# SPEC-SURFCOMP-001 리서치

## 기존 코드 분석

### Surface Lifecycle (현재 구현)

**`pkg/orchestra/interactive_surface.go`** (107 lines)

- `validateSurface()` (L22-25): ReadScreen 호출로 surface 유효성 판단. 경량이 아님 — 전체 화면 내용 읽기.
- `recreatePane()` (L32-106): PipePaneStop → Close → SplitPane → CreateTemp → PipePaneStart (3회 재시도) → SendLongText (CLI 재실행) → pollUntilSessionReady → **time.Sleep(2s)** 하드코딩 딜레이.
  - 이 2초 딜레이는 cmux surface 초기화 타이밍 이슈를 해킹으로 해결한 것. paste-buffer가 새 surface에서 exit status 1로 실패하는 문제 방지.
  - 전체 복구 시간: SplitPane (~1s) + PipePaneStart 재시도 (~3s) + CLI 실행 (~3-5s) + session ready 대기 (~5-15s) + 2s sleep = **14-26초**.

**호출 경로**: `interactive_debate.go:executeRound()` (L167-183) → Round 2+에서만 `validateSurface()` 호출 → 실패 시 `recreatePane()`.

### Completion Detection (현재 구현)

**`pkg/orchestra/interactive_completion.go`** (74 lines)

- `waitForCompletion()` (L23-73): 2초 간격 ticker + ReadScreen 폴링.
  - Phase 1: `isPromptVisible()` 호출 — 첫 번째 매칭
  - Phase 2: 다음 tick에서 재확인 — 2회 연속 매칭으로 확정 (false positive 방지)
  - Idle fallback: 60초 동안 2-phase 실패 시 → outputFile의 mtime 기반 30초 idle 체크
  - 문제점: ReadScreen의 ANSI escape 오염으로 regex 매칭 실패, 프로바이더 출력 중 중간 `>` 문자로 false positive

**`pkg/orchestra/interactive_detect.go`** (253 lines)

- `isPromptVisible()` (L165-183): ANSI strip 후 CompletionPattern + defaultPromptPatterns regex 매칭
- `isProviderWorking()` (L218-226): 진행 상태 패턴 ("Generating", "Thinking" 등) 감지 — idle fallback 억제용
- `isOutputIdle()` (L230-236): outputFile의 mtime 기반 idle 판단

### Hook Signal Protocol (기존 File IPC)

**`pkg/orchestra/hook_signal.go`** (222 lines)

- `HookSession`: `/tmp/autopus/{session-id}/` 디렉토리 기반 파일 신호 프로토콜
- `WaitForDone()`: 200ms 간격 파일 폴링 (os.Stat)
- `WaitForDoneRoundCtx()`: 라운드별 done 파일 대기
- `RoundSignalName()`: `{provider}-round{N}-{suffix}` 형식 생성

**`pkg/orchestra/interactive_debate_ipc.go`** (31 lines)

- `tryFileIPC()`: WaitForReady → WriteInputRound → 성공 시 SendLongText 스킵

### Terminal Interface

**`pkg/terminal/terminal.go`** (52 lines)

- `Terminal` 인터페이스: 11개 메서드 (Name, CreateWorkspace, SplitPane, SendCommand, SendLongText, Notify, ReadScreen, PipePaneStart, PipePaneStop, Close)
- `@AX:ANCHOR` 마커: "core public API contract — all adapters (cmux, tmux, plain) implement this interface"
- **설계 결정**: AX:ANCHOR가 이 인터페이스를 stable boundary로 표시 → 새 메서드를 직접 추가하면 모든 어댑터 파괴. **옵셔널 인터페이스 패턴** 필수.

**`pkg/terminal/cmux.go`** (205 lines)

- `CmuxAdapter`: workspaceRef 저장, parseCmuxRef로 cmux CLI 출력 파싱
- 모든 메서드가 `execCommand("cmux", ...)` 패턴 사용 → 새 메서드도 동일 패턴으로 구현

### Debate 실행 루프

**`pkg/orchestra/interactive_debate.go`** (325 lines)

- `executeRound()` (L163-324): 가장 복잡한 함수
  - L167-183: Surface validation + recreatePane (Round 2+)
  - L187-194: Baseline 캡처 (surface validation 후)
  - L247-276: SendLongText 실패 → recreatePane → retry (exponential backoff)
  - L294-301: 프롬프트 전송 후 baseline 재캡처
  - L304-308: InitialDelay (기본 10s) 대기
  - L312-316: Hook mode vs screen polling 분기
- `runPaneDebate()` (L78-160): debate 루프 + 조기 합의 + judge round

## cmux API 테스트 결과 (v0.62.2)

### surface-health

```bash
$ cmux surface-health
surface:7 type=terminal in_window=true
```

- 출력 파싱: `surface:{N} type={type} in_window={bool}`
- ReadScreen보다 훨씬 경량 — 화면 내용 전송 없이 상태만 반환
- stale surface: 명령 실패 (exit code 1) 또는 `in_window=false`

### wait-for (수신)

```bash
$ cmux wait-for "done-claude"
# 블로킹 — "done-claude" 신호가 올 때까지 대기
# 신호 수신 시 즉시 리턴 (exit 0)
```

### wait-for -S (송신)

```bash
$ cmux wait-for -S "done-claude"
# "done-claude" 신호 전송 — 대기 중인 wait-for 즉시 해제
```

### set-hook (제한적)

```bash
$ cmux set-hook pane-died 'echo died'
# 등록은 성공하지만 pane exit 시 실제로 fire 되지 않음
# → 신뢰할 수 없어 사용하지 않기로 결정
```

### claude-hook

```bash
$ cmux claude-hook
# Available hooks: session-start, stop, session-end, notification, prompt-submit, pre-tool-use
# → prompt-submit 후크에서 완료 신호 전송 가능성 있으나, 응답 완료가 아닌 프롬프트 제출 시점
# → 사용하지 않고, 프로바이더 CLI 출력 완료 감지 후 wait-for -S로 신호 전송하는 별도 후크 스크립트 사용
```

## 설계 결정

### D1: 옵셔널 인터페이스 패턴 (Terminal 확장)

**결정**: Terminal 인터페이스에 새 메서드를 추가하지 않고, `SignalCapable` 인터페이스를 별도 정의하여 type assertion으로 기능 감지.

**이유**: `terminal.go`의 `@AX:ANCHOR` 마커가 이 인터페이스를 stable boundary로 표시. 메서드 추가 시 CmuxAdapter, TmuxAdapter, PlainAdapter 모든 구현체 동시 수정 필요. 옵셔널 인터페이스는 Go 표준 라이브러리에서도 `io.ReaderFrom`, `http.Flusher` 등으로 널리 사용되는 패턴.

**대안 검토**:
- A) Terminal 인터페이스에 직접 추가: AX:ANCHOR 위반, 3개 어댑터 동시 수정 필요
- B) Wrapper struct: 호출부 복잡도 증가, 모든 Terminal 전달 경로에서 wrapping 필요
- C) 옵셔널 인터페이스 (채택): 기존 코드 무변경, 점진적 채택 가능

### D2: cmux wait-for 기반 완료 탐지 (Primary)

**결정**: cmux의 `wait-for` / `wait-for -S` 신호 메커니즘을 primary 완료 탐지로 사용.

**이유**: 폴링 기반 ReadScreen은 (1) 2초 간격으로 지연 발생, (2) ANSI escape 오염으로 regex 매칭 불안정, (3) false positive/negative 발생. `wait-for`는 이벤트 기반으로 즉각 반응, 정확도 100%.

**대안 검토**:
- A) set-hook 기반: pane-died/pane-exited 후크 등록했으나 실제로 fire 되지 않음 → 불가
- B) claude-hook: prompt-submit은 응답 완료가 아닌 제출 시점 → 부적합
- C) pipe-pane output 감시 (fsnotify): 가능하나 cmux wait-for보다 복잡, P1 FileIPCDetector로 예약
- D) cmux wait-for (채택): 테스트에서 즉시 동작 확인, 가장 간단하고 정확

### D3: SurfaceManager 백그라운드 모니터링 (Proactive)

**결정**: 라운드 경계에서만 검사하는 reactive 방식 대신, 5초 간격 백그라운드 헬스 모니터링.

**이유**: 현재 `validateSurface()`는 Round 2 시작 시점에서만 호출. Surface가 Round 1 중간에 stale 되면 Round 2 시작 시 15-20초 복구 지연. 백그라운드 모니터링으로 사전 감지하면 복구를 미리 시작하거나 warm pool에서 즉시 교체 가능.

**대안 검토**:
- A) 현재 방식 유지 (reactive): 15-20초 복구 지연 지속
- B) ReadScreen 기반 헬스체크: 화면 내용 전체 전송으로 무겁고, 프로바이더 출력과 간섭
- C) surface-health 기반 (채택): 경량, 비파괴적, 5초 간격이면 최대 5초 지연

### D4: Warm Pool은 P1로 분류

**결정**: warm spare surface 사전 초기화를 P1으로 분류.

**이유**: P0의 surface-health 모니터링 + 기존 recreatePane()만으로도 현재보다 크게 개선. Warm pool은 추가 리소스(여유 surface + CLI 세션) 관리 복잡도가 높아 P0 범위를 초과. SurfaceManager 구조에 warm pool 확장 포인트만 P0에서 설계.

### D5: 프로바이더 후크 접근법

**결정**: 프로바이더 CLI의 내장 후크(claude-hook 등)를 사용하지 않고, 별도 완료 감지 + `cmux wait-for -S` 전송 후크 스크립트 사용.

**이유**: claude-hook의 prompt-submit은 응답 완료가 아닌 프롬프트 제출 시점. 응답 완료 이벤트는 프로바이더별로 다르고 일관된 후크 포인트 없음. 대신 기존 pipe-pane output의 idle 감지 또는 ScreenPollDetector가 완료를 감지한 후 `cmux wait-for -S`를 전송하는 하이브리드 접근.

## 리스크 및 완화

| 리스크 | 확률 | 영향 | 완화 |
|--------|------|------|------|
| cmux wait-for가 특정 시나리오에서 실패 | 낮음 | 높음 | ScreenPollDetector 자동 폴백 |
| surface-health 출력 포맷 변경 (cmux 업데이트) | 중간 | 중간 | 파싱 실패 시 ReadScreen 기반 validateSurface 폴백 |
| Warm pool surface가 사용 전에 stale | 낮음 | 낮음 | warm pool도 헬스체크 대상, stale 시 recreatePane 폴백 |
| 후크 스크립트가 프로바이더별로 호환 안됨 | 중간 | 중간 | 후크 없이도 ScreenPollDetector로 동작, 후크는 선택적 성능 향상 |
