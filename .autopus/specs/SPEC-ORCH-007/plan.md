# SPEC-ORCH-007 구현 계획

## Phase 구조

| Phase | 범위 | 태스크 | 우선순위 |
|-------|------|--------|---------|
| Phase 1 | Hook 파일 시그널 핵심 구현 | T1-T5 | P0 |
| Phase 2 | Codex → opencode 전환 | T6 | P1 |
| Phase 3 | 오케스트레이션 전략 Hook 연동 | T7 | P1 |
| Phase 4 | 테스트 작성 | T8 | P0 |

## 태스크 목록

- [ ] T1: Hook 파일 시그널 프로토콜 정의 (types, interfaces)
- [ ] T2: Claude Code Stop hook 스크립트 + 자동 주입
- [ ] T3: Gemini CLI AfterAgent hook 스크립트 + 자동 주입
- [ ] T4: opencode plugin 스크립트 + 자동 주입
- [ ] T5: interactive.go 리팩토링 (ReadScreen → 파일 감시)
- [ ] T6: Codex → opencode 프로바이더 전환
- [ ] T7: debate/relay/consensus 전략 hook 연동
- [ ] T8: 테스트 작성

## 태스크 상세

### T1: Hook 파일 시그널 프로토콜 정의

| 항목 | 값 |
|------|---|
| Agent | executor |
| Mode | bypassPermissions |
| 의존성 | 없음 (독립 시작 가능) |
| 예상 줄 수 | ~140줄 |

**File Ownership**:
- `pkg/orchestra/hook_signal.go` (신규) — HookResult 구조체, 세션 디렉토리 생성/정리, result.json 파싱, done 파일 감시

**작업 내용**:
1. `HookResult` 구조체 정의: `SessionID`, `Provider`, `Response`, `Timestamp`
2. `NewHookSession(providers []string) (*HookSession, error)` — `/tmp/autopus/{session-id}/` 디렉토리 생성
3. `(*HookSession) WaitForDone(ctx, provider) (string, error)` — os.Stat 200ms 폴링으로 done 파일 감시
4. `(*HookSession) ReadResult(provider) (*HookResult, error)` — result.json 파싱
5. `(*HookSession) Cleanup()` — 세션 디렉토리 삭제
6. `(*HookSession) HasHook(provider) bool` — 프로바이더의 hook 설정 존재 여부 확인

### T2: Claude Code Stop Hook 스크립트 + 자동 주입

| 항목 | 값 |
|------|---|
| Agent | executor |
| Mode | bypassPermissions |
| 의존성 | T1 (프로토콜 정의) |
| 예상 줄 수 | 스크립트 ~25줄, 주입 로직 ~30줄 |

**File Ownership**:
- `content/hooks/hook-claude-stop.sh` (신규) — Stop hook 스크립트
- `pkg/adapter/claude/claude_settings.go` (수정) — Stop hook 등록

**작업 내용**:
1. Shell 스크립트 작성: stdin JSON → `python3 -c` 파싱 → result.json 저장 → done 시그널
2. `claude_settings.go`의 `mergeHooks()`에 Stop hook 엔트리 추가
3. `auto init` 시 hook 스크립트를 프로젝트 `.claude/hooks/` 에 복사

### T3: Gemini CLI AfterAgent Hook 스크립트 + 자동 주입

| 항목 | 값 |
|------|---|
| Agent | executor |
| Mode | bypassPermissions |
| 의존성 | T1 (프로토콜 정의) |
| 예상 줄 수 | 스크립트 ~25줄, 주입 로직 ~40줄 |

**File Ownership**:
- `content/hooks/hook-gemini-afteragent.sh` (신규) — AfterAgent hook 스크립트
- `pkg/adapter/gemini/gemini.go` (수정) — AfterAgent hook 등록

**작업 내용**:
1. Shell 스크립트 작성: stdin JSON → `python3 -c` 파싱 → result.json 저장 → done 시그널
2. `gemini.go`에 hook 주입 로직 추가 (`.gemini/settings.json` 생성/머지)
3. 기존 사용자 hook 보존

### T4: opencode Plugin 스크립트 + 자동 주입

| 항목 | 값 |
|------|---|
| Agent | executor |
| Mode | bypassPermissions |
| 의존성 | T1 (프로토콜 정의) |
| 예상 줄 수 | 스크립트 ~25줄, 어댑터 ~150줄 |

**File Ownership**:
- `content/hooks/hook-opencode-complete.ts` (신규) — opencode plugin 스크립트
- `pkg/adapter/opencode/opencode.go` (신규) — opencode PlatformAdapter 구현

**작업 내용**:
1. TypeScript plugin 스크립트 작성: text 필드 → result.json 저장 → done 시그널
2. opencode PlatformAdapter 구현: `PlatformAdapter` 인터페이스, `opencode.json` 생성/머지
3. plugin 자동 등록 로직

### T5: interactive.go 리팩토링 (ReadScreen → 파일 감시)

| 항목 | 값 |
|------|---|
| Agent | executor |
| Mode | bypassPermissions |
| 의존성 | T1 (프로토콜 정의) |
| 예상 줄 수 | hook_watcher.go ~140줄, interactive.go 수정 ~30줄 |

