# Acceptance Criteria: SPEC-MULTIPLATFORM-001

**SPEC**: SPEC-MULTIPLATFORM-001 — Multi-Platform Harness Integration
**Status**: draft
**Created**: 2026-04-01

---

## 1. Unit Test Criteria — Codex Adapter

### UT-C1: Codex 커스텀 커맨드 생성

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestCodexGeneratePrompts` | `generatePrompts(cfg)` 호출 | `.codex/prompts/auto/` 하위에 plan.md, go.md, fix.md, review.md, sync.md, idea.md 6개 파일 생성 |
| `TestCodexPromptYAMLFrontmatter` | 생성된 prompt 파일 파싱 | 각 파일에 유효한 YAML front matter (`name`, `description`, `trigger` 키) 존재 |
| `TestCodexPromptContent` | auto-plan.md 내용 검증 | 스킬 라우팅 지시가 포함되어 있음 |

### UT-C2: Codex 에이전트 생성

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestCodexGenerateAgents` | `generateAgents(cfg)` 호출 | `.codex/agents/` 하위에 최소 5개 TOML 파일 생성 |
| `TestCodexAgentTOMLValidity` | 생성된 TOML 파일 파싱 | `name`, `description`, `developer_instructions`, `model` 필드 존재, 유효한 TOML |
| `TestCodexAgentConsistency` | 에이전트 이름 비교 | Claude `.claude/agents/autopus/*.md`의 에이전트와 1:1 매칭 |

### UT-C3: Codex 훅 생성

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestCodexGenerateHooks` | `generateHooks(cfg)` 호출 | `hooks.json` 파일 생성, 유효한 JSON |
| `TestCodexHookEvents` | hooks.json 파싱 | SessionStart, PreToolUse, PostToolUse, Stop 4개 이벤트 존재 |
| `TestCodexSupportsHooks` | `SupportsHooks()` 호출 | `true` 반환 |
| `TestCodexInstallHooks` | `InstallHooks(ctx, hooks, perms)` 호출 | 에러 없이 완료, hooks.json 업데이트됨 |

### UT-C4: Codex MCP/설정

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestCodexGenerateConfig` | `generateConfig(cfg)` 호출 | `config.toml`에 `[mcp_servers]` 섹션 존재 |
| `TestCodexPermissionMapping` | MapPermission 호출 | plan->"on-request", act->"auto", bypass->"never" |
| `TestCodexConfigMerge` | 기존 config.toml + 새 생성 | 사용자 기존 설정 보존, `[mcp_servers]`만 추가/업데이트 |

### UT-C5: Codex 규칙 인라인 (32KB)

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestCodexRulesInline` | `inlineRules(cfg)` 호출 | lore-commit, file-size-limit, subagent-delegation, language-policy 인라인됨 |
| `TestCodexRules32KBLimit` | AGENTS.md 마커 섹션 크기 | <= 32,768 bytes |
| `TestCodexRulesOverflow` | 규칙이 32KB 초과 시 | 초과분이 스킬 참조로 대체됨 |

### UT-C6: Codex 파일 분리

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestCodexFileLineCount` | `pkg/adapter/codex/*.go` 줄 수 | 모든 파일 300줄 이하 |
| `TestCodexExistingTests` | 기존 codex_test.go, codex_extra_test.go | 모든 기존 테스트 통과 (regression free) |

---

## 2. Unit Test Criteria — Gemini Adapter

### UT-G1: Gemini 커스텀 커맨드 생성

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestGeminiGenerateCommands` | `generateCommands(cfg)` 호출 | `.gemini/commands/auto/` 하위에 plan.toml, go.toml 등 6개 파일 생성 |
| `TestGeminiCommandTOMLValidity` | 생성된 TOML 파일 파싱 | `name`, `description`, `handler` 필드 존재, 유효한 TOML |
| `TestGeminiCommandNamespace` | 커맨드 namespace 검증 | `/auto:plan` 형식으로 호출 가능한 namespace 설정 |

### UT-G2: Gemini 에이전트 생성

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestGeminiGenerateAgents` | `generateAgents(cfg)` 호출 | `.gemini/agents/` 하위에 최소 5개 Markdown 파일 생성 |
| `TestGeminiAgentFrontmatter` | 생성된 MD 파일 파싱 | YAML frontmatter에 `name`, `description`, `model`, `tools` 필드 존재 |
| `TestGeminiAgentConsistency` | 에이전트 이름 비교 | Claude 에이전트와 1:1 매칭 |

