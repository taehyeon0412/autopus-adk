# Implementation Plan: SPEC-MULTIPLATFORM-001

**SPEC**: SPEC-MULTIPLATFORM-001 — Multi-Platform Harness Integration
**Status**: draft
**Created**: 2026-04-01

---

## 전체 구조

```
Phase 0: 사전 준비 (공유 인프라) ─── sequential
Phase 1: 어댑터 리팩터링 ────────── parallel (Codex ∥ Gemini)
Phase 2: 기능 확장 ──────────────── parallel (Codex ∥ Gemini)
Phase 3: 통합 테스트 ────────────── sequential
Phase 4: 마무리 ─────────────────── sequential
```

---

## Phase 0: Shared Infrastructure (순차 실행 — Phase 1/2 이전)

### T0-1: 크로스 플랫폼 템플릿 헬퍼

| Field | Value |
|-------|-------|
| **Task ID** | T0-1 |
| **Description** | 플랫폼 공통 헬퍼 함수 구현 (TruncateToBytes, MapPermission, SkillList) |
| **Agent** | executor |
| **Mode** | sequential |
| **Complexity** | MEDIUM |
| **Requirements** | R21 |

**File Ownership:**
- `pkg/template/helpers.go` — NEW, 헬퍼 함수 구현
- `pkg/template/helpers_test.go` — NEW, 단위 테스트

**구현 상세:**
- `TruncateToBytes(content string, maxBytes int) string`: UTF-8 경계를 존중하며 바이트 수 기준 자르기
- `MapPermission(claudeMode, targetPlatform string) string`: Claude plan/act/bypass -> 플랫폼별 매핑
- `SkillList(cfg *config.HarnessConfig) []SkillMeta`: 하네스 설정에서 스킬 메타데이터 목록 추출

### T0-2: 공유 규칙/에이전트 템플릿

| Field | Value |
|-------|-------|
| **Task ID** | T0-2 |
| **Description** | 공유 규칙 및 에이전트 메타데이터 템플릿 생성 |
| **Agent** | executor |
| **Mode** | sequential |
| **Complexity** | MEDIUM |
| **Requirements** | R22, R23 |

**File Ownership:**
- `content/rules/lore-commit.md` — 기존 소스 (있으면 활용, 없으면 NEW)
- `content/rules/file-size-limit.md` — 기존 소스 (있으면 활용, 없으면 NEW)
- `content/rules/subagent-delegation.md` — 기존 소스 (있으면 활용, 없으면 NEW)
- `content/rules/language-policy.md` — 기존 소스 (있으면 활용, 없으면 NEW)
- `content/agents/executor.md` — 기존 소스 (있으면 활용, 없으면 NEW)
- `content/agents/reviewer.md` — 기존 소스 (있으면 활용, 없으면 NEW)
- `content/agents/planner.md` — 기존 소스 (있으면 활용, 없으면 NEW)
- `content/agents/debugger.md` — 기존 소스 (있으면 활용, 없으면 NEW)
- `content/agents/tester.md` — 기존 소스 (있으면 활용, 없으면 NEW)

### T0-3: 테스트 유틸리티

| Field | Value |
|-------|-------|
| **Task ID** | T0-3 |
| **Description** | 어댑터 공통 테스트 헬퍼 구현 |
| **Agent** | executor |
| **Mode** | sequential |
| **Complexity** | LOW |
| **Requirements** | R24 |

**File Ownership:**
- `pkg/adapter/testutil_test.go` — NEW, setupTestDir, assertFileExists, assertFileContains 등

---

## Phase 1: Adapter Refactoring (병렬 실행 — Codex ∥ Gemini)

### T1-1: Codex 어댑터 파일 분리

| Field | Value |
|-------|-------|
| **Task ID** | T1-1 |
| **Description** | codex.go (399줄) -> 관심사별 6개 파일로 분리, 300줄 제한 준수 |
| **Agent** | executor |
| **Mode** | parallel (with T1-2) |
| **Complexity** | HIGH |
| **Requirements** | R9 |

**File Ownership:**
- `pkg/adapter/codex/codex.go` — MODIFY, 코어 로직만 남김 (New, Generate, Update, Validate, Clean, Name, Version 등 ~150줄)
- `pkg/adapter/codex/codex_skills.go` — NEW, renderSkillTemplates + prepareSkillFiles 추출 (~100줄)
- `pkg/adapter/codex/codex_marker.go` — NEW, injectMarkerSection, replaceMarkerSection, removeMarkerSection 추출 (~80줄)
- `pkg/adapter/codex/codex_test.go` — MODIFY, 기존 테스트 유지 확인

