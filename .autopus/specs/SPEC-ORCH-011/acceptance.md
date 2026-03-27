# SPEC-ORCH-011 수락 기준

## 시나리오

### S1: opencode 프로바이더 interactive pane 프롬프트 전달

- **Given**: opencode 프로바이더가 interactive pane 모드로 구성되어 있고, `PromptViaArgs=true`이며, pane이 정상 생성된 상태
- **When**: `sendPrompts()`가 opencode pane에 사용자 프롬프트를 전달할 때
- **Then**: opencode가 프롬프트를 수신하여 실질적 AI 응답(기본 인사가 아닌)을 반환해야 한다

### S2: gemini 프로바이더 긴 프롬프트 전달 (tmux)

- **Given**: gemini 프로바이더가 tmux 터미널의 interactive pane에서 실행 중이고, 프롬프트 길이가 2000자 이상
- **When**: `sendPrompts()`가 해당 프롬프트를 gemini pane에 전달할 때
- **Then**: 프롬프트가 truncation 없이 전체 전달되어, gemini가 "메시지가 잘렸다"가 아닌 프롬프트에 대한 실질적 응답을 반환해야 한다

### S3: gemini 프로바이더 긴 프롬프트 전달 (cmux)

- **Given**: gemini 프로바이더가 cmux 터미널의 interactive pane에서 실행 중이고, 프롬프트 길이가 2000자 이상
- **When**: `sendPrompts()`가 해당 프롬프트를 gemini pane에 전달할 때
- **Then**: 프롬프트가 truncation 없이 전체 전달되어야 한다

### S4: 짧은 프롬프트 하위 호환성

- **Given**: claude 프로바이더가 interactive pane 모드에서 실행 중이고, 프롬프트 길이가 500바이트 미만
- **When**: `sendPrompts()`가 프롬프트를 전달할 때
- **Then**: 기존 `SendCommand` 방식과 동일하게 정상 전달되어야 한다

### S5: debate 라운드에서 긴 rebuttal 프롬프트 전달

- **Given**: debate 전략으로 3라운드 실행 중이고, round 2의 rebuttal 프롬프트가 topicIsolationInstruction + 원본 프롬프트 + 이전 라운드 응답으로 구성되어 3000자 이상
- **When**: `executeRound()`가 각 pane에 rebuttal 프롬프트를 전달할 때
- **Then**: 모든 프로바이더가 truncation 없이 전체 rebuttal 프롬프트를 수신해야 한다

### S6: 프롬프트 전달 실패 시 graceful 처리

- **Given**: 특정 프로바이더의 pane에 프롬프트 전달이 실패(터미널 에러)
- **When**: `sendPrompts()`가 에러를 감지할 때
- **Then**: 해당 프로바이더를 `FailedProvider`에 추가하고 `skipWait=true`로 마킹하며, 나머지 프로바이더는 계속 실행되어야 한다

### S7: TmuxAdapter.SendLongText load-buffer/paste-buffer 동작

- **Given**: TmuxAdapter가 초기화된 상태이고, paneID가 유효
- **When**: 2000바이트 이상의 텍스트로 `SendLongText`를 호출할 때
- **Then**: 임시 파일 생성 → `tmux load-buffer` → `tmux paste-buffer` 순으로 실행되고, 임시 파일이 cleanup되어야 한다

### S8: SendLongText 짧은 텍스트 fallback

- **Given**: 임의의 Terminal 어댑터가 초기화된 상태
- **When**: 500바이트 미만의 텍스트로 `SendLongText`를 호출할 때
- **Then**: 기존 `SendCommand`로 위임되어 동일하게 동작해야 한다
