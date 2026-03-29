# SPEC-ORCH-017 리서치

## 기존 코드 분석

### 결과 수집 프로토콜 (현재 동작 중 — 참조 구현)

**`pkg/orchestra/hook_signal.go`** — `HookSession` 구조체:
- `WaitForDone(timeout, providers...)` — 200ms polling으로 `{provider}-done` 파일 감시
- `WaitForDoneRound(timeout, provider, round)` — `RoundSignalName`으로 라운드별 done 파일 감시
- `WaitForDoneRoundCtx(ctx, timeout, provider, round)` — context 지원 버전
- `waitForFileCtx(ctx, timeout, filename)` — 내부 폴링 루프 (200ms 간격)
- `ReadResult(providers...)` / `ReadResultRound(provider, round)` — JSON 파싱
- `Dir()`, `SessionID()`, `HasHook(provider)`, `Cleanup()`

**`pkg/orchestra/round_signal.go`** — 라운드 시그널 유틸리티:
- `RoundSignalName(provider, round, suffix)` — `"{provider}-round{N}-{suffix}"` 생성
- `CleanRoundSignals(session, round)` — `*-round{N}-done` 패턴 삭제
- `SetRoundEnv(round)` / `SendRoundEnvToPane(ctx, term, paneID, round)`

**`pkg/orchestra/hook_watcher.go`** — `WaitAndCollectHookResults`:
- 프로바이더별 goroutine으로 병렬 결과 수집
- `collectSingleProvider` — HasHook 확인 → WaitForDone → ReadResult → fallback 체인

### 프롬프트 전달 (현재 문제점)

**`pkg/orchestra/interactive_debate.go`** — `executeRound`:
- Lines 231-268: Round 2+ 프롬프트 전달에 `SendLongText` 사용
- `SendLongText` 실패 시 `recreatePane` → 재시도 3회 → skipWait 설정
- `SendCommand(ctx, paneID, "\n")` — Enter 키 전달로 프롬프트 제출
- **문제**: paste-buffer 경유 → surface stale, PTY 4KB 한계, 경합 조건

**`pkg/orchestra/interactive_surface.go`** — `recreatePane`:
- Surface 유효성 검증 실패 시 pane 재생성
- 재생성 후 `SendLongText`로 CLI relaunch — 이것도 paste-buffer 문제에 노출

### Hook 스크립트 (현재 구현)

**`content/hooks/hook-claude-stop.sh`**:
- POSIX shell, python3 의존
- stdin JSON → `last_assistant_message` 추출 → result.json + done 시그널
- AUTOPUS_ROUND 환경변수로 라운드별 파일명 결정
- 50줄, 결과 출력 후 즉시 종료

**`content/hooks/hook-gemini-afteragent.sh`**:
- 구조 동일, `prompt_response` 필드 추출
- 50줄, 결과 출력 후 즉시 종료

**`content/hooks/hook-opencode-complete.ts`**:
- TypeScript, Node.js stdin 스트림 기반
- `text` 필드 추출 → result.json + done 시그널
- 42줄, 비동기 stdin 완료 후 종료

### 어댑터 (hook 배포)

**`pkg/adapter/claude/`**:
- `claude_settings.go` — Claude Code 설정 관리
- `claude.go` — hook 파일 배포 로직 포함
- `claude_hooks_test.go` — hook 배포 테스트

### 세션 디렉토리 패턴

```
/tmp/autopus/{session-id}/
  {provider}-round{N}-result.json   // 기존 output
  {provider}-round{N}-done          // 기존 완료 시그널
  {provider}-round{N}-input.json    // NEW: input 프롬프트
  {provider}-round{N}-ready         // NEW: CLI 수신 준비
```

## 설계 결정

### D1: Hook 확장 vs. 별도 Watcher 프로세스

**결정**: 기존 hook 스크립트를 확장하여 input reader 기능 추가