**File Ownership**:
- `pkg/orchestra/hook_watcher.go` (신규) — Hook 모드 대기/수집 로직
- `pkg/orchestra/interactive.go` (수정) — hook/ReadScreen 분기 추가
- `pkg/orchestra/types.go` (수정) — HookMode, SessionID 필드 추가

**작업 내용**:
1. `hook_watcher.go` 작성:
   - `waitAndCollectHookResults()` — 병렬로 각 프로바이더의 done 파일 감시 + result.json 파싱
   - 프로바이더별 hook 존재 여부에 따라 hook/ReadScreen 분기
2. `interactive.go` 수정:
   - `RunInteractivePaneOrchestra()`에서 HookSession 생성
   - `waitAndCollectResults()` 호출부를 hook/fallback 분기로 변경
   - 환경변수 `AUTOPUS_SESSION_ID` 설정하여 pane에 전달
3. `types.go` 수정:
   - `OrchestraConfig`에 `HookMode bool`, `SessionID string` 추가

### T6: Codex → opencode 프로바이더 전환

| 항목 | 값 |
|------|---|
| Agent | executor |
| Mode | bypassPermissions |
| 의존성 | T4 (opencode 어댑터) |
| 예상 줄 수 | ~40줄 수정 |

**File Ownership**:
- `pkg/config/migrate.go` (수정) — codex → opencode 마이그레이션
- `pkg/orchestra/relay.go` (수정) — `agenticArgs()` opencode 분기 추가

**작업 내용**:
1. `defaultProviderEntries`에 opencode 엔트리 추가
2. `PlatformToProvider()`에 opencode 매핑 추가
3. `MigrateOrchestraConfig()`에 codex → opencode 자동 변환 마이그레이션 추가
4. `agenticArgs()`에 opencode 분기 추가

### T7: debate/relay/consensus 전략 Hook 연동

| 항목 | 값 |
|------|---|
| Agent | executor |
| Mode | bypassPermissions |
| 의존성 | T5 (interactive.go 리팩토링) |
| 예상 줄 수 | ~20줄 수정 |

**File Ownership**:
- `pkg/orchestra/debate.go` (수정) — hook 결과 기반 rebuttal
- `pkg/orchestra/relay_pane.go` (수정) — hook 결과 기반 relay

**작업 내용**:
1. hook 결과가 `ProviderResponse.Output`에 이미 저장되므로 기존 전략 코드 변경 최소화
2. 인터랙티브 모드 debate에서 rebuttal 시 hook 결과 활용 확인
3. relay_pane에서 hook 결과를 다음 프로바이더 프롬프트에 주입 확인

**참고**: research.md D4 결정에 의해 hook 결과가 `.Output`에 저장되므로, 전략 코드 자체의 변경은 거의 없다. 주로 인터랙티브 모드에서의 통합 테스트 수준.

### T8: 테스트 작성

| 항목 | 값 |
|------|---|
| Agent | tester |
| Mode | bypassPermissions |
| 의존성 | T1-T7 전체 |
| 예상 줄 수 | ~250줄 (2-3 파일) |

**File Ownership**:
- `pkg/orchestra/hook_signal_test.go` (신규) — HookSession 유닛 테스트
- `pkg/orchestra/hook_watcher_test.go` (신규) — Hook 대기/수집 유닛 테스트
- `pkg/config/migrate_test.go` (수정) — codex → opencode 마이그레이션 테스트

**작업 내용**:
1. HookSession 테스트: 세션 생성, result.json 파싱, done 파일 감지, cleanup
2. Hook watcher 테스트: hook/ReadScreen 분기, 타임아웃, graceful degradation
3. Config 마이그레이션 테스트: codex → opencode 전환 검증

## 구현 전략

### 접근 방법

1. **점진적 전환**: hook 모드를 기존 ReadScreen 위에 레이어로 추가하여 fallback 보장
2. **Output 필드 재사용**: hook 결과를 기존 `ProviderResponse.Output`에 저장하여 전략 코드 변경 최소화
3. **프로바이더별 독립 hook**: 각 프로바이더의 hook은 독립적으로 동작하여 부분 실패 허용
4. **기존 인프라 활용**: `randomHex()`, `sanitizeProviderName()`, `mergeHooks()` 등 재사용

### 병렬 실행 가능한 태스크

```
T1 ──────────────┐
                  ├── T5 ──── T7 ──── T8
T2 (T1 의존) ────┘
T3 (T1 의존) ────┘
T4 (T1 의존) ──── T6
```

- T1 완료 후 T2, T3, T4, T5를 병렬로 시작 가능
- T6은 T4 완료 후 시작
- T7은 T5 완료 후 시작
- T8은 모든 태스크 완료 후 시작

### 변경 범위

| 카테고리 | 신규 | 수정 |
|---------|------|------|
| Go 소스 | 3개 (hook_signal.go, hook_watcher.go, opencode.go) | 7개 |
| Hook 스크립트 | 3개 (shell x2, TS x1) | 0 |
| 테스트 | 2-3개 | 1개 |
| 총 예상 줄 수 | ~500줄 | ~160줄 |
