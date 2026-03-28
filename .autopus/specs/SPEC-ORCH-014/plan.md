# SPEC-ORCH-014 구현 계획

## 태스크 목록

- [ ] T1: autopus.yaml opencode provider 설정 변경 — `interactive_input` 제거, PaneArgs에서 `run` 제거
- [ ] T2: `pkg/config/defaults.go` opencode 기본 PaneArgs 수정 (`run` 제거, `-m openai/gpt-5.4`만 유지)
- [ ] T3: `pkg/config/migrate.go` opencode interactive TUI 마이그레이션 함수 추가 (`MigrateOpencodeToTUI`)
- [ ] T4: `pkg/config/migrate.go` `defaultProviderEntries` opencode 항목 PaneArgs 갱신
- [ ] T5: `pkg/orchestra/interactive_launch.go` opencode TUI launch 커맨드 빌드 동작 검증 및 필요 시 수정
- [ ] T6: `pkg/orchestra/interactive_debate.go` round 1 args 스킵 로직이 opencode에 영향 없음 검증
- [ ] T7: `pkg/adapter/opencode/opencode.go` `InjectOrchestraPlugin` 호출 흐름 확인
- [ ] T8: 기존 테스트 수정 — `InteractiveInput: "args"` → 빈 문자열로 변경된 케이스 반영
- [ ] T9: 새 테스트 추가 — opencode TUI 모드 launch 커맨드, 마이그레이션, debate round 프롬프트 전달

## 구현 전략

### 접근 방법

핵심 변경은 opencode의 `InteractiveInput`을 `"args"`에서 빈 문자열(기본값)로 전환하는 것이다.
이 변경으로 opencode는 claude/gemini와 동일한 코드 경로를 사용하게 된다:

1. **launch 시**: `opencode -m openai/gpt-5.4` (TUI 모드, `run` 없음)
2. **round 1 프롬프트**: SendLongText로 전달 (args 스킵 안 함)
3. **round 2+ 프롬프트**: SendLongText로 rebuttal 전달 (세션 유지)
4. **완료 감지**: hook 기반 (`hook-opencode-complete.ts`) 또는 screen polling (`> ` 패턴)

### 기존 코드 활용

- `buildInteractiveLaunchCmd()`: `InteractiveInput != "args"`이면 `run` 서브커맨드를 자동 스킵하므로 PaneArgs에서 `run` 제거만으로 충분
- `executeRound()`: `InteractiveInput == "args" && round == 1` 스킵 로직이 비활성화됨 → SendLongText 전달
- `hook-opencode-complete.ts`: 이미 round-scoped 시그널 파일 생성 지원 (`AUTOPUS_ROUND` env)
- `DefaultCompletionPatterns()`: opencode 패턴 `^>\s*$` 이미 등록됨

### 변경 범위

- **config layer**: defaults.go, migrate.go, autopus.yaml (3파일)
- **orchestra layer**: interactive_launch.go 확인만 (실제 변경 불필요할 수 있음)
- **test layer**: 기존 테스트 케이스 InteractiveInput 값 변경, 마이그레이션 테스트 추가

### 위험 요소

- opencode TUI가 `opencode` (인자 없이 실행) vs `opencode -m model`로 실행 시 동작 차이 확인 필요
- 기존 `opencode run -m` 비인터랙티브 모드 (Args)는 변경 없이 유지해야 함
- hook 플러그인이 TUI 모드에서도 `text.complete` 이벤트를 발화하는지 실환경 검증 필요
