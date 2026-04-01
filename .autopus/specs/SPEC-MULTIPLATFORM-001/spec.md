---
id: SPEC-MULTIPLATFORM-001
title: Multi-Platform Harness Integration
status: completed
target_module: autopus-adk
created: 2026-04-01
---

# SPEC-MULTIPLATFORM-001: Multi-Platform Harness Integration

**Status**: completed
**Created**: 2026-04-01
**Target Module**: autopus-adk

## 목적

Codex CLI와 Gemini CLI 어댑터를 확장하여 Claude Code 수준의 하네스 패리티를 달성한다.
현재 스킬 6개만 생성하는 어댑터를 커스텀 커맨드, 에이전트 정의, 규칙, 훅, MCP 설정까지 포함하도록 확장한다.

---

## Domain 1: Codex Adapter Extension (R1-R10)

### R1 — Codex 커스텀 커맨드 생성

> WHEN `auto init --platform codex` 실행 시,
> THE SYSTEM SHALL `/auto` 서브커맨드(plan, go, fix, review, sync, idea)당 하나의 `.md` 파일을 `.codex/prompts/`에 flat 구조로 생성한다. Codex는 최상위 Markdown 파일만 스캔하며 서브디렉토리는 무시한다.

**Acceptance Criteria:**
- AC-1.1: `.codex/prompts/auto-plan.md`, `auto-go.md`, `auto-fix.md`, `auto-review.md`, `auto-sync.md`, `auto-idea.md` 총 6개 파일 존재 (flat, 서브디렉토리 아님)
- AC-1.2: 각 파일에 YAML front matter (`description`, `argument-hint`) 포함
- AC-1.3: 템플릿 소스는 `templates/codex/prompts/auto-*.md.tmpl`

**File Ownership:** `pkg/adapter/codex/codex_prompts.go`, `templates/codex/prompts/auto-*.md.tmpl`

### R2 — Codex 에이전트 정의 생성

> WHEN `auto init --platform codex` 실행 시,
> THE SYSTEM SHALL `.codex/agents/`에 executor, reviewer, planner 등 핵심 에이전트를 TOML 파일로 생성한다.

**Acceptance Criteria:**
- AC-2.1: `.codex/agents/executor.toml`, `reviewer.toml`, `planner.toml`, `debugger.toml`, `tester.toml` 최소 5개 파일 존재
- AC-2.2: 각 TOML 파일에 `name`, `description`, `developer_instructions`, `model` 필드 포함
- AC-2.3: `developer_instructions`는 Claude `.claude/agents/autopus/*.md`의 핵심 지시를 변환한 내용

**File Ownership:** `pkg/adapter/codex/codex_agents.go`, `templates/codex/agents/*.toml.tmpl`

### R3 — Codex 훅 설정 생성

> WHEN `auto init --platform codex` 실행 시,
> THE SYSTEM SHALL `hooks.json`에 SessionStart, PreToolUse, PostToolUse, Stop 이벤트를 등록한다.

**Acceptance Criteria:**
- AC-3.1: `.codex/hooks.json` 파일이 유효한 JSON으로 생성됨
- AC-3.2: SessionStart 이벤트에 하네스 검증 훅 등록
- AC-3.3: PreToolUse 이벤트에 파일 크기 체크 훅 등록
- AC-3.4: PostToolUse 이벤트에 로그 수집 훅 등록
- AC-3.5: `SupportsHooks()`가 `true`를 반환하도록 변경

**File Ownership:** `pkg/adapter/codex/codex_hooks.go`, `templates/codex/hooks.json.tmpl`

### R4 — Codex MCP 설정 생성

> WHEN `auto init --platform codex` 실행 시,
> THE SYSTEM SHALL `config.toml`에 `[mcp_servers]` 섹션을 추가하여 Context7 등 MCP 서버를 등록한다.

