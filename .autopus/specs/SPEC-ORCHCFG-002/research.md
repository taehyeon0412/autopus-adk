# SPEC-ORCHCFG-002 리서치

## 기존 코드 분석

### 핵심 변경 대상

#### 1. `pkg/config/migrate.go`

- **L6-11 `defaultProviderEntries`**: 현재 codex 엔트리는 `{Binary: "codex", Args: ["--quiet"], PaneArgs: ["--quiet"], PromptViaArgs: true}`. 새 설정으로 교체 필요: `{Binary: "codex", Args: ["exec", "--approval-mode", "full-auto", "--quiet", "-m", "gpt-5.4"], PaneArgs: ["-m", "gpt-5.4"], PromptViaArgs: false}`.
- **L29-36 Migration 1**: `codex.PromptViaArgs`를 true로 강제하는 로직. codex 새 설정에서는 PromptViaArgs=false이므로 이 migration 제거 또는 방향 역전 필요.
- **L38-41 Migration 1.5**: `MigrateCodexToOpencode` 호출. → `MigrateOpencodeToCodex`로 교체.
- **L43-46 Migration 1.6**: `MigrateOpencodeToTUI` 호출. → 제거.
- **L127-140 `PlatformToProvider`**: `"opencode" → "opencode"` 매핑을 `"opencode" → "codex"`로 변경.
- **L142-174 `MigrateCodexToOpencode`**: opencode→codex 방향으로 역전. 함수명을 `MigrateOpencodeToCodex`로 변경하고, 내부 로직에서 "codex"와 "opencode"를 swap.
- **L176-219 `MigrateOpencodeToTUI`**: 전체 삭제. codex는 TUI 마이그레이션이 불필요.

#### 2. `pkg/config/defaults.go`

- **L65-69**: `DefaultFullConfig`의 Orchestra.Providers에서 `"opencode"` 키를 `"codex"`로 교체. args를 `["exec", "--approval-mode", "full-auto", "--quiet", "-m", "gpt-5.4"]`, pane_args를 `["-m", "gpt-5.4"]`, PromptViaArgs=false.
- **L70-75**: Commands의 Providers 리스트에서 `"opencode"`를 `"codex"`로 교체.

#### 3. `internal/cli/orchestra_helpers.go`

- **L95-100 `buildProviderConfigs`**: knownProviders의 `"codex"` 엔트리를 새 args로 업데이트. `"opencode"` 엔트리는 하위 호환을 위해 유지하거나 제거 (fallback이므로 제거 권장).

#### 4. `pkg/orchestra/types.go`

- **L96-103 `DefaultCompletionPatterns`**: codex(`codex>`)와 opencode(`Ask anything`) 패턴 모두 이미 존재. 변경 불필요.

#### 5. `pkg/orchestra/hook_signal.go`

- **L30**: `defaultHookProviders`에 codex가 이미 설정됨. 변경 불필요.

### 테스트 파일 현황

| 파일 | 현재 상태 | 필요 조치 |
|------|-----------|-----------|
| `pkg/config/migrate_opencode_test.go` | MigrateCodexToOpencode 테스트 | MigrateOpencodeToCodex로 역전 재작성 |
| `pkg/config/migrate_opencode_tui_test.go` | MigrateOpencodeToTUI 테스트 | 삭제 (함수 제거됨) |
| `pkg/config/defaults_opencode_tui_test.go` | opencode TUI 기본값 테스트 | 삭제 |
| `pkg/config/migrate_spec015_test.go` | SPEC-015 마이그레이션 테스트 | codex 방향으로 업데이트 |
| `pkg/config/defaults_test.go` | DefaultFullConfig 테스트 | opencode → codex 검증으로 변경 |
| `pkg/config/migrate_helpers_test.go` | PlatformToProvider 등 헬퍼 테스트 | opencode→codex 매핑 업데이트 |
| `internal/cli/orchestra_config_test.go` | buildProviderConfigs 테스트 | codex 엔트리 검증으로 변경 |
| `pkg/orchestra/hook_signal_test.go` | hook provider 테스트 | 이미 codex로 변경됨, 변경 불필요 |

### 이미 완료된 변경

다음 파일은 이미 codex로 전환 완료:
- `autopus.yaml` (root) — codex 설정
- `autopus-adk/autopus.yaml` — codex 설정
- `content/hooks/hook-codex-stop.sh` — codex stop hook
- `pkg/orchestra/hook_signal.go:30` — defaultHookProviders에 codex

## 설계 결정

### D1: 마이그레이션 방향 역전 (역전 vs 새 함수)

**결정**: `MigrateCodexToOpencode`를 `MigrateOpencodeToCodex`로 rename하고 로직을 역전.

**이유**: 기존 함수와 1:1 대응이 명확하고, 코드 리뷰에서 변경 범위가 쉽게 추적됨.

**대안**: 새 함수를 별도로 작성하고 기존 함수를 deprecated 처리. → 불필요한 코드 잔존.

### D2: opencode 엔트리 완전 제거 vs 유지

**결정**: defaultProviderEntries에서 opencode 엔트리 유지 (하위 호환). DefaultFullConfig에서는 제거.

**이유**: 사용자가 autopus.yaml에 직접 opencode를 추가하는 경우, defaultProviderEntries에서 기본값을 제공할 수 있어야 함. 다만 시스템이 자동으로 opencode를 기본 설정하지는 않도록 DefaultFullConfig에서는 제거.

### D3: Migration 1 (PromptViaArgs 강제) 처리

**결정**: Migration 1 제거.

**이유**: codex 새 설정에서 PromptViaArgs=false. 기존 Migration 1은 codex의 PromptViaArgs를 true로 강제하는데, 이는 새 설정과 충돌. MigrateOpencodeToCodex가 올바른 기본값을 설정하므로 별도 강제 불필요.

### D4: codex args 설계

**결정**: `[exec, --approval-mode, full-auto, --quiet, -m, gpt-5.4]`

**이유**:
- `exec`: 비대화형 모드 (opencode의 `run`에 해당)
- `--approval-mode full-auto`: 자동 승인으로 자율 실행
- `--quiet`: 불필요한 출력 억제
- `-m gpt-5.4`: 모델 지정
- PromptViaArgs=false: stdin 기반 프롬프트 전달 (긴 프롬프트에서 ENAMETOOLONG 방지)

### D5: PlatformToProvider "opencode" 매핑

**결정**: `"opencode"` platform을 `"codex"` provider로 매핑.

**이유**: opencode platform을 사용하는 기존 사용자가 자동으로 codex provider를 얻도록 함. platform 이름 자체를 변경하지 않아 기존 autopus.yaml의 platforms 리스트를 깨뜨리지 않음.