**이유**:
- 별도 watcher 프로세스는 관리 복잡성 증가 (프로세스 생명주기, 좀비 프로세스)
- Hook은 이미 AUTOPUS_SESSION_ID, AUTOPUS_ROUND 환경변수에 접근 가능
- Hook 실행 시점이 정확히 "결과 출력 직후 = 다음 입력 수신 준비"
- BS-011의 ICE C2 평가: "hook fire 시 input 파일 자동 읽기, 감시 불필요" — hook 실행 자체가 트리거

**대안 검토**:
- fsnotify/inotify 기반 독립 watcher → 프로세스 관리 부담, 크로스플랫폼 복잡성
- Named pipe (FIFO) → POSIX 호환이지만 blocking 특성으로 데드락 위험
- Unix domain socket → 양방향 통신 가능하나 hook 스크립트에서 사용 복잡

### D2: Atomic Write 전략

**결정**: tmp 파일 작성 → `os.Rename` (same filesystem 보장)

**이유**:
- `/tmp/autopus/` 내에서 tmp → rename이므로 같은 filesystem
- Hook의 polling이 불완전한 JSON을 읽는 것을 방지
- Go의 `os.Rename`은 POSIX `rename(2)` — atomic on same filesystem

### D3: Ready 시그널 필요성

**결정**: ready 시그널 도입 (WaitForReady)

**이유**:
- Hook이 결과를 출력하고 input 감시 루프에 진입하기까지 시간 차이 존재
- Orchestra가 ready 전에 input을 쓰면 hook이 감지하지 못할 수 있음
- Ready 시그널로 명시적 핸드셰이크 보장

**대안 검토**:
- Input 파일을 먼저 쓰고 hook이 나중에 발견 → 타이밍 의존적, 이전 라운드 잔여 파일과 혼동 가능
- Done 시그널을 ready로 재사용 → 의미 혼동, done은 "결과 완료"이지 "입력 준비"가 아님

### D4: Fallback 공존 설계

**결정**: 프로바이더별 독립 경로 선택 (file IPC or paste-buffer)

**이유**:
- 모든 프로바이더가 hook을 지원하지 않을 수 있음 (커스텀 프로바이더)
- Hook이 실패한 특정 프로바이더만 fallback하고 나머지는 file IPC 유지
- `HasHook(provider)` 체크가 이미 존재하여 분기 조건이 자연스러움

### D5: Hook 내 프롬프트 주입 방식

**결정**: 각 CLI별 최적 방식 연구 필요 (구현 단계에서 확정)

**Claude Code**: Stop hook 내에서 다음 프롬프트를 어떻게 주입하는지가 핵심 과제. 가능한 접근:
  - hook 스크립트가 input JSON을 읽고 `claude --prompt "..."` 형태로 새 명령 실행 (but 기존 세션 유지 불가)
  - hook이 input 파일 경로를 stdout으로 출력하여 Claude Code가 인식 (Claude Code의 hook output 처리 지원 여부 확인 필요)
  - 별도의 pre-tool hook 또는 notification hook이 있는지 확인

**Gemini CLI**: AfterAgent hook 이후 Gemini가 stdin에서 대기하는지 확인 필요. 대기한다면 hook에서 stdin에 쓰기 가능.

**opencode**: TS 플러그인이 비동기 이벤트로 input을 주입할 수 있는지 플러그인 API 확인 필요.

이 부분은 T4-T6 구현 시 각 CLI의 구체적인 입력 메커니즘을 조사하여 확정한다.

### D6: SendLongText 대체 범위

**결정**: Round 2+ 프롬프트 전달만 전환. Round 1과 recreatePane의 CLI launch는 기존 방식 유지.

**이유**:
- Round 1은 CLI args로 프롬프트를 전달하거나 최초 1회 SendLongText — 문제 빈도 낮음
- recreatePane의 launch 명령은 짧은 CLI 명령어 — paste-buffer 한계에 걸리지 않음
- 점진적 전환으로 위험 최소화