**Acceptance Criteria:**
- AC-4.1: `.codex/config.toml` 또는 프로젝트 루트 `config.toml`에 `[mcp_servers]` 섹션 존재
- AC-4.2: Context7 서버 엔트리가 포함됨 (command, args, env)
- AC-4.3: 기존 config.toml 내용이 보존됨 (OverwriteMerge 정책)

**File Ownership:** `pkg/adapter/codex/codex_settings.go`, `templates/codex/config.toml.tmpl`

### R5 — Codex 스킬 확장 (기존 6개 -> 전체)

> WHEN `auto init --platform codex` 실행 시,
> THE SYSTEM SHALL Claude의 모든 스킬을 Codex 네이티브 포맷으로 변환하여 `.codex/skills/`에 생성한다.

**Acceptance Criteria:**
- AC-5.1: Claude `templates/claude/skills/*.md.tmpl`의 모든 스킬이 대응하는 Codex 스킬로 변환됨
- AC-5.2: 기존 6개 스킬(auto-plan, auto-go, auto-fix, auto-review, auto-sync, auto-idea) 유지
- AC-5.3: 추가 스킬 (agent-pipeline, frontend-verify 등) 포함

**File Ownership:** `pkg/adapter/codex/codex_skills.go` (기존 renderSkillTemplates 추출), `templates/codex/skills/*.md.tmpl`

### R6 — Codex AGENTS.md 규칙 인라인 (32KB 제한)

> WHEN AGENTS.md 생성 시,
> THE SYSTEM SHALL 핵심 규칙(lore-commit, file-size-limit, subagent-delegation, language-policy)을 32KB 제한 내에서 인라인하고, 초과분은 스킬 참조로 대체한다.

**Acceptance Criteria:**
- AC-6.1: AGENTS.md의 AUTOPUS 마커 섹션이 32KB (32,768 bytes) 이하
- AC-6.2: lore-commit, file-size-limit, subagent-delegation, language-policy 규칙이 인라인됨
- AC-6.3: 인라인되지 않은 규칙은 "See `.codex/skills/{rule-name}.md` for details" 참조문 포함
- AC-6.4: 사용자 기존 AGENTS.md 마커 외부 내용 보존

**File Ownership:** `pkg/adapter/codex/codex_rules.go`, `templates/codex/agents-md.tmpl`

### R7 — Codex 권한 모드 매핑

> WHEN `auto init --platform codex` 실행 시 권한 설정이 포함된 경우,
> THE SYSTEM SHALL Claude의 권한 모드를 Codex 네이티브 모드로 매핑한다.

**Acceptance Criteria:**
- AC-7.1: Claude `plan` -> Codex `on-request` 매핑
- AC-7.2: Claude `act` -> Codex `untrusted` 매핑
- AC-7.3: Claude `bypass` -> Codex `never` 매핑
- AC-7.4: 매핑된 권한이 config.toml에 반영됨

**File Ownership:** `pkg/adapter/codex/codex_settings.go`

### R8 — Codex 서브에이전트 패턴 매핑

> WHEN 파이프라인 스킬(auto-go) 변환 시,
> THE SYSTEM SHALL Claude의 `Agent()` 호출을 Codex의 `spawn_agent` 패턴으로 매핑한다.

**Acceptance Criteria:**
- AC-8.1: auto-go 스킬 내 서브에이전트 위임 지시가 Codex `spawn_agent` 구문 사용
- AC-8.2: 에이전트 이름이 `.codex/agents/` 내 정의된 에이전트와 일치

**File Ownership:** `templates/codex/skills/auto-go.md.tmpl`

### R9 — Codex Generate() 리팩터링

> WHEN Codex 어댑터의 Generate() 실행 시,
> THE SYSTEM SHALL prompts, agents, skills, hooks, settings 각 파일 그룹을 별도 함수에서 생성한다.

**Acceptance Criteria:**
- AC-9.1: `codex.go`가 200줄 이하 (Generate/Update/Validate/Clean 코어 로직)
- AC-9.2: `codex_prompts.go`, `codex_agents.go`, `codex_hooks.go`, `codex_settings.go`, `codex_skills.go`, `codex_rules.go` 분리
- AC-9.3: 각 파일이 300줄 이하
- AC-9.4: 기존 테스트 `codex_test.go`, `codex_extra_test.go` 통과

