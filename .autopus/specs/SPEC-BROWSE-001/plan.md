# SPEC-BROWSE-001 구현 계획

## 태스크 목록

- [ ] T1: `pkg/browse/backend.go` — BrowserBackend 인터페이스 + SessionID 타입 + NewBackend 팩토리
  - BrowserBackend 인터페이스 (Open, Snapshot, Click, Fill, Screenshot, Close, Name)
  - NewBackend(term terminal.Terminal) BrowserBackend 팩토리 함수
  - cmux → CmuxBrowserBackend, 그 외 → AgentBrowserBackend

- [ ] T2: `pkg/browse/cmux.go` — CmuxBrowserBackend 구현
  - `cmux browser open <url>` 실행 → surface ref 파싱
  - `cmux browser --surface <ref> snapshot/click/fill/screenshot` 래핑
  - `cmux close-surface --surface <ref>` 정리
  - shell escape 적용

- [ ] T3: `pkg/browse/agent.go` — AgentBrowserBackend 구현
  - `agent-browser open/snapshot/click/fill/screenshot` 래핑
  - 기존 browser-automation.md 스킬의 명령 패턴 준수

- [ ] T4: 테스트 — backend_test.go, cmux_test.go, agent_test.go
  - mock exec 패턴으로 CLI 호출 검증
  - 팩토리 라우팅 테스트 (cmux → CmuxBrowserBackend)
  - fallback 테스트 (cmux 실패 → AgentBrowserBackend)

## 구현 전략

### 접근 방법

Strategy 패턴으로 BrowserBackend를 추상화한다. 각 백엔드는 해당 CLI 도구의 래퍼이며, `os/exec.Command`를 통해 실행한다.

### 파일 소유권

| Task | Files | Mode |
|------|-------|------|
| T1 | backend.go, backend_test.go | sequential (first) |
| T2 | cmux.go, cmux_test.go | parallel with T3 |
| T3 | agent.go, agent_test.go | parallel with T2 |
| T4 | 통합 테스트 추가 | sequential (last) |

### 변경 범위

- **신규**: `pkg/browse/` 패키지 (6파일, 각 ~80-120행)
- **수정 없음**: 기존 `pkg/terminal/`, `pkg/orchestra/` 변경 없음

### 예상 코드량

| File | Lines |
|------|-------|
| backend.go | ~50 |
| cmux.go | ~120 |
| agent.go | ~100 |
| backend_test.go | ~60 |
| cmux_test.go | ~120 |
| agent_test.go | ~100 |
| **합계** | **~550** |
