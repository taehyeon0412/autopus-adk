# SPEC-ORCH-007 수락 기준

## P0 시나리오

### S1: Hook 파일 시그널로 결과 수집 (R1, R5)

- Given: hook이 자동 주입된 상태에서 3개 프로바이더(claude, gemini, opencode)가 설정됨
- When: `auto orchestra --multi "테스트 프롬프트"` 실행
- Then: 각 프로바이더 완료 후 `/tmp/autopus/{session-id}/{provider}-result.json`에 구조화된 결과가 저장됨
- And: `/tmp/autopus/{session-id}/{provider}-done` 시그널 파일이 생성됨
- And: 오케스트레이터가 done 파일을 감지하여 result.json을 파싱하고 `ProviderResponse.Output`에 깨끗한 응답 텍스트가 저장됨

### S2: Claude Code Stop Hook 동작 (R2)

- Given: `.claude/settings.json`에 Stop hook이 등록된 상태
- When: Claude Code CLI가 응답을 완료하고 Stop hook이 실행됨
- Then: hook 스크립트가 stdin JSON에서 `last_assistant_message`를 추출
- And: `/tmp/autopus/{session-id}/claude-result.json`에 `{ session_id, provider: "claude", response: "...", timestamp: "..." }` 형식으로 저장
- And: `/tmp/autopus/{session-id}/claude-done` 빈 파일이 생성됨
- And: result.json 파일 권한이 0o600

### S3: Gemini CLI AfterAgent Hook 동작 (R3)

- Given: `.gemini/settings.json`에 AfterAgent hook이 등록된 상태
- When: Gemini CLI가 에이전트 응답을 완료하고 AfterAgent hook이 실행됨
- Then: hook 스크립트가 stdin JSON에서 `prompt_response`를 추출
- And: `/tmp/autopus/{session-id}/gemini-result.json`에 결과가 저장됨
- And: `/tmp/autopus/{session-id}/gemini-done` 시그널 파일이 생성됨

### S4: opencode Plugin 동작 (R4)

- Given: `opencode.json`에 `experimental.text.complete` plugin이 등록된 상태
- When: opencode가 텍스트 완성을 수행하고 plugin이 실행됨
- Then: plugin이 `text` 필드를 추출
- And: `/tmp/autopus/{session-id}/opencode-result.json`에 결과가 저장됨
- And: `/tmp/autopus/{session-id}/opencode-done` 시그널 파일이 생성됨

### S5: 파일 감시로 완료 감지 (R5)

- Given: hook 모드가 활성화된 프로바이더가 있는 상태
- When: `waitForCompletion()`이 실행됨
- Then: ReadScreen 폴링 대신 `os.Stat()` 200ms 폴링으로 done 파일을 확인
- And: done 파일 감지 후 500ms 이내에 result.json 파싱 완료
- And: result.json 파싱 시간이 10ms 미만

### S6: Hook 자동 주입 — Claude (R6)

- Given: Claude Code CLI가 설치된 상태에서 `.claude/settings.json`에 사용자 커스텀 hook이 존재
- When: `auto init` 실행
- Then: `.claude/settings.json`의 `hooks.Stop`에 autopus 결과 수집 스크립트가 추가됨
- And: 기존 사용자 커스텀 hook이 보존됨 (merge 방식)

### S7: Hook 자동 주입 — Gemini (R6)

- Given: Gemini CLI가 설치된 상태
- When: `auto init` 실행
- Then: `.gemini/settings.json`의 `hooks.AfterAgent`에 결과 수집 스크립트가 등록됨

### S8: Hook 자동 주입 — opencode (R6)

- Given: opencode가 설치된 상태
- When: `auto init` 실행
- Then: `opencode.json`의 `experimental.text.complete` plugin에 autopus 결과 수집 plugin이 등록됨

### S9: ANSI 파싱 코드 격리 (R7)

- Given: hook 모드가 활성화된 프로바이더의 결과 수집 경로
- When: hook 결과를 수집할 때
- Then: `cleanScreenOutput()`, `stripANSI()`, `filterPromptLines()`가 호출되지 않음
- And: JSON의 `response` 필드가 그대로 `ProviderResponse.Output`에 저장됨