**File Ownership:** `pkg/adapter/codex/*.go`

### R10 — Codex Update() 마커/머지 전략

> WHEN `auto update --platform codex` 실행 시,
> THE SYSTEM SHALL AGENTS.md는 마커 전략, hooks.json/config.toml은 머지 전략으로 업데이트한다.

**Acceptance Criteria:**
- AC-10.1: AGENTS.md의 AUTOPUS:BEGIN/END 마커 섹션만 업데이트, 외부 보존
- AC-10.2: hooks.json의 autopus 관련 훅만 업데이트, 사용자 추가 훅 보존
- AC-10.3: config.toml의 `[mcp_servers]` 섹션만 업데이트

**File Ownership:** `pkg/adapter/codex/codex.go`

---

## Domain 2: Gemini Adapter Extension (R11-R20)

### R11 — Gemini 커스텀 커맨드 생성

> WHEN `auto init --platform gemini` 실행 시,
> THE SYSTEM SHALL `/auto` 서브커맨드당 하나의 `.toml` 파일을 `.gemini/commands/auto/`에 namespace 지원으로 생성한다.

**Acceptance Criteria:**
- AC-11.1: `.gemini/commands/auto/plan.toml`, `go.toml`, `fix.toml`, `review.toml`, `sync.toml`, `idea.toml` 총 6개 파일 존재
- AC-11.2: 각 TOML 파일에 `prompt` 필드 포함 (필수), `description` 선택
- AC-11.3: namespace가 `auto`로 설정되어 `/auto:plan` 형식으로 호출 가능

**File Ownership:** `pkg/adapter/gemini/gemini_commands.go`, `templates/gemini/commands/auto/*.toml.tmpl`

### R12 — Gemini 에이전트 정의 생성

> WHEN `auto init --platform gemini` 실행 시,
> THE SYSTEM SHALL `.gemini/agents/`에 executor, reviewer, planner 등 핵심 에이전트를 YAML frontmatter Markdown 파일로 생성한다.

**Acceptance Criteria:**
- AC-12.1: `.gemini/agents/executor.md`, `reviewer.md`, `planner.md`, `debugger.md`, `tester.md` 최소 5개 파일 존재
- AC-12.2: 각 파일에 YAML frontmatter (`name`, `description`, `model`, `tools`) 포함
- AC-12.3: Markdown 본문에 에이전트 지시사항 포함

**File Ownership:** `pkg/adapter/gemini/gemini_agents.go`, `templates/gemini/agents/*.md.tmpl`

### R13 — Gemini 규칙 파일 생성 + @import

> WHEN `auto init --platform gemini` 실행 시,
> THE SYSTEM SHALL `.gemini/rules/autopus/`에 규칙 파일을 생성하고 GEMINI.md에서 `@path` 네이티브 구문으로 참조한다.

**Acceptance Criteria:**
- AC-13.1: `.gemini/rules/autopus/` 디렉터리에 `lore-commit.md`, `file-size-limit.md`, `subagent-delegation.md` 등 규칙 파일 생성
- AC-13.2: GEMINI.md에 `@.gemini/rules/autopus/lore-commit.md` 형식의 네이티브 `@path` 참조 포함
- AC-13.3: 규칙 내용이 Claude `.claude/rules/autopus/`의 대응 파일과 동등

**File Ownership:** `pkg/adapter/gemini/gemini_rules.go`, `templates/gemini/rules/autopus/*.md.tmpl`

### R14 — Gemini 훅 설정 생성

> WHEN `auto init --platform gemini` 실행 시,
> THE SYSTEM SHALL `settings.json`에 BeforeAgent, AfterAgent, BeforeTool, AfterTool 이벤트를 등록한다.