**리팩터링 전략:**
1. `renderSkillTemplates()` -> `codex_skills.go`
2. `injectMarkerSection()`, `replaceMarkerSection()`, `removeMarkerSection()` -> `codex_marker.go`
3. `prepareFiles()` -> 각 파일 그룹별로 분리된 함수 호출로 변경
4. `checksum()` -> `codex_marker.go` 또는 별도 util

### T1-2: Gemini 어댑터 파일 분리

| Field | Value |
|-------|-------|
| **Task ID** | T1-2 |
| **Description** | gemini.go (424줄) -> 관심사별 6개 파일로 분리, 300줄 제한 준수 |
| **Agent** | executor |
| **Mode** | parallel (with T1-1) |
| **Complexity** | HIGH |
| **Requirements** | R19 |

**File Ownership:**
- `pkg/adapter/gemini/gemini.go` — MODIFY, 코어 로직만 남김 (~150줄)
- `pkg/adapter/gemini/gemini_skills.go` — NEW, renderSkillTemplates 추출 (~100줄)
- `pkg/adapter/gemini/gemini_marker.go` — NEW, 마커 관련 함수 추출 (~80줄)
- `pkg/adapter/gemini/gemini_test.go` — MODIFY, 기존 테스트 유지 확인

---

## Phase 2: Feature Extension (병렬 실행 — Codex ∥ Gemini)

### T2-1: Codex 커스텀 커맨드 + 템플릿

| Field | Value |
|-------|-------|
| **Task ID** | T2-1 |
| **Description** | Codex 커스텀 커맨드 생성 로직 + 템플릿 구현 |
| **Agent** | executor |
| **Mode** | parallel (with T2-2, T2-3, T2-4) |
| **Complexity** | MEDIUM |
| **Requirements** | R1 |

**File Ownership:**
- `pkg/adapter/codex/codex_prompts.go` — NEW, generatePrompts() (~120줄)
- `templates/codex/prompts/auto-plan.md.tmpl` — NEW (flat, 서브디렉토리 아님)
- `templates/codex/prompts/auto-go.md.tmpl` — NEW
- `templates/codex/prompts/auto-fix.md.tmpl` — NEW
- `templates/codex/prompts/auto-review.md.tmpl` — NEW
- `templates/codex/prompts/auto-sync.md.tmpl` — NEW
- `templates/codex/prompts/auto-idea.md.tmpl` — NEW

### T2-2: Codex 에이전트 + 훅 + 설정

| Field | Value |
|-------|-------|
| **Task ID** | T2-2 |
| **Description** | Codex 에이전트 정의, 훅, MCP/권한 설정 구현 |
| **Agent** | executor |
| **Mode** | parallel (with T2-1, T2-3, T2-4) |
| **Complexity** | HIGH |
| **Requirements** | R2, R3, R4, R6, R7 |

**File Ownership:**
- `pkg/adapter/codex/codex_agents.go` — NEW, generateAgents() (~150줄)
- `pkg/adapter/codex/codex_hooks.go` — NEW, generateHooks(), InstallHooks() 재구현 (~120줄)
- `pkg/adapter/codex/codex_settings.go` — NEW, generateConfig(), mapPermissions() (~120줄)
- `pkg/adapter/codex/codex_rules.go` — NEW, inlineRules(), overflowToSkills() (~100줄)
- `templates/codex/agents/executor.toml.tmpl` — NEW
- `templates/codex/agents/reviewer.toml.tmpl` — NEW
- `templates/codex/agents/planner.toml.tmpl` — NEW
- `templates/codex/agents/debugger.toml.tmpl` — NEW
- `templates/codex/agents/tester.toml.tmpl` — NEW
- `templates/codex/hooks.json.tmpl` — NEW
- `templates/codex/config.toml.tmpl` — NEW
- `templates/codex/agents-md.tmpl` — NEW (32KB 제한 규칙 인라인 템플릿)

### T2-3: Gemini 커스텀 커맨드 + 규칙

| Field | Value |
|-------|-------|
| **Task ID** | T2-3 |
| **Description** | Gemini 커스텀 커맨드 + @import 규칙 파일 구현 |
| **Agent** | executor |
| **Mode** | parallel (with T2-1, T2-2, T2-4) |
| **Complexity** | MEDIUM |
| **Requirements** | R11, R13 |

