# SPEC-ORCH-017 구현 계획

## 태스크 목록

- [ ] T1: `hook_input.go` — HookInput 구조체, atomic write 헬퍼 함수 구현
- [ ] T2: `hook_signal.go` — WriteInput/WriteInputRound, WaitForReady/WaitForReadyCtx 메서드 추가
- [ ] T3: `round_signal.go` — CleanRoundSignals 확장 (input.json, ready 파일 정리 패턴 추가)
- [ ] T4: `hook-claude-stop.sh` — 결과 출력 후 ready 시그널 생성 + input 파일 감시/제출 루프
- [ ] T5: `hook-gemini-afteragent.sh` — 결과 출력 후 ready 시그널 생성 + input 파일 감시/제출 루프
- [ ] T6: `hook-opencode-complete.ts` — 결과 출력 후 ready 시그널 생성 + input 파일 감시/제출
- [ ] T7: `interactive_debate.go` — executeRound Round 2+ 경로를 file IPC 우선으로 전환, SendLongText fallback 유지
- [ ] T8: 어댑터 배포 로직 업데이트 — 확장된 hook 파일이 auto init/update 시 정상 배포되는지 확인
- [ ] T9: 통합 테스트 — file IPC 경로 정상 동작 + fallback 경로 동작 검증

## 구현 전략

### Phase 1: Core Protocol (T1-T3)

Go 코드에 input/ready 시그널 프로토콜을 추가한다. 기존 `HookSession` 구조체를 확장하되, 기존 output 경로는 전혀 변경하지 않는다.

- `hook_input.go`에 `HookInput` 구조체와 `atomicWriteJSON` 유틸리티를 분리
- `hook_signal.go`에 `WriteInput*`, `WaitForReady*` 메서드를 추가 — 기존 `WaitForDone*`과 동일한 패턴
- `round_signal.go`의 `CleanRoundSignals`에 `*-round{N}-input.json`, `*-round{N}-ready` glob 패턴 추가

### Phase 2: Hook Scripts (T4-T6)

각 CLI hook 스크립트에 양방향 기능을 추가한다.

**핵심 패턴 (모든 hook 공통):**
1. 기존 결과 출력 로직 유지
2. done 시그널 작성 직후 ready 시그널 작성
3. input 파일 감시 루프 진입 (200ms polling, 120s timeout)
4. input 감지 시 JSON에서 prompt 추출 → CLI에 전달
5. input 전달 후 루프 종료 (다음 라운드는 새 hook 실행에서 처리)

**CLI별 프롬프트 주입 방식:**
- Claude Code: Stop hook이 결과 수집 + 다음 입력 준비까지 처리. stdin을 통해 다음 프롬프트를 CLI에 전달하는 것은 불가능하므로, input 파일 경로를 별도 watcher가 감시하거나 claude의 `--input-file` 등 CLI 기능을 활용
- Gemini: AfterAgent hook 종료 후 gemini가 다음 입력을 기다리는 상태에서 hook이 input을 감지하여 stdin에 쓰기
- opencode: TS 플러그인이 비동기 listener로 input 감시

### Phase 3: executeRound 전환 (T7)

`interactive_debate.go`의 `executeRound`에서 Round 2+ 프롬프트 전달 경로를 분기:

```
if hookSession != nil && hookSession.HasHook(provider) {
    // File IPC path
    WaitForReadyCtx → WriteInputRound → (hook이 CLI에 주입)
} else {
    // Legacy paste-buffer path (기존 코드 유지)
    SendLongText → SendCommand("\n")
}
```

기존 SendLongText 코드는 삭제하지 않고 else 분기로 보존하여 fallback으로 기능한다.

### Phase 4: Adapter & Test (T8-T9)

- `pkg/adapter/` 각 플랫폼의 hook 배포 로직이 새 hook 버전을 올바르게 복사하는지 확인
- 통합 테스트에서 file IPC 경로와 fallback 경로 모두 검증

### 의존 관계

```
T1 ──→ T2 ──→ T7
         ↘
T3       T4, T5, T6 (병렬) ──→ T8 ──→ T9
```

### 기존 코드 활용

- `HookSession` 구조체의 기존 패턴(WaitForDone, ReadResult)을 그대로 따름
- `RoundSignalName` 함수를 input/ready 시그널 이름 생성에도 재사용
- `waitForFileCtx` 내부 메서드를 WaitForReady에도 재사용
- Hook 스크립트의 기존 POSIX shell / TS 구조를 최대한 유지하고 input 감시 루프만 추가

### 변경 범위

| 영역 | 신규 | 수정 |
|------|------|------|
| Go (pkg/orchestra/) | 1 파일 | 3 파일 |
| Hook (content/hooks/) | 0 | 3 파일 |
| Adapter (pkg/adapter/) | 0 | 배포 로직 확인 |
| Test | 1+ 파일 | 기존 테스트 업데이트 |