**Acceptance Criteria:**
- AC-14.1: `.gemini/settings.json` 파일이 유효한 JSON으로 생성됨
- AC-14.2: `hooks` 객체에 BeforeAgent, AfterAgent, BeforeTool, AfterTool 이벤트 등록
- AC-14.3: `SupportsHooks()`가 `true`를 반환하도록 변경
- AC-14.4: 기존 `gemini_hooks.go`와 통합 (orchestra 훅과 공존)

**File Ownership:** `pkg/adapter/gemini/gemini_hooks.go` (기존 파일 확장), `templates/gemini/settings.json.tmpl`

### R15 — Gemini MCP 설정 생성

> WHEN `auto init --platform gemini` 실행 시,
> THE SYSTEM SHALL `settings.json`에 `mcpServers`를 추가하여 Context7 등 MCP 서버를 등록한다.

**Acceptance Criteria:**
- AC-15.1: `.gemini/settings.json`의 `mcpServers` 키에 Context7 서버 엔트리 포함
- AC-15.2: 기존 settings.json 내용이 보존됨 (JSON merge 전략)
- AC-15.3: MCP 서버 설정이 `command`, `args`, `env` 필드를 포함

**File Ownership:** `pkg/adapter/gemini/gemini_settings.go`, `templates/gemini/settings.json.tmpl`

### R16 — Gemini 스킬 확장 (기존 6개 -> 전체)

> WHEN `auto init --platform gemini` 실행 시,
> THE SYSTEM SHALL Claude의 모든 스킬을 Gemini 네이티브 포맷으로 변환하여 `.gemini/skills/autopus/`에 생성한다.

**Acceptance Criteria:**
- AC-16.1: 기존 6개 스킬 유지 + 추가 스킬 포함
- AC-16.2: 각 스킬이 `.gemini/skills/autopus/{name}/SKILL.md` 디렉터리 구조 유지
- AC-16.3: 템플릿 소스는 `templates/gemini/skills/{name}/SKILL.md.tmpl`

**File Ownership:** `pkg/adapter/gemini/gemini_skills.go` (기존 renderSkillTemplates 추출), `templates/gemini/skills/*/SKILL.md.tmpl`

### R17 — Gemini 권한 모드 매핑

> WHEN `auto init --platform gemini` 실행 시 권한 설정이 포함된 경우,
> THE SYSTEM SHALL Claude의 권한 모드를 Gemini 네이티브 모드로 매핑한다.

**Acceptance Criteria:**
- AC-17.1: Claude `plan` -> Gemini `plan` 매핑
- AC-17.2: Claude `act` -> Gemini `auto_edit` 매핑
- AC-17.3: Claude `bypass` -> Gemini `yolo` 매핑
- AC-17.4: 매핑된 권한이 settings.json에 반영됨

**File Ownership:** `pkg/adapter/gemini/gemini_settings.go`

### R18 — Gemini 서브에이전트 패턴 매핑

> WHEN 파이프라인 스킬(auto-go) 변환 시,
> THE SYSTEM SHALL Claude의 `Agent()` 호출을 Gemini의 `@agent` tool 패턴으로 매핑한다.

**Acceptance Criteria:**
- AC-18.1: auto-go 스킬 내 서브에이전트 위임 지시가 Gemini `@agent` 구문 사용
- AC-18.2: 에이전트 이름이 `.gemini/agents/` 내 정의된 에이전트와 일치

**File Ownership:** `templates/gemini/skills/auto-go/SKILL.md.tmpl`

### R19 — Gemini Generate() 리팩터링

> WHEN Gemini 어댑터의 Generate() 실행 시,
> THE SYSTEM SHALL commands, agents, skills, hooks, settings, rules 각 파일 그룹을 별도 함수에서 생성한다.

**Acceptance Criteria:**
- AC-19.1: `gemini.go`가 200줄 이하 (Generate/Update/Validate/Clean 코어 로직)
- AC-19.2: `gemini_commands.go`, `gemini_agents.go`, `gemini_hooks.go`, `gemini_settings.go`, `gemini_skills.go`, `gemini_rules.go` 분리
- AC-19.3: 각 파일이 300줄 이하
- AC-19.4: 기존 테스트 `gemini_test.go`, `gemini_extra_test.go`, `gemini_hooks_test.go` 통과

