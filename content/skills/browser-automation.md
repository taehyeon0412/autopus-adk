---
name: browser-automation
description: 브라우저 자동화 및 E2E 테스트 스킬
triggers:
  - browser
  - automation
  - e2e
  - playwright
  - 브라우저 자동화
  - e2e 테스트
category: testing
level1_metadata: "Playwright/Puppeteer E2E 테스트, 브라우저 자동화"
---

# Browser Automation Skill

Playwright를 활용한 브라우저 자동화 및 E2E 테스트 스킬입니다.

## 기본 패턴

### Playwright 기본 구조
```typescript
import { test, expect } from '@playwright/test';

test('사용자 로그인 플로우', async ({ page }) => {
  // 1. 페이지 이동
  await page.goto('/login');

  // 2. 요소 상호작용
  await page.fill('[name=email]', 'user@example.com');
  await page.fill('[name=password]', 'password');
  await page.click('[type=submit]');

  // 3. 결과 검증
  await expect(page).toHaveURL('/dashboard');
  await expect(page.locator('[data-testid=welcome]')).toBeVisible();
});
```

## 핵심 원칙

### 사용자 관점 테스트
```typescript
// ❌ 구현 세부사항 테스트
expect(component.state.isLoggedIn).toBe(true);

// ✅ 사용자 관점 테스트
await expect(page.locator('text=환영합니다')).toBeVisible();
```

### 안정적인 셀렉터
```typescript
// 우선순위 순서 (높음 → 낮음)
page.getByRole('button', { name: '로그인' })    // 최선
page.getByTestId('login-button')               // 좋음
page.locator('[data-testid=login-button]')     // 좋음
page.locator('.login-btn')                     // 지양 (CSS 의존)
page.locator('//button[1]')                    // 금지 (XPath 취약)
```

## E2E 테스트 구조

### Page Object Model
```typescript
class LoginPage {
  constructor(private page: Page) {}

  async login(email: string, password: string) {
    await this.page.fill('[name=email]', email);
    await this.page.fill('[name=password]', password);
    await this.page.click('[type=submit]');
  }

  async expectLoginSuccess() {
    await expect(this.page).toHaveURL('/dashboard');
  }
}
```

### 테스트 격리
```typescript
// 각 테스트는 독립적인 상태로 시작
test.beforeEach(async ({ page }) => {
  await page.evaluate(() => localStorage.clear());
  await page.goto('/');
});
```

## CI/CD 통합

```yaml
# GitHub Actions
- name: Run E2E Tests
  run: npx playwright test --reporter=html
  env:
    BASE_URL: ${{ env.STAGING_URL }}

- name: Upload Test Report
  uses: actions/upload-artifact@v3
  with:
    name: playwright-report
    path: playwright-report/
```

## 체크리스트

- [ ] 핵심 사용자 플로우 커버
- [ ] 사용자 관점 셀렉터 사용
- [ ] 각 테스트 독립 실행 가능
- [ ] CI/CD에 통합됨
- [ ] 실패 시 스크린샷 자동 캡처
