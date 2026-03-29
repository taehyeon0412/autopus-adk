# SPEC-ORCH-017 수락 기준

## 시나리오

### S1: Input 파일 원자적 쓰기

- Given: HookSession이 세션 디렉토리 `/tmp/autopus/{session-id}/`로 초기화됨
- When: `WriteInputRound("claude", 2, "rebuttal prompt text")`를 호출
- Then: `{session-dir}/claude-round2-input.json`이 생성되고, JSON 내용이 `{"prompt":"rebuttal prompt text","round":2}`이며, 파일 퍼미션이 0o600

### S2: Atomic write 보장

- Given: WriteInputRound 실행 중
- When: 파일 쓰기가 진행되는 동안 다른 프로세스가 해당 경로를 읽으려 시도
- Then: 불완전한 JSON을 읽지 않음 (tmp 파일 → rename 패턴)

### S3: Ready 시그널 생성 (Claude hook)

- Given: AUTOPUS_SESSION_ID와 AUTOPUS_ROUND 환경변수가 설정됨
- When: hook-claude-stop.sh가 결과를 성공적으로 출력한 후
- Then: `{session-dir}/claude-round{N}-ready` 파일이 생성됨

### S4: Ready 시그널 대기

- Given: HookSession이 초기화되고 WaitForReadyCtx 호출
- When: 5초 후 `claude-round2-ready` 파일이 생성됨
- Then: WaitForReadyCtx가 에러 없이 반환됨

### S5: Ready 타임아웃 → Fallback

- Given: HookSession이 초기화되고 WaitForReadyCtx 호출 (10초 타임아웃)
- When: ready 파일이 타임아웃 내에 생성되지 않음
- Then: 에러 반환, 호출측에서 SendLongText fallback 경로로 진입, 경고 로그 출력

### S6: Hook Input 감시 및 프롬프트 전달 (Claude)

- Given: hook-claude-stop.sh가 ready 시그널을 작성하고 input 감시 루프 진입
- When: Orchestra가 `claude-round2-input.json`을 세션 디렉토리에 작성
- Then: hook이 JSON에서 prompt를 추출하고 Claude CLI에 프롬프트를 전달함

### S7: Hook Input 감시 타임아웃

- Given: hook이 input 감시 루프에 진입
- When: 120초 내에 input 파일이 생성되지 않음
- Then: hook이 정상 종료 (exit 0), 에러 발생하지 않음

### S8: executeRound File IPC 경로

- Given: hookSession이 활성화되고 HasHook("claude") == true
- When: executeRound가 round=2 프롬프트를 전송
- Then: SendLongText가 호출되지 않고, WaitForReady → WriteInputRound 순서로 실행

### S9: executeRound Fallback 경로

- Given: hookSession이 nil이거나 HasHook("custom-provider") == false
- When: executeRound가 round=2 프롬프트를 전송
- Then: 기존 SendLongText + SendCommand("\n") 경로가 실행됨

### S10: Round Signal 정리

- Given: 세션 디렉토리에 round 1의 done, input.json, ready 파일이 모두 존재
- When: `CleanRoundSignals(session, 1)` 호출
- Then: `*-round1-done`, `*-round1-input.json`, `*-round1-ready` 파일이 모두 삭제됨

### S11: 양방향 통신 전체 흐름 (E2E)

- Given: 3개 프로바이더(claude, gemini, opencode)로 2라운드 토론 실행
- When: Round 1 완료 후 Round 2 진입
- Then: 각 프로바이더에 대해 ready → input → result → done 순서로 파일 시그널이 생성/소비되고, paste-buffer가 사용되지 않음

### S12: Hook 배포 갱신

- Given: 확장된 hook 파일이 content/hooks/에 존재
- When: `auto init` 실행
- Then: 각 CLI 설정 디렉토리에 input reader 기능이 포함된 새 hook이 배포됨

### S13: 혼합 모드 — 일부 프로바이더만 File IPC

- Given: claude, gemini는 hook 지원, custom-llm은 hook 미지원
- When: 3개 프로바이더로 토론 실행
- Then: claude, gemini는 file IPC 경로, custom-llm은 SendLongText fallback 경로로 각각 독립 동작