**File Ownership:** `pkg/adapter/gemini/*.go`

### R20 — Gemini Update() 마커/머지 전략

> WHEN `auto update --platform gemini` 실행 시,
> THE SYSTEM SHALL GEMINI.md는 마커 전략, settings.json은 머지 전략으로 업데이트한다.

**Acceptance Criteria:**
- AC-20.1: GEMINI.md의 AUTOPUS:BEGIN/END 마커 섹션만 업데이트, 외부 보존
- AC-20.2: settings.json의 autopus 관련 키만 업데이트, 사용자 설정 보존
- AC-20.3: `.gemini/rules/autopus/` 파일은 항상 덮어쓰기 (OverwriteAlways)

**File Ownership:** `pkg/adapter/gemini/gemini.go`

---

## Domain 3: Shared Infrastructure (R21-R25)

### R21 — 크로스 플랫폼 템플릿 헬퍼

> WHEN 템플릿 렌더링 시,
> THE SYSTEM SHALL 플랫폼 공통 헬퍼 함수(truncateToBytes, mapPermission, skillList)를 제공한다.

**Acceptance Criteria:**
- AC-21.1: `pkg/template/helpers.go`에 `TruncateToBytes(content string, maxBytes int) string` 구현
- AC-21.2: `MapPermission(claudeMode string, targetPlatform string) string` 구현
- AC-21.3: `SkillList(cfg *config.HarnessConfig) []SkillMeta` 구현
- AC-21.4: 각 헬퍼에 단위 테스트 존재

**File Ownership:** `pkg/template/helpers.go`, `pkg/template/helpers_test.go`

### R22 — 공유 규칙 템플릿

> WHEN Codex/Gemini 규칙 생성 시,
> THE SYSTEM SHALL `content/rules/` 내 공유 규칙 소스를 플랫폼별 포맷으로 변환한다.

**Acceptance Criteria:**
- AC-22.1: `content/rules/` 디렉토리의 공통 규칙 소스를 각 플랫폼 포맷으로 변환
- AC-22.2: Codex는 Markdown 형식, Gemini는 Markdown + frontmatter 형식으로 렌더링
- AC-22.3: 규칙 내용의 본문이 플랫폼 간 동일

**File Ownership:** `content/rules/*.md`, `templates/codex/rules/*.md.tmpl`, `templates/gemini/rules/*.md.tmpl`

### R23 — 에이전트 메타데이터 공유

> WHEN Codex/Gemini 에이전트 생성 시,
> THE SYSTEM SHALL `content/agents/` 내 공유 에이전트 메타데이터를 플랫폼별 포맷으로 변환한다.

**Acceptance Criteria:**
- AC-23.1: 에이전트 name, description, core instructions가 플랫폼 간 일관성 유지
- AC-23.2: 플랫폼별 포맷 차이(TOML vs YAML frontmatter)만 분기

**File Ownership:** `content/agents/*.md`, `templates/codex/agents/*.toml.tmpl`, `templates/gemini/agents/*.md.tmpl`

### R24 — 테스트 유틸리티

> WHEN 어댑터 단위 테스트 작성 시,
> THE SYSTEM SHALL 공통 테스트 헬퍼(임시 디렉터리, 파일 검증, 매니페스트 비교)를 제공한다.

**Acceptance Criteria:**
- AC-24.1: `pkg/adapter/testutil_test.go`에 `setupTestDir()`, `assertFileExists()`, `assertFileContains()` 헬퍼 존재
- AC-24.2: Codex, Gemini 테스트에서 공통 헬퍼 사용

**File Ownership:** `pkg/adapter/testutil_test.go`

### R25 — 매니페스트 호환성

> WHEN 새 파일 타입(prompts, agents, hooks, settings)이 매니페스트에 추가될 때,
> THE SYSTEM SHALL 기존 매니페스트 스키마와 하위 호환성을 유지한다.