### S10: Graceful Degradation — 혼합 모드 (R8)

- Given: claude hook은 설정되었으나 gemini hook은 미설정인 상태
- When: `auto orchestra --multi` 실행
- Then: claude는 hook 기반으로 done 파일 감시 → result.json 파싱으로 결과 수집
- And: gemini는 기존 ReadScreen + idle 감지 fallback으로 결과 수집
- And: 두 프로바이더의 결과가 모두 `mergeByStrategy()`에 정상 전달됨

### S11: Graceful Degradation — 전체 fallback (R8)

- Given: 어떤 프로바이더도 hook이 설정되지 않은 상태
- When: `auto orchestra --multi` 실행
- Then: 전체가 SPEC-ORCH-006의 ReadScreen 모드로 fallback
- And: 기능 회귀 없이 기존과 동일하게 동작

### S12: Hook 타임아웃 fallback (R8)

- Given: hook이 설정되었으나 프로바이더가 응답하지 않는 상태
- When: done 파일이 타임아웃 내에 생성되지 않음
- Then: 해당 프로바이더는 ReadScreen fallback으로 전환
- And: `ProviderResponse.TimedOut`이 true로 설정됨
- And: 다른 프로바이더의 hook 결과 수집에 영향 없음

## P1 시나리오

### S13: Codex → opencode 마이그레이션 (R9)

- Given: `autopus.yaml`에 `codex` 프로바이더가 설정된 상태
- When: `MigrateOrchestraConfig()` 실행
- Then: `codex` 엔트리가 `opencode`로 자동 변환됨
- And: 경고 메시지가 출력됨
- And: 바이너리 경로가 `opencode`로 업데이트됨

### S14: Debate 전략 Hook 연동 (R11)

- Given: 3개 프로바이더의 hook 결과가 모두 수집된 상태에서 debate 전략 실행
- When: Phase 2 rebuttal 라운드 실행
- Then: 각 프로바이더에 다른 프로바이더의 hook 결과(`response` 필드)가 rebuttal 프롬프트에 포함됨
- And: ANSI 코드 없이 깨끗한 텍스트가 주입됨

### S15: Relay 전략 Hook 연동 (R12)

- Given: relay 전략이 hook 모드에서 실행 중
- When: Provider A 완료 후 Provider B 프롬프트 생성
- Then: Provider A의 hook 결과(`response` 필드)가 Provider B 프롬프트에 포함됨

### S16: Consensus 전략 Hook 연동 (R13)

- Given: 3개 프로바이더의 hook 결과가 수집된 상태
- When: consensus 전략의 `MergeConsensus()` 실행
- Then: hook 결과 JSON의 `response` 필드가 `ProviderResponse.Output`으로 사용됨
- And: 66% 합의 판정이 깨끗한 텍스트 기반으로 수행됨

## 비기능 시나리오

### S17: 세션 디렉토리 정리

- Given: 오케스트레이션이 정상 완료된 상태
- When: cleanup 로직 실행
- Then: `/tmp/autopus/{session-id}/` 디렉토리와 모든 하위 파일이 삭제됨

### S18: 세션 디렉토리 보안

- Given: 세션 디렉토리 생성
- When: `/tmp/autopus/{session-id}/` 디렉토리 생성
- Then: 디렉토리 권한이 0o700 (소유자만 접근)
- And: result.json 파일 권한이 0o600

### S19: Hook 실패 격리

- Given: Claude Code Stop hook 스크립트가 에러로 실패
- When: hook 스크립트가 비정상 종료
- Then: Claude Code CLI 동작에 영향 없음 (fire-and-forget)
- And: done 파일이 미생성되어 타임아웃 후 ReadScreen fallback 발동

### S20: 파일 감지 성능

- Given: hook이 done 파일을 생성
- When: 오케스트레이터의 os.Stat 폴링이 실행 중
- Then: done 파일 생성 후 500ms 이내에 감지
- And: result.json 파싱 시간 < 10ms