### UT-G3: Gemini 규칙 + @import

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestGeminiGenerateRules` | `generateRules(cfg)` 호출 | `.gemini/rules/autopus/` 하위에 규칙 파일 생성 |
| `TestGeminiRulesImport` | GEMINI.md 내용 검증 | `@import .gemini/rules/autopus/lore-commit.md` 형식 참조 포함 |
| `TestGeminiRulesContent` | 규칙 파일 내용 비교 | Claude 규칙과 본문 동일 |

### UT-G4: Gemini 훅 + 설정

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestGeminiGenerateSettings` | `generateSettings(cfg)` 호출 | `settings.json` 파일 생성, 유효한 JSON |
| `TestGeminiHookEvents` | settings.json hooks 파싱 | BeforeAgent, AfterAgent, BeforeTool, AfterTool 이벤트 존재 |
| `TestGeminiMCPServers` | settings.json mcpServers 파싱 | Context7 서버 엔트리 포함 |
| `TestGeminiSupportsHooks` | `SupportsHooks()` 호출 | `true` 반환 |
| `TestGeminiPermissionMapping` | MapPermission 호출 | plan->"plan", act->"act", bypass->"yolo" |
| `TestGeminiSettingsMerge` | 기존 settings.json + 새 생성 | 사용자 기존 설정 보존, autopus 키만 추가/업데이트 |