**File Ownership:**
- `pkg/adapter/gemini/gemini_commands.go` — NEW, generateCommands() (~120줄)
- `pkg/adapter/gemini/gemini_rules.go` — NEW, generateRules() (~100줄)
- `templates/gemini/commands/auto/plan.toml.tmpl` — NEW
- `templates/gemini/commands/auto/go.toml.tmpl` — NEW
- `templates/gemini/commands/auto/fix.toml.tmpl` — NEW
- `templates/gemini/commands/auto/review.toml.tmpl` — NEW
- `templates/gemini/commands/auto/sync.toml.tmpl` — NEW
- `templates/gemini/commands/auto/idea.toml.tmpl` — NEW
- `templates/gemini/rules/autopus/lore-commit.md.tmpl` — NEW
- `templates/gemini/rules/autopus/file-size-limit.md.tmpl` — NEW
- `templates/gemini/rules/autopus/subagent-delegation.md.tmpl` — NEW
- `templates/gemini/rules/autopus/language-policy.md.tmpl` — NEW

### T2-4: Gemini 에이전트 + 설정

| Field | Value |
|-------|-------|
| **Task ID** | T2-4 |
| **Description** | Gemini 에이전트 정의, MCP/권한 설정 구현 |
| **Agent** | executor |
| **Mode** | parallel (with T2-1, T2-2, T2-3) |
| **Complexity** | HIGH |
| **Requirements** | R12, R14, R15, R17 |

**File Ownership:**
- `pkg/adapter/gemini/gemini_agents.go` — NEW, generateAgents() (~150줄)
- `pkg/adapter/gemini/gemini_settings.go` — NEW, generateSettings(), mapPermissions() (~120줄)
- `templates/gemini/agents/executor.md.tmpl` — NEW
- `templates/gemini/agents/reviewer.md.tmpl` — NEW
- `templates/gemini/agents/planner.md.tmpl` — NEW
- `templates/gemini/agents/debugger.md.tmpl` — NEW
- `templates/gemini/agents/tester.md.tmpl` — NEW
- `templates/gemini/settings.json.tmpl` — NEW (기존 gemini_hooks.go의 orchestra 훅과 통합)

### T2-5: 서브에이전트 패턴 매핑 (스킬 템플릿 업데이트)

| Field | Value |
|-------|-------|
| **Task ID** | T2-5 |
| **Description** | auto-go 등 파이프라인 스킬의 서브에이전트 호출을 플랫폼별 패턴으로 업데이트 |
| **Agent** | executor |
| **Mode** | sequential (after T2-1~T2-4) |
| **Complexity** | MEDIUM |
| **Requirements** | R8, R15, R18 |

**File Ownership:**
- `templates/codex/skills/auto-go.md.tmpl` — MODIFY, spawn_agent 패턴 추가
- `templates/gemini/skills/auto-go/SKILL.md.tmpl` — MODIFY, @agent 패턴 추가

---

## Phase 3: Integration Tests (순차 실행 — Phase 2 완료 후)

### T3-1: 어댑터 단위 테스트

| Field | Value |
|-------|-------|
| **Task ID** | T3-1 |
| **Description** | Codex/Gemini 새 기능 단위 테스트 작성 |
| **Agent** | executor |
| **Mode** | parallel (Codex tests ∥ Gemini tests) |
| **Complexity** | MEDIUM |
| **Requirements** | R9, R19, R25 |

**File Ownership:**
- `pkg/adapter/codex/codex_prompts_test.go` — NEW
- `pkg/adapter/codex/codex_agents_test.go` — NEW
- `pkg/adapter/codex/codex_hooks_test.go` — NEW
- `pkg/adapter/codex/codex_settings_test.go` — NEW
- `pkg/adapter/codex/codex_rules_test.go` — NEW
- `pkg/adapter/gemini/gemini_commands_test.go` — NEW
- `pkg/adapter/gemini/gemini_agents_test.go` — NEW
- `pkg/adapter/gemini/gemini_settings_test.go` — NEW
- `pkg/adapter/gemini/gemini_rules_test.go` — NEW

### T3-2: E2E 통합 테스트

| Field | Value |
|-------|-------|
| **Task ID** | T3-2 |
| **Description** | auto init/update E2E 시나리오 테스트 (3 플랫폼) |
| **Agent** | executor |
| **Mode** | sequential (after T3-1) |
| **Complexity** | MEDIUM |
| **Requirements** | R26, R27, R28 |

**File Ownership:**
- `pkg/adapter/integration_test.go` — NEW, TestInitCodex, TestInitGemini, TestUpdatePreservation

---

## Phase 4: Finalization (순차 실행)

### T4-1: 300줄 제한 검증 + 매니페스트 업데이트

