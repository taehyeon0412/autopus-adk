# SPEC-ORCHCFG-002: Orchestra 기본 프로바이더 opencode에서 codex로 마이그레이션

**Status**: completed
**Created**: 2026-03-29
**Domain**: ORCHCFG

## 목적

codex CLI 0.117.0에 hooks 기능이 추가되어 opencode 대신 codex를 Orchestra 기본 프로바이더로 전환한다. 현재 코드에 `MigrateCodexToOpencode` 함수가 존재하여 autopus.yaml에 codex를 설정해도 config 로드 시 자동으로 opencode로 되돌리는 문제가 있다. 이 SPEC은 마이그레이션 방향을 역전하고, 모든 하드코딩된 기본값을 codex로 교체한다.

## 요구사항

### R1: 기본 프로바이더 엔트리 교체
WHEN the system loads default provider entries,
THE SYSTEM SHALL use codex as the default GPT provider instead of opencode, with binary `codex`, args `[exec, --approval-mode, full-auto, --quiet, -m, gpt-5.4]`, pane_args `[-m, gpt-5.4]`, PromptViaArgs `false`.

### R2: 마이그레이션 방향 역전
WHEN a config contains an opencode provider entry,
THE SYSTEM SHALL automatically migrate it to codex by removing the opencode entry, adding codex with default settings, and replacing opencode references in all command provider lists.

### R3: MigrateOpencodeToTUI 제거
WHEN performing orchestra config migration,
THE SYSTEM SHALL NOT execute the MigrateOpencodeToTUI migration step, as codex does not require TUI mode migration.

### R4: PlatformToProvider 매핑 업데이트
WHEN the platform is "opencode",
THE SYSTEM SHALL map it to the "codex" provider name in PlatformToProvider.

### R5: Fallback 프로바이더 업데이트
WHEN building fallback provider configs in orchestra_helpers.go,
THE SYSTEM SHALL use codex settings (binary: codex, args: [exec, --approval-mode, full-auto, --quiet, -m, gpt-5.4]) instead of opencode as the GPT-based fallback.

### R6: 완료 패턴 하위 호환
WHEN detecting provider completion patterns,
THE SYSTEM SHALL retain the opencode completion pattern (`Ask anything`) alongside the codex pattern (`codex>`) to support users who manually configure opencode.

### R7: 기존 사용자 자동 마이그레이션
WHEN a user's existing autopus.yaml contains opencode provider configuration,
THE SYSTEM SHALL automatically convert it to codex during config load, preserving the user's custom model settings where applicable.

### R8: DefaultFullConfig 업데이트
WHEN generating the default full configuration,
THE SYSTEM SHALL list codex (not opencode) in Orchestra providers and in all command provider lists (review, plan, secure, brainstorm).

## 생성 파일 상세

| 파일 | 변경 유형 | 역할 |
|------|-----------|------|
| `pkg/config/migrate.go` | 수정 | defaultProviderEntries codex 업데이트, MigrateCodexToOpencode → MigrateOpencodeToCodex 역전, MigrateOpencodeToTUI 제거, PlatformToProvider 업데이트 |
| `pkg/config/defaults.go` | 수정 | DefaultFullConfig에서 opencode → codex |
| `pkg/config/schema.go` | 수정 | validProviders에 codex 유지 확인 (이미 존재) |
| `internal/cli/orchestra_helpers.go` | 수정 | fallback knownProviders codex 업데이트 |
| `pkg/orchestra/types.go` | 확인 | opencode 패턴 유지, codex 패턴 확인 (이미 완료) |
| `pkg/config/migrate_opencode_test.go` | 수정/재작성 | MigrateOpencodeToCodex 테스트 |
| `pkg/config/migrate_opencode_tui_test.go` | 삭제 | MigrateOpencodeToTUI 제거에 따른 삭제 |
| `pkg/config/defaults_opencode_tui_test.go` | 삭제/수정 | opencode TUI 테스트 제거 |
| `pkg/config/migrate_helpers_test.go` | 수정 | PlatformToProvider 매핑 테스트 업데이트 |
| `pkg/config/defaults_test.go` | 수정 | DefaultFullConfig codex 검증 |
| `internal/cli/orchestra_config_test.go` | 수정 | fallback provider codex 검증 |
