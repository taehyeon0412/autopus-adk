# SPEC-ORCH-012 수락 기준

## 시나리오

### S1: cmux SendLongText — 긴 텍스트 buffer 경로
- Given: cmux 어댑터와 2000바이트 이상의 프롬프트 텍스트
- When: `SendLongText(ctx, "surface:7", longText)`를 호출하면
- Then: `cmux set-buffer --name <unique> <text>` → `cmux paste-buffer --name <unique> --surface surface:7` → `cmux delete-buffer --name <unique>` 순서로 실행된다

### S2: cmux SendLongText — 짧은 텍스트 기존 경로
- Given: cmux 어댑터와 100바이트 텍스트
- When: `SendLongText(ctx, "surface:7", shortText)`를 호출하면
- Then: `cmux send --surface surface:7 <text>` (기존 SendCommand 경로)로 실행된다

### S3: 병렬 버퍼 이름 충돌 방지
- Given: 3개 프로바이더가 동시에 SendLongText를 호출하는 상황
- When: 각각 다른 paneID(surface:1, surface:2, surface:3)로 호출하면
- Then: 생성된 버퍼 이름이 모두 다르다 (paneID + timestamp 기반 unique naming)

### S4: set-buffer 실패 시 fallback
- Given: `cmux set-buffer`가 에러를 반환하는 상황
- When: `SendLongText`가 500B 이상 텍스트로 호출되면
- Then: warning 로그를 출력하고 `SendCommand` (cmux send) 경로로 fallback하여 텍스트를 전달한다

### S5: delete-buffer 실패 시 best-effort
- Given: `cmux delete-buffer`가 에러를 반환하는 상황
- When: set-buffer → paste-buffer가 성공한 후 cleanup 단계에서 실패하면
- Then: 에러를 무시하고 nil을 반환한다 (텍스트 전달 자체는 성공)

### S6: launch command를 SendLongText로 전달
- Given: interactive 모드에서 opencode(args 프로바이더)의 launch command가 500B 이상
- When: `launchInteractiveSessions`가 실행되면
- Then: launch command 본문은 `SendLongText`로, Enter는 별도 `SendCommand("\n")`로 전달된다

### S7: launch command 짧은 경우에도 SendLongText 경로
- Given: gemini의 launch command가 짧은 경우 (100B 미만)
- When: `launchInteractiveSessions`가 실행되면
- Then: `SendLongText`가 호출되고, 내부적으로 `SendCommand`에 위임된다 (동작 동일)

### S8: tmux 환경 회귀 없음
- Given: tmux 어댑터를 사용하는 환경
- When: 기존 orchestra 플로우를 실행하면
- Then: tmux `SendLongText`의 load-buffer/paste-buffer 경로가 변경 없이 동작한다

### S9: 한글/특수문자 포함 텍스트 전달
- Given: 한글, 개행(`\n`), 인용부호(`"`, `'`) 포함 2000B 텍스트
- When: cmux `SendLongText`로 전달하면
- Then: `exec.Command`가 shell bypass하므로 특수문자가 손상 없이 `set-buffer`에 전달된다
