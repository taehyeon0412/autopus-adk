# SPEC-BROWSE-001 리서치

## cmux browser API 분석

### 브라우저 열기

```bash
cmux browser open https://example.com
# → 브라우저 pane이 cmux workspace에 임베딩
# → surface ref 반환 (예: surface:5)

cmux browser open-split https://example.com
# → 현재 pane을 분할하여 브라우저 열기
```

### 핵심 자동화 명령

```bash
# 스냅샷 (접근성 트리)
cmux browser --surface surface:5 snapshot
cmux browser --surface surface:5 snapshot --interactive  # 인터랙티브 모드
cmux browser --surface surface:5 snapshot --compact       # 간결 모드
cmux browser --surface surface:5 snapshot --selector "div.main"  # 특정 영역만

# 클릭/상호작용
cmux browser --surface surface:5 click "button.submit"
cmux browser --surface surface:5 click "button.submit" --snapshot-after  # 클릭 후 자동 스냅샷
cmux browser --surface surface:5 fill "input#email" "test@example.com"
cmux browser --surface surface:5 type "textarea" "hello"
cmux browser --surface surface:5 press Enter

# 네비게이션
cmux browser --surface surface:5 navigate https://other.com
cmux browser --surface surface:5 back
cmux browser --surface surface:5 reload

# 대기
cmux browser --surface surface:5 wait --selector "div.loaded"
cmux browser --surface surface:5 wait --text "Welcome"
cmux browser --surface surface:5 wait --load-state complete

# 상태 확인
cmux browser --surface surface:5 is visible "button.submit"
cmux browser --surface surface:5 is enabled "input#email"
cmux browser --surface surface:5 get text "h1"
cmux browser --surface surface:5 get title

# 스크린샷
cmux browser --surface surface:5 screenshot --out /tmp/page.png

# JavaScript 실행
cmux browser --surface surface:5 eval "document.title"

# 쿠키/스토리지
cmux browser --surface surface:5 cookies get
cmux browser --surface surface:5 storage local get --key "token"

# 의미론적 로케이터
cmux browser --surface surface:5 find role button --name "Save"
cmux browser --surface surface:5 find text "Sign In"
cmux browser --surface surface:5 find testid "submit-btn"
```

### 닫기

```bash
cmux close-surface --surface surface:5
```

## agent-browser API 비교

```bash
agent-browser open https://example.com
agent-browser snapshot                    # → @e1, @e2 참조 반환
agent-browser click @e3
agent-browser fill @e4 "text"
agent-browser screenshot /tmp/page.png
agent-browser is visible @e3
agent-browser find role button --name "Save"
```

## 핵심 차이점

| 항목 | cmux browser | agent-browser |
|------|-------------|---------------|
| **셀렉터** | CSS 셀렉터 (`button.submit`) | 접근성 트리 참조 (`@e3`) |
| **브라우저 위치** | cmux workspace 내 임베딩 | 별도 Chromium 프로세스 |
| **사용자 시각화** | cmux 안에서 직접 보임 | `--headed`일 때만 별도 창 |
| **surface 관리** | surface ref 기반 | 암묵적 세션 |
| **고급 기능** | eval, network.route, trace, screencast | 기본 자동화만 |
| **설치 요구** | cmux 필수 | npm install -g agent-browser |

## 설계 결정

### D1: BrowserBackend 인터페이스 추상화

**결정**: Strategy 패턴으로 BrowserBackend 인터페이스를 정의하고, cmux/agent-browser를 각각 구현체로 분리

**이유:**
- 호출자(스킬, 파이프라인)는 백엔드 세부사항을 몰라도 됨
- 터미널 환경에 따라 자동 라우팅
- 테스트에서 mock 백엔드 주입 가능

**대안:**
- 분기문으로 처리 — 확장성 부족, 코드 중복
- cmux 전용으로 구현 — tmux/plain 사용자 지원 불가

### D2: 셀렉터 통일 전략

**결정**: BrowserBackend 인터페이스는 `selector string`으로 통일. 각 백엔드가 내부적으로 해석.

**이유:**
- cmux는 CSS 셀렉터를 직접 사용
- agent-browser는 snapshot 후 @e1 참조를 사용
- 호출자가 환경별 셀렉터를 알 필요 없음
- snapshot → 요소 선택 → 조작의 워크플로우는 백엔드 내부에서 처리

### D3: tmux 브라우저 미지원

**결정**: tmux에는 네이티브 브라우저 기능이 없으므로 agent-browser로 fallback

**이유:**
- tmux는 터미널 멀티플렉서일 뿐, 브라우저 임베딩 불가
- agent-browser는 tmux pane 안에서 headless/headed로 실행 가능
- 불필요한 추상화 방지