| Field | Value |
|-------|-------|
| **Task ID** | T4-1 |
| **Description** | 모든 Go 파일 300줄 이하 확인, 매니페스트 JSON 업데이트 |
| **Agent** | executor |
| **Mode** | sequential |
| **Complexity** | LOW |
| **Requirements** | R29, R30 |

**File Ownership:**
- `.autopus/claude-code-manifest.json` — MODIFY
- CI pipeline verification

### T4-2: E2E 시나리오 문서 업데이트

| Field | Value |
|-------|-------|
| **Task ID** | T4-2 |
| **Description** | E2E 시나리오 문서에 Codex/Gemini 시나리오 추가 |
| **Agent** | executor |
| **Mode** | sequential |
| **Complexity** | LOW |
| **Requirements** | — |

**File Ownership:**
- `.autopus/project/scenarios.md` — MODIFY

---

## 태스크 요약 테이블

| Task ID | Description | Agent | Mode | File Ownership | Complexity | Phase |
|---------|-------------|-------|------|----------------|------------|-------|
| T0-1 | 크로스 플랫폼 템플릿 헬퍼 | executor | sequential | `pkg/template/helpers.go` | MEDIUM | 0 |
| T0-2 | 공유 규칙/에이전트 소스 | executor | sequential | `content/rules/*.md`, `content/agents/*.md` | MEDIUM | 0 |
| T0-3 | 테스트 유틸리티 | executor | sequential | `pkg/adapter/testutil_test.go` | LOW | 0 |
| T1-1 | Codex 어댑터 파일 분리 | executor | parallel | `pkg/adapter/codex/*.go` | HIGH | 1 |
| T1-2 | Gemini 어댑터 파일 분리 | executor | parallel | `pkg/adapter/gemini/*.go` | HIGH | 1 |
| T2-1 | Codex 커스텀 커맨드 + 템플릿 | executor | parallel | `pkg/adapter/codex/codex_prompts.go`, `templates/codex/prompts/auto-*.md.tmpl` | MEDIUM | 2 |
| T2-2 | Codex 에이전트 + 훅 + 설정 | executor | parallel | `pkg/adapter/codex/codex_agents.go` 등 | HIGH | 2 |
| T2-3 | Gemini 커스텀 커맨드 + 규칙 | executor | parallel | `pkg/adapter/gemini/gemini_commands.go` 등 | MEDIUM | 2 |
| T2-4 | Gemini 에이전트 + 설정 | executor | parallel | `pkg/adapter/gemini/gemini_agents.go` 등 | HIGH | 2 |
| T2-5 | 서브에이전트 패턴 매핑 | executor | sequential | `templates/codex/skills/auto-go.md.tmpl` 등 | MEDIUM | 2 |
| T3-1 | 어댑터 단위 테스트 | executor | parallel | `pkg/adapter/codex/*_test.go`, `pkg/adapter/gemini/*_test.go` | MEDIUM | 3 |
| T3-2 | E2E 통합 테스트 | executor | sequential | `pkg/adapter/integration_test.go` | MEDIUM | 3 |
| T4-1 | 300줄 제한 + 매니페스트 | executor | sequential | `.autopus/claude-code-manifest.json` | LOW | 4 |
| T4-2 | E2E 시나리오 문서 | executor | sequential | `.autopus/project/scenarios.md` | LOW | 4 |

---

## 의존성 그래프

```
T0-1, T0-2, T0-3  (Phase 0 — sequential)
       │
       ▼
T1-1 ──┬── T1-2  (Phase 1 — parallel)
       │
       ▼
T2-1 ──┬── T2-2 ──┬── T2-3 ──┬── T2-4  (Phase 2 — parallel)
       │           │           │
       └───────────┴───────────┘
                   │
                   ▼
                 T2-5  (Phase 2 — sequential tail)
                   │
                   ▼
         T3-1 (parallel: codex ∥ gemini tests)
                   │
                   ▼
                 T3-2  (Phase 3 — E2E)
                   │
                   ▼
         T4-1 ──── T4-2  (Phase 4 — sequential)
```

---

## 예상 워크로드

| Phase | 태스크 수 | 예상 시간 | 병렬화 |
|-------|----------|----------|--------|
| Phase 0 | 3 | ~30분 | 순차 |
| Phase 1 | 2 | ~40분 | 병렬 (x2) |
| Phase 2 | 5 | ~60분 | 병렬 (x4) + 순차 (x1) |
| Phase 3 | 2 | ~40분 | 병렬 + 순차 |
| Phase 4 | 2 | ~15분 | 순차 |
| **Total** | **14** | **~3시간** | — |
