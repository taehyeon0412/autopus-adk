# SPEC-BROWSE-001 수락 기준

## 시나리오

### S1: cmux 환경에서 브라우저 열기

- Given: cmux가 설치되어 있고 DetectTerminal()이 CmuxAdapter를 반환
- When: NewBackend(term)을 호출
- Then: CmuxBrowserBackend가 반환됨
- And: Open("https://example.com") 호출 시 `cmux browser open https://example.com`이 실행됨
- And: surface ref가 SessionID로 반환됨

### S2: cmux 환경에서 스냅샷

- Given: CmuxBrowserBackend가 Open으로 surface ref를 획득한 상태
- When: Snapshot()을 호출
- Then: `cmux browser --surface <ref> snapshot`이 실행됨
- And: 접근성 트리 문자열이 반환됨

### S3: cmux 환경에서 요소 클릭

- Given: CmuxBrowserBackend가 활성 상태
- When: Click("button.submit")을 호출
- Then: `cmux browser --surface <ref> click button.submit`이 실행됨

### S4: cmux 환경에서 입력

- Given: CmuxBrowserBackend가 활성 상태
- When: Fill("input#email", "test@example.com")을 호출
- Then: `cmux browser --surface <ref> fill input#email test@example.com`이 실행됨

### S5: cmux 환경에서 스크린샷

- Given: CmuxBrowserBackend가 활성 상태
- When: Screenshot("/tmp/test.png")을 호출
- Then: `cmux browser --surface <ref> screenshot --out /tmp/test.png`이 실행됨

### S6: tmux/plain 환경에서 fallback

- Given: DetectTerminal()이 TmuxAdapter 또는 PlainAdapter를 반환
- When: NewBackend(term)을 호출
- Then: AgentBrowserBackend가 반환됨
- And: Open() 호출 시 `agent-browser open`이 실행됨

### S7: cmux 실패 시 fallback

- Given: cmux가 감지되었으나 `cmux browser open`이 에러를 반환
- When: Open()을 호출
- Then: 경고 로그가 출력됨
- And: AgentBrowserBackend로 fallback

### S8: 세션 정리

- Given: CmuxBrowserBackend로 브라우저가 열린 상태
- When: Close()를 호출
- Then: `cmux close-surface --surface <ref>`가 실행됨
- And: surface가 정리됨

### S9: 셀렉터 전달

- Given: CmuxBrowserBackend가 활성 상태
- When: Click("button.submit")을 호출
- Then: CSS 셀렉터가 그대로 `cmux browser click`에 전달됨 (cmux는 CSS 셀렉터 사용)

### S10: agent-browser 셀렉터 전달

- Given: AgentBrowserBackend가 활성 상태
- When: Click("@e3")을 호출
- Then: 접근성 트리 참조가 그대로 `agent-browser click`에 전달됨