**Acceptance Criteria:**
- AC-25.1: 새 파일이 기존 `ManifestFromFiles()` 함수로 정상 등록됨
- AC-25.2: 이전 버전 매니페스트에서 `LoadManifest()` 호출 시 오류 없이 로드됨
- AC-25.3: 매니페스트 파일 수가 Codex +20, Gemini +20 증가

**File Ownership:** `pkg/adapter/manifest.go`, `pkg/adapter/manifest_test.go`

---

## Domain 4: Integration (R26-R30)

### R26 — auto init E2E 테스트 (3 플랫폼)

> WHEN `auto init --platform {codex|gemini|claude}` 실행 시,
> THE SYSTEM SHALL 각 플랫폼에 대해 생성 파일 수, 디렉터리 구조, 파일 내용을 검증하는 E2E 테스트를 통과한다.

**Acceptance Criteria:**
- AC-26.1: Codex init 후 `.codex/prompts/auto/`, `.codex/agents/`, `.codex/skills/`, `hooks.json` 존재 검증
- AC-26.2: Gemini init 후 `.gemini/commands/auto/`, `.gemini/agents/`, `.gemini/skills/autopus/`, `.gemini/rules/autopus/` 존재 검증
- AC-26.3: Claude init 후 기존 동작 무변경 (regression test)

**File Ownership:** `pkg/adapter/integration_test.go`

### R27 — auto update E2E 테스트

> WHEN `auto update` 실행 시,
> THE SYSTEM SHALL 사용자 커스터마이징이 보존되고 autopus 섹션만 업데이트됨을 검증하는 E2E 테스트를 통과한다.

**Acceptance Criteria:**
- AC-27.1: AGENTS.md 마커 외부 사용자 내용 보존 검증
- AC-27.2: GEMINI.md 마커 외부 사용자 내용 보존 검증
- AC-27.3: settings.json 사용자 키 보존 검증
- AC-27.4: hooks.json 사용자 훅 보존 검증

**File Ownership:** `pkg/adapter/integration_test.go`

### R28 — 매니페스트 파일 수 검증

> WHEN 어댑터가 Generate()를 완료할 때,
> THE SYSTEM SHALL 매니페스트의 파일 수가 Codex >= 26, Gemini >= 26임을 검증한다.

**Acceptance Criteria:**
- AC-28.1: Codex 매니페스트: AGENTS.md(1) + skills(10+) + prompts(6) + agents(5+) + hooks(1) + config(1) >= 24
- AC-28.2: Gemini 매니페스트: GEMINI.md(1) + skills(10+) + commands(6) + agents(5+) + rules(5+) + settings(1) >= 28

**File Ownership:** `pkg/adapter/codex/codex_test.go`, `pkg/adapter/gemini/gemini_test.go`

### R29 — 300줄 제한 준수 검증

> WHEN 새 Go 파일이 생성될 때,
> THE SYSTEM SHALL 이 SPEC에서 새로 생성하거나 수정한 Go 파일이 300줄 이하임을 검증한다.

**Acceptance Criteria:**
- AC-29.1: `pkg/adapter/codex/`에서 **신규 또는 수정된** 파일 300줄 이하
- AC-29.2: `pkg/adapter/gemini/`에서 **신규 또는 수정된** 파일 300줄 이하
- AC-29.3: 기존 300줄 초과 파일은 이 SPEC 범위 외 (별도 리팩토링 SPEC으로 처리)

**File Ownership:** CI pipeline, `Makefile`

### R30 — 매니페스트 JSON 스키마 업데이트

> WHEN 새 파일 타입이 추가될 때,
> THE SYSTEM SHALL `claude-code-manifest.json`에 새 어댑터 파일 정보를 반영한다.

**Acceptance Criteria:**
- AC-30.1: `.autopus/claude-code-manifest.json`에 codex, gemini 어댑터 파일 목록 업데이트
- AC-30.2: 매니페스트 버전 범프

**File Ownership:** `.autopus/claude-code-manifest.json`
