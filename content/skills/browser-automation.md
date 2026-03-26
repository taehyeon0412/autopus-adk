---
name: browser-automation
description: agent-browser CLI 기반 브라우저 자동화 스킬 — AI 에이전트가 직접 웹 페이지를 조작하고 검증
triggers:
  - browser
  - browse
  - 브라우저
  - 웹 테스트
  - 운영 확인
  - UI 확인
  - 페이지 확인
category: testing
level1_metadata: "agent-browser CLI, 접근성 트리 snapshot, 운영환경 UI 검증"
---

# Browser Automation Skill

agent-browser CLI를 활용하여 AI 에이전트가 직접 웹 페이지를 조작하고 검증하는 스킬입니다.

## 도구 선택

| 도구 | 용도 | 선택 기준 |
|------|------|-----------|
| **agent-browser** | 에이전트 직접 조작 | 운영환경 확인, 수동 탐색, 빠른 검증 (기본) |
| **Playwright** | E2E 테스트 스위트 | 반복 실행, CI/CD, 회귀 테스트 |

기본 도구는 **agent-browser**. Playwright는 테스트 코드 작성 시에만 사용.

## 사전 조건

```bash
# 설치 확인
agent-browser --version

# 미설치 시
npm install -g agent-browser
agent-browser install
```

## 핵심 워크플로우: Snapshot-Act-Verify

AI 에이전트가 브라우저를 조작하는 3단계 루프:

### Step 1: Open — 페이지 열기

```bash
agent-browser open <url>
```

### Step 2: Snapshot — 접근성 트리 + 참조 획득

```bash
agent-browser snapshot
```

snapshot은 페이지의 접근성 트리를 반환하며, 각 요소에 `@e1`, `@e2` 등의 참조를 할당한다. AI 에이전트는 이 참조를 사용하여 요소를 조작한다.

**snapshot 출력 예시:**
```
- @e1 heading "AI Settings"
- @e2 button "Provider Mode"
- @e3 switch "Auto Fallback" [checked]
- @e4 checkbox "Anthropic" [checked]
- @e5 checkbox "OpenAI" [checked]
- @e6 checkbox "Google" [unchecked]
- @e7 button "Save"
```

### Step 3: Act — 요소 상호작용

```bash
agent-browser click @e3        # 토글 클릭
agent-browser fill @e4 "text"  # 입력
agent-browser press Enter      # 키 입력
```

### Step 4: Verify — 상태 확인 + 스크린샷

```bash
agent-browser snapshot         # 변경 후 상태 재확인
agent-browser screenshot /tmp/verify.png  # 시각적 증거
agent-browser is visible @e3   # 요소 표시 여부
agent-browser is checked @e3   # 체크박스/토글 상태
agent-browser get text @e1     # 텍스트 내용 확인
```

## 주요 명령어 레퍼런스

### 네비게이션

```bash
agent-browser open <url>           # 페이지 이동
agent-browser get url              # 현재 URL
agent-browser get title            # 페이지 제목
agent-browser wait --load networkidle  # 로드 완료 대기
agent-browser wait --text "Welcome"    # 텍스트 출현 대기
```

### 상호작용

```bash
agent-browser click <ref>          # 클릭
agent-browser fill <ref> <text>    # 입력 필드 채우기
agent-browser type <ref> <text>    # 타이핑
agent-browser hover <ref>         # 마우스 오버
agent-browser scroll down 500     # 스크롤
agent-browser press Enter         # 키 입력
```

### 의미론적 로케이터 (snapshot 없이 직접 찾기)

```bash
agent-browser find role button click --name "Save"
agent-browser find text "Sign In" click
agent-browser find label "Email" fill "test@test.com"
```

### 상태 확인

```bash
agent-browser is visible <ref>    # 표시 여부
agent-browser is enabled <ref>    # 활성화 상태
agent-browser is checked <ref>    # 체크 상태
agent-browser get text <ref>      # 텍스트
agent-browser get html <ref>      # HTML
```

### 뷰포트 & 디바이스

```bash
agent-browser set viewport 1280 800
agent-browser set device "iPhone 14"
agent-browser set media dark      # 다크 모드
```

### 쿠키 & 인증

```bash
agent-browser cookies                      # 쿠키 목록
agent-browser cookies set <name> <value>   # 쿠키 설정
agent-browser storage local                # localStorage
```

### 배치 실행

```bash
echo '[["open","https://example.com"],["snapshot"],["click","@e1"]]' \
  | agent-browser batch --json
```

## 실행 모드

| 모드 | 플래그 | 용도 |
|------|--------|------|
| **Headless** (기본) | (없음) | CI/CD, 백그라운드 검증, 자동화 |
| **Headed** | `--headed` | 시각적 확인, 디버깅, 데모 |

```bash
agent-browser open https://autopus.co --headed    # 브라우저 창 표시
agent-browser open https://autopus.co              # 헤드리스 (기본)
```

## 운영환경 검증 패턴

### 패턴 1: UI 컴포넌트 존재 확인

```bash
agent-browser open https://example.com/settings/ai
agent-browser snapshot
# → @e3 switch "Auto Fallback" 이 존재하면 렌더링 성공
agent-browser screenshot /tmp/ai-settings.png
```

### 패턴 2: 토글 동작 검증

```bash
agent-browser snapshot                  # 초기 상태 확인
agent-browser click @e3                 # 토글 클릭
agent-browser wait 1000                 # 상태 변경 대기
agent-browser snapshot                  # 변경 후 상태 확인
```

### 패턴 3: 인증 후 테스트

```bash
agent-browser open https://example.com/login
agent-browser snapshot
agent-browser fill @e2 "user@example.com"
agent-browser fill @e3 "password"
agent-browser click @e4
agent-browser wait --load networkidle
agent-browser open https://example.com/settings/ai
agent-browser snapshot
```

## 판정 기준

| 판정 | 기준 |
|------|------|
| PASS | 기대 요소가 snapshot에 존재하고 올바른 상태 |
| WARN | 요소는 존재하나 상태가 예상과 다름 |
| FAIL | 기대 요소가 snapshot에 없거나 에러 발생 |

## 주의사항

- 기본 모드는 **헤드리스** — `--headed` 추가 시 브라우저 창이 화면에 표시됨
- 각 명령은 이전 세션을 유지 — 쿠키/로그인 상태가 보존됨
- `snapshot`은 **접근성 트리**만 반환 — CSS 스타일은 포함 안 됨 (시각 확인은 `screenshot` 사용)
- 운영환경 테스트 시 **쓰기 작업**(삭제, 설정 변경)은 신중하게 — 되돌릴 수 없을 수 있음
