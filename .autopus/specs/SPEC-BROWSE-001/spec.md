# SPEC-BROWSE-001: 브라우저 자동화 터미널 어댑터 통합

**Status**: completed
**Created**: 2026-03-26
**Updated**: 2026-03-26
**Domain**: BROWSE

## 목적

`/auto browse`가 터미널 환경에 따라 최적의 브라우저 백엔드를 자동 선택하도록 한다. cmux 환경에서는 네이티브 `cmux browser` API를 사용하고, cmux가 없는 환경에서는 `agent-browser`로 fallback한다. 사용자와 AI 모두 동일한 `/auto browse` 인터페이스로 브라우저를 조작할 수 있다.

## 핵심 설계: Backend 라우팅

```
/auto browse <url>
  └─ DetectTerminal()
      ├─ cmux  → CmuxBrowserBackend (cmux browser CLI)
      ├─ tmux  → AgentBrowserBackend (agent-browser CLI)
      └─ plain → AgentBrowserBackend (agent-browser CLI)
```

cmux는 `cmux browser open/snapshot/click/fill` 등 완전한 브라우저 자동화 API를 네이티브로 제공한다. tmux와 plain에는 이런 기능이 없으므로 기존 `agent-browser`를 사용한다.

## 요구사항

### R1: BrowserBackend 인터페이스

THE SYSTEM SHALL define a `BrowserBackend` interface that abstracts browser automation across different terminal environments.

```go
type BrowserBackend interface {
    Open(ctx context.Context, url string) (SessionID, error)
    Snapshot(ctx context.Context) (string, error)
    Click(ctx context.Context, selector string) error
    Fill(ctx context.Context, selector string, text string) error
    Screenshot(ctx context.Context, outPath string) error
    Close(ctx context.Context) error
    Name() string
}
```

### R2: 터미널 감지 및 백엔드 자동 선택

WHEN the user invokes `/auto browse`
THE SYSTEM SHALL detect the current terminal via `terminal.DetectTerminal()` and select the appropriate backend:
- cmux → CmuxBrowserBackend
- tmux → AgentBrowserBackend
- plain → AgentBrowserBackend

### R3: CmuxBrowserBackend 구현

WHEN the detected terminal is cmux
THE SYSTEM SHALL use `cmux browser` CLI commands:
- `cmux browser open <url>` → 브라우저 pane 열기 (surface ref 반환)
- `cmux browser --surface <ref> snapshot` → 접근성 트리
- `cmux browser --surface <ref> click <selector>` → 요소 클릭
- `cmux browser --surface <ref> fill <selector> <text>` → 입력
- `cmux browser --surface <ref> screenshot --out <path>` → 스크린샷
- `cmux browser --surface <ref> close` 또는 `cmux close-surface --surface <ref>` → 정리

### R4: AgentBrowserBackend 구현

WHEN the detected terminal is tmux or plain
THE SYSTEM SHALL use `agent-browser` CLI commands:
- `agent-browser open <url>` → 브라우저 열기
- `agent-browser snapshot` → 접근성 트리
- `agent-browser click <ref>` → 요소 클릭
- `agent-browser fill <ref> <text>` → 입력
- `agent-browser screenshot <path>` → 스크린샷

### R5: cmux 브라우저 pane 임베딩

WHILE CmuxBrowserBackend is active
THE SYSTEM SHALL embed the browser directly in the cmux workspace, allowing the user to visually see the page while the AI agent interacts with it.

### R6: 에러 시 fallback

WHEN CmuxBrowserBackend fails (cmux browser open error)
THE SYSTEM SHALL fall back to AgentBrowserBackend and log a warning.

### R7: 세션 정리

WHEN the browse session ends
THE SYSTEM SHALL close the browser surface (cmux) or process (agent-browser).

## 생성 파일 상세

| 파일 | 역할 |
|------|------|
| `pkg/browse/backend.go` | BrowserBackend 인터페이스 정의 + NewBackend(term) 팩토리 |
| `pkg/browse/backend_test.go` | 팩토리 테스트 |
| `pkg/browse/cmux.go` | CmuxBrowserBackend — cmux browser CLI 래핑 |
| `pkg/browse/cmux_test.go` | CmuxBrowserBackend 유닛 테스트 |
| `pkg/browse/agent.go` | AgentBrowserBackend — agent-browser CLI 래핑 |
| `pkg/browse/agent_test.go` | AgentBrowserBackend 유닛 테스트 |

## 기존 코드 재사용

- `pkg/terminal.Terminal` 인터페이스 — `Name()` 메서드로 터미널 종류 감지
- `pkg/terminal.DetectTerminal()` — cmux > tmux > plain 우선순위
- `pkg/pipeline/team_pane.go:teamShellEscape` — shell escape 패턴

## cmux browser vs agent-browser 명령 매핑

| 동작 | cmux browser | agent-browser |
|------|-------------|---------------|
| 열기 | `cmux browser open <url>` | `agent-browser open <url>` |
| 스냅샷 | `cmux browser snapshot` | `agent-browser snapshot` |
| 클릭 | `cmux browser click <css>` | `agent-browser click <ref>` |
| 입력 | `cmux browser fill <css> <text>` | `agent-browser fill <ref> <text>` |
| 스크린샷 | `cmux browser screenshot --out <path>` | `agent-browser screenshot <path>` |
| 네비게이션 | `cmux browser navigate <url>` | `agent-browser open <url>` |
| 대기 | `cmux browser wait --selector <css>` | `agent-browser wait --text <text>` |
| 상태 확인 | `cmux browser is visible <css>` | `agent-browser is visible <ref>` |
| 닫기 | `cmux close-surface --surface <ref>` | (프로세스 종료) |

**셀렉터 차이**: cmux는 CSS 셀렉터, agent-browser는 `@e1` 접근성 트리 참조. BrowserBackend 인터페이스는 `selector string`으로 통일하고, 각 백엔드가 내부적으로 변환한다.
