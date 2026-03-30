# SPEC-ORCHCFG-002 구현 계획

## 태스크 목록

- [ ] T1: `pkg/config/migrate.go` — defaultProviderEntries의 codex 엔트리를 새 args로 업데이트 (exec, --approval-mode, full-auto, --quiet, -m, gpt-5.4), PromptViaArgs를 false로 변경
- [ ] T2: `pkg/config/migrate.go` — `MigrateCodexToOpencode` 함수를 `MigrateOpencodeToCodex`로 역전 (opencode 감지 → codex로 변환)
- [ ] T3: `pkg/config/migrate.go` — `MigrateOpencodeToTUI` 함수 및 호출 제거
- [ ] T4: `pkg/config/migrate.go` — `MigrateOrchestraConfig` 내 migration 호출 순서 업데이트 (1.5: MigrateOpencodeToCodex, 1.6 제거)
- [ ] T5: `pkg/config/migrate.go` — `PlatformToProvider`에서 "opencode" → "codex" 매핑
- [ ] T6: `pkg/config/defaults.go` — DefaultFullConfig의 Providers에서 opencode → codex, Commands의 provider 리스트에서 opencode → codex
- [ ] T7: `internal/cli/orchestra_helpers.go` — buildProviderConfigs의 knownProviders에서 codex 엔트리를 새 args로 업데이트
- [ ] T8: 테스트 업데이트 — migrate_opencode_test.go를 MigrateOpencodeToCodex 방향으로 재작성
- [ ] T9: 테스트 정리 — migrate_opencode_tui_test.go 삭제, defaults_opencode_tui_test.go 삭제 또는 codex용으로 수정
- [ ] T10: 테스트 업데이트 — migrate_helpers_test.go, defaults_test.go, orchestra_config_test.go에서 opencode → codex 참조 업데이트

## 구현 전략

### 접근 방법

1. **마이그레이션 함수 역전이 핵심**: `MigrateCodexToOpencode`의 로직을 그대로 반전하여 `MigrateOpencodeToCodex`를 만든다. opencode 키 감지 → 삭제 → codex 기본값 추가 → 커맨드 리스트에서 opencode를 codex로 교체.

2. **기존 `replaceInSlice` 유틸리티 재사용**: 커맨드 프로바이더 리스트 교체에 이미 존재하는 헬퍼를 활용한다.

3. **하위 호환성 유지**: opencode completion pattern은 `types.go`에 그대로 유지하여, 사용자가 수동으로 opencode를 설정하는 경우를 지원한다.

4. **Migration 1 (codex PromptViaArgs) 제거 검토**: codex의 새 설정에서 PromptViaArgs가 false이므로, Migration 1의 PromptViaArgs=true 강제 로직도 제거하거나 역전해야 한다.

### 변경 범위

- Go 소스 파일 3개 수정 (migrate.go, defaults.go, orchestra_helpers.go)
- 테스트 파일 5-6개 수정/삭제
- 총 예상 변경: ~150줄 수정, ~80줄 삭제

### 위험 요소

- Migration 1 (PromptViaArgs 강제) 로직이 codex 새 설정과 충돌할 수 있음 — 반드시 함께 업데이트
- opencode를 수동 설정한 기존 사용자의 config가 의도치 않게 codex로 변환될 수 있음 — MigrateOpencodeToCodex에서 사용자 커스텀 모델 보존 로직 필요