### UT-G5: Gemini 파일 분리

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestGeminiFileLineCount` | `pkg/adapter/gemini/*.go` 줄 수 | 모든 파일 300줄 이하 |
| `TestGeminiExistingTests` | 기존 gemini_test.go, gemini_extra_test.go, gemini_hooks_test.go | 모든 기존 테스트 통과 |
| `TestGeminiOrchestraHooksPreserved` | orchestra 훅 기능 검증 | `InjectOrchestraAfterAgentHook` 등 기존 훅 로직 정상 동작 |

---

## 3. Template Rendering Validation

### TR-1: 렌더링 정확성

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestCodexTemplateRendering` | 모든 Codex 템플릿을 테스트 cfg로 렌더링 | 에러 없이 렌더링 완료, 플레이스홀더 잔존 없음 |
| `TestGeminiTemplateRendering` | 모든 Gemini 템플릿을 테스트 cfg로 렌더링 | 에러 없이 렌더링 완료, 플레이스홀더 잔존 없음 |
| `TestSharedTemplateRendering` | 공유 템플릿을 양 플랫폼으로 렌더링 | 본문 내용 동일, 포맷만 상이 |

### TR-2: 렌더링 성능

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `BenchmarkCodexGenerate` | Codex Generate() 벤치마크 | < 2초 |
| `BenchmarkGeminiGenerate` | Gemini Generate() 벤치마크 | < 2초 |

---

## 4. E2E Integration Tests

### E2E-1: auto init 전체 플랫폼

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestE2EInitCodex` | `auto init --platform codex` 시뮬레이션 | 아래 파일/디렉터리 존재: |
| | | - `AGENTS.md` (AUTOPUS 마커 포함) |
| | | - `.codex/prompts/auto/` (6개 .md) |
| | | - `.codex/agents/` (5+ .toml) |
| | | - `.codex/skills/` (10+ .md) |
| | | - `.codex/hooks.json` |
| | | - `config.toml` ([mcp_servers]) |
| `TestE2EInitGemini` | `auto init --platform gemini` 시뮬레이션 | 아래 파일/디렉터리 존재: |
| | | - `GEMINI.md` (AUTOPUS 마커 + @import) |
| | | - `.gemini/commands/auto/` (6개 .toml) |
| | | - `.gemini/agents/` (5+ .md) |
| | | - `.gemini/skills/autopus/` (10+ dirs) |
| | | - `.gemini/rules/autopus/` (4+ .md) |
| | | - `.gemini/settings.json` (hooks + mcpServers) |
| `TestE2EInitClaude` | `auto init --platform claude` 시뮬레이션 | 기존 동작과 동일 (regression test) |

### E2E-2: auto update 보존성

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestE2EUpdateCodexPreserve` | AGENTS.md에 사용자 커스텀 추가 후 update | 마커 외부 커스텀 내용 보존, 마커 내부만 업데이트 |
| `TestE2EUpdateGeminiPreserve` | GEMINI.md + settings.json에 사용자 설정 추가 후 update | 마커 외부 보존, 사용자 settings.json 키 보존 |
| `TestE2EUpdateIdempotent` | 동일 설정으로 2번 연속 update | 파일 내용 동일 (멱등성) |

### E2E-3: 파일 수 검증

| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| `TestE2ECodexFileCount` | Codex 매니페스트 파일 수 | >= 24 files |
| `TestE2EGeminiFileCount` | Gemini 매니페스트 파일 수 | >= 28 files |

---

## 5. 300-Line Limit Compliance

### LC-1: 소스 코드 줄 수 검증

| Scope | Validation |
|-------|-----------|
| `pkg/adapter/codex/*.go` | 각 파일 `wc -l` <= 300 |
| `pkg/adapter/gemini/*.go` | 각 파일 `wc -l` <= 300 |
| `pkg/template/helpers.go` | `wc -l` <= 300 |
| `pkg/adapter/testutil_test.go` | `wc -l` <= 300 |

### LC-2: 자동 검증 스크립트

```bash
# CI에서 실행할 검증 명령
find pkg/ -name "*.go" -exec sh -c 'lines=$(wc -l < "$1"); if [ "$lines" -gt 300 ]; then echo "FAIL: $1 ($lines lines)"; exit 1; fi' _ {} \;
```

---

## 6. Cross-Platform Consistency

### CC-1: 기능 패리티 매트릭스

| Feature | Claude | Codex | Gemini | Parity Required |
|---------|--------|-------|--------|----------------|
| Custom Commands | SKILL.md | .codex/prompts/auto/*.md | .gemini/commands/auto/*.toml | YES |
| Agent Definitions | .claude/agents/*.md | .codex/agents/*.toml | .gemini/agents/*.md | YES |
| Rules | .claude/rules/*.md | AGENTS.md inline + skills | .gemini/rules/*.md + @import | YES |
| Skills | .claude/skills/ | .codex/skills/ | .gemini/skills/autopus/ | YES |
| Hooks | settings.json | hooks.json | settings.json | YES |
| MCP Config | settings.json | config.toml | settings.json | YES |
| Permission Modes | plan/act/bypass | on-request/auto/never | plan/act/yolo | YES |
| Subagent Pattern | Agent() | spawn_agent | @agent | YES |
| Marker Update | CLAUDE.md marker | AGENTS.md marker | GEMINI.md marker | YES |

### CC-2: 검증 방법

| Test | Method |
|------|--------|
| 커맨드 수 동일 | Claude /auto 서브커맨드 수 == Codex prompts 수 == Gemini commands 수 |
| 에이전트 수 동일 | 3개 플랫폼 에이전트 이름 집합 비교 |
| 스킬 커버리지 | Claude 스킬 목록과 Codex/Gemini 스킬 목록 비교, >= 90% 매칭 |
| 규칙 내용 동일 | 공유 규칙 본문의 diff가 포맷 차이만 있음 (내용 동일) |

---

## 7. Acceptance Checklist (최종 검증)

- [ ] `go test ./pkg/adapter/codex/... -cover` >= 85%
- [ ] `go test ./pkg/adapter/gemini/... -cover` >= 85%
- [ ] `go test ./pkg/template/... -cover` >= 85%
- [ ] `go test ./pkg/adapter/... -run TestE2E` 전체 통과
- [ ] `pkg/adapter/codex/*.go` 모든 파일 300줄 이하
- [ ] `pkg/adapter/gemini/*.go` 모든 파일 300줄 이하
- [ ] Codex 매니페스트 파일 수 >= 24
- [ ] Gemini 매니페스트 파일 수 >= 28
- [ ] Codex `SupportsHooks()` == true
- [ ] Gemini `SupportsHooks()` == true
- [ ] Claude 기존 테스트 regression 없음
- [ ] 기존 `codex_test.go`, `codex_extra_test.go` 통과
- [ ] 기존 `gemini_test.go`, `gemini_extra_test.go`, `gemini_hooks_test.go` 통과
- [ ] AGENTS.md 마커 섹션 32KB 이하
- [ ] 템플릿 렌더링 < 2초 (벤치마크)
- [ ] `auto update` 멱등성 검증 통과
