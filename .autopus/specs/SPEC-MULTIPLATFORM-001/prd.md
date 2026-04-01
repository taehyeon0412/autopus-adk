# PRD: Multi-Platform Harness Integration — Codex CLI + Gemini CLI Harness Parity

> Product Requirements Document — Standard (11-section format).

- **SPEC-ID**: SPEC-MULTIPLATFORM-001
- **Author**: Autopus Spec-Writer Agent
- **Status**: Draft
- **Date**: 2026-04-01
- **Target Module**: autopus-adk

---

## 1. Problem & Context

현재 autopus-adk의 하네스는 Claude Code에서만 완전하게 작동한다.
Codex CLI와 Gemini CLI 어댑터는 각각 **6개의 기본 스킬만** 생성하며, 다음 기능이 완전히 누락되어 있다:

| 누락 기능 | Codex 현황 | Gemini 현황 | Claude 현황 |
|-----------|-----------|------------|------------|
| Custom Commands (/auto routing) | 없음 | 없음 | `.claude/skills/auto/SKILL.md` |
| Agent Definitions | 없음 | 없음 | `.claude/agents/autopus/*.md` |
| Rules (branding, lore-commit 등) | AGENTS.md 인라인 일부 | GEMINI.md 인라인 일부 | `.claude/rules/autopus/*.md` |
| Hooks (PreToolUse, PostToolUse) | `SupportsHooks()=false` | `SupportsHooks()=false` | settings.json hooks |
| MCP Config | 없음 | 없음 | `.claude/settings.json` mcpServers |
| Permission Modes | 없음 | 없음 | plan/act/bypass |

**현재 코드 구조 분석:**

- `pkg/adapter/codex/codex.go` (399줄): AGENTS.md 마커 섹션 + `.codex/skills/` 6개 스킬만 생성
- `pkg/adapter/gemini/gemini.go` (424줄): GEMINI.md 마커 섹션 + `.gemini/skills/autopus/` 6개 스킬만 생성
- `templates/codex/skills/`: 6개 `.md.tmpl` 파일 (auto-plan, auto-go, auto-fix, auto-sync, auto-review, auto-idea)
- `templates/gemini/skills/`: 6개 디렉터리, 각각 `SKILL.md.tmpl`

2026년 3월 기준 Codex CLI와 Gemini CLI 모두 네이티브로 스킬, 서브에이전트, 워크트리, 커스텀 커맨드, 훅을 지원하게 되어 **하네스 완전 패리티가 달성 가능**하다.

### 근본 원인

1. **초기 어댑터 MVP**: Codex/Gemini 어댑터는 스킬 생성만을 목적으로 구현됨
2. **플랫폼 미성숙**: 구현 시점에 Codex/Gemini의 훅, 커맨드, 에이전트 기능이 미완성이었음
3. **템플릿 부재**: Claude용 templates/claude/ 에 있는 rules, commands, mcp 템플릿의 Codex/Gemini 대응이 없음

---

## 2. Goals & Success Metrics

| Goal | Metric | Target |
|------|--------|--------|
| 플랫폼 패리티 | Claude 대비 기능 커버리지 % | Codex >= 90%, Gemini >= 90% |
| 어댑터 완성도 | 생성 파일 수 | Codex: +20 files, Gemini: +20 files |
| 테스트 커버리지 | `go test -cover` on new/modified files | 85%+ |
| auto init/update E2E | 3개 플랫폼 시나리오 통과 | All pass |
| 파일 크기 제한 | 모든 Go 파일 300줄 이하 | 100% compliance |
| 템플릿 렌더링 성능 | 전체 파일 생성 시간 | < 2초 |

---

## 3. Target Users

1. **Codex CLI 사용자**: OpenAI Codex를 주요 코딩 에이전트로 사용하는 개발자. `auto init --platform codex`로 완전한 하네스를 기대.
2. **Gemini CLI 사용자**: Google Gemini CLI를 주요 코딩 에이전트로 사용하는 개발자. GEMINI.md + @import rules + /auto 커스텀 커맨드를 기대.
3. **멀티 프로바이더 사용자**: Orchestra의 brainstorm/review에서 3개 플랫폼을 모두 사용하는 팀. 모든 플랫폼에 동일한 컨텍스트(rules, agents, skills)가 구성되어야 한다.

---

## 4. User Stories

### US-1: Codex CLI 완전 초기화
> As a Codex CLI user,
> I want to run `auto init --platform codex` and get a fully configured harness with skills, agents, custom commands, hooks, and MCP config,
> so that I can use /auto workflows (plan, go, fix, review, sync) with the same quality as Claude Code users.

### US-2: Gemini CLI 완전 초기화
> As a Gemini CLI user,
> I want to run `auto init --platform gemini` and get GEMINI.md with @import rules, /auto:plan custom commands, agent definitions, hooks, and MCP config,
> so that I can use the full Autopus development workflow.

### US-3: 멀티 프로바이더 Orchestra
> As a multi-provider user running orchestra brainstorm/review,
> I want all 3 platforms (Claude, Codex, Gemini) configured with equivalent rules, agents, and skills,
> so that each provider in the orchestra has full context and can produce consistent results.

### US-4: 하네스 업데이트 안전성
> As a developer who has customized my AGENTS.md or GEMINI.md,
> I want `auto update` to update only the AUTOPUS:BEGIN/END marker sections,
> so that my custom instructions outside the markers are preserved.

### US-5: 커스텀 커맨드 사용
> As a Codex CLI user,
> I want to type `/auto plan` in Codex and have it routed to the correct skill,
> so that I don't need to remember different command patterns per platform.

---

## 5. Functional Requirements (EARS format, MoSCoW)

### P0 -- Must Have

| ID | Requirement |
|----|-------------|
| FR-01 | WHEN `auto init --platform codex` 실행 시, THE SYSTEM SHALL `.codex/prompts/`, `.codex/agents/`, `.codex/skills/`, `hooks.json`, `AGENTS.md`를 생성한다 |
| FR-02 | WHEN `auto init --platform gemini` 실행 시, THE SYSTEM SHALL `.gemini/commands/auto/`, `.gemini/agents/`, `.gemini/skills/`, `settings.json` hooks, `GEMINI.md`를 생성한다 |
| FR-03 | WHEN Codex 커스텀 커맨드 생성 시, THE SYSTEM SHALL `/auto` 서브커맨드당 하나의 `.md` 파일을 `.codex/prompts/`에 YAML front matter와 함께 생성한다 |
| FR-04 | WHEN Gemini 커스텀 커맨드 생성 시, THE SYSTEM SHALL `/auto` 서브커맨드당 하나의 `.toml` 파일을 `.gemini/commands/auto/`에 namespace 지원으로 생성한다 |
| FR-05 | WHEN Codex 에이전트 생성 시, THE SYSTEM SHALL `.codex/agents/`에 name, description, developer_instructions, model 필드를 포함한 TOML 파일을 생성한다 |
| FR-06 | WHEN Gemini 에이전트 생성 시, THE SYSTEM SHALL `.gemini/agents/`에 YAML frontmatter를 포함한 Markdown 파일을 생성한다 |
| FR-07 | WHEN 스킬 생성 시, THE SYSTEM SHALL 모든 Claude 스킬을 플랫폼 네이티브 포맷으로 변환한다 (Codex: `.codex/skills/SKILL.md`, Gemini: `.gemini/skills/autopus/{name}/SKILL.md`) |
| FR-08 | WHEN Codex 훅 생성 시, THE SYSTEM SHALL `hooks.json`에 SessionStart, PreToolUse, PostToolUse, Stop 이벤트를 생성한다 |
| FR-09 | WHEN Gemini 훅 생성 시, THE SYSTEM SHALL `settings.json`에 BeforeAgent, AfterAgent, BeforeTool, AfterTool 이벤트를 추가한다 |
| FR-10 | WHEN Codex MCP 설정 생성 시, THE SYSTEM SHALL `config.toml`에 `[mcp_servers]` 섹션을 추가한다 |
| FR-11 | WHEN Gemini MCP 설정 생성 시, THE SYSTEM SHALL `settings.json`에 `mcpServers`를 추가한다 |
| FR-12 | WHEN `auto update`가 어떤 플랫폼에서든 실행될 때, THE SYSTEM SHALL AUTOPUS:BEGIN/END 마커 또는 merge 전략을 사용하여 사용자 커스터마이징을 보존하면서 기존 파일을 업데이트한다 |

### P1 -- Should Have

| ID | Requirement |
|----|-------------|
| FR-13 | WHEN Codex AGENTS.md 생성 시, THE SYSTEM SHALL 핵심 규칙을 32KB 제한 내에서 인라인하고, 초과 규칙은 스킬 참조로 분리한다 |
| FR-14 | WHEN GEMINI.md 생성 시, THE SYSTEM SHALL `@import` 구문을 사용하여 `.gemini/rules/autopus/` 내 규칙 파일을 참조한다 |
| FR-15 | WHEN 파이프라인 스킬(auto-go) 변환 시, THE SYSTEM SHALL Claude의 Agent() 호출을 플랫폼 네이티브 서브에이전트 패턴으로 매핑한다 (Codex: spawn_agent, Gemini: @agent tool) |
| FR-16 | WHEN 권한 모드 변환 시, THE SYSTEM SHALL Claude plan -> Codex on-request / Gemini plan, Claude bypass -> Codex never / Gemini yolo로 매핑한다 |

### P2 -- Could Have

| ID | Requirement |
|----|-------------|
| FR-17 | THE SYSTEM COULD 플랫폼별 README.md에 셋업 지시를 생성한다 |
| FR-18 | THE SYSTEM COULD 생성 파일을 플랫폼 스키마에 대해 검증한다 (Codex TOML validity, Gemini YAML frontmatter) |

---

## 6. Non-Functional Requirements

| NFR | Description | Threshold |
|-----|-------------|-----------|
| NFR-01 | 모든 생성된 Go 파일은 300줄 이하 | Hard limit |
| NFR-02 | 템플릿 렌더링 전체 완료 시간 | < 2초 |
| NFR-03 | go.mod 기존 의존성 외 추가 금지 | 0 new deps |
| NFR-04 | 기존 Claude 어댑터 동작 무변경 | Zero regression |
| NFR-05 | 기존 codex.go, gemini.go 300줄 초과 방지 | 관심사별 파일 분리 |

---

## 7. Technical Constraints

- **Go 1.26**, 기존 cobra CLI 프레임워크 사용
- **템플릿 엔진**: `pkg/template/` (Handlebars-style), `templates/` embed.go를 통한 임베디드 FS
- **어댑터 인터페이스**: `pkg/adapter/adapter.go` (PlatformAdapter, 10 methods)
- **파일 덮어쓰기 정책**: `OverwriteAlways`, `OverwriteNever`, `OverwriteMarker`, `OverwriteMerge`
- **기존 파일 크기 이슈**: `codex.go` 399줄, `gemini.go` 424줄 — 이미 300줄 제한 초과이므로 이번 작업에서 분리 필수
- **매니페스트 시스템**: `adapter.ManifestFromFiles()`, `adapter.LoadManifest()` — update 시 변경 감지

---

## 8. Out of Scope

- Orchestra 엔진 변경 (이미 3 프로바이더 지원)
- Bridge 프로바이더 변경
- Codex/Gemini 외 새 플랫폼 추가
- Claude Code 어댑터 수정
- 플랫폼별 IDE 익스텐션
- `pkg/adapter/opencode/` 어댑터 변경

---

## 9. Risks & Open Questions

| Risk | Severity | Mitigation |
|------|----------|------------|
| Codex hooks.json이 experimental | Medium | Feature-gated, `SupportsHooks()` true 전환 시 graceful degradation |
| Codex AGENTS.md 32KB 제한 | Medium | 핵심 규칙만 인라인, 나머지는 `.codex/skills/` 스킬 참조 |
| Gemini --worktree가 preview | Low | Shell-based worktree fallback을 스킬에 포함 |
| Codex skills 기능이 feature flag 뒤 | Medium | README에 `--enable skills` 요구사항 문서화 |
| 기존 codex.go/gemini.go 300줄 초과 | High | 이번 작업에서 관심사별 파일 분리 (prompts, agents, hooks, settings) |

### Open Questions

1. **Codex TOML agent format 확정**: Codex 2026.3의 공식 에이전트 TOML 스키마가 안정화되었는가?
2. **Gemini @import 깊이 제한**: GEMINI.md에서 @import가 몇 단계까지 지원되는가?
3. **hooks.json vs config.toml 훅**: Codex 훅이 hooks.json 단독인지 config.toml 통합인지?

---

## 10. Pre-mortem

- **"6개월 후 실패 — 템플릿 전량 재작성"**: Codex/Gemini가 breaking change를 발표하면 모든 템플릿 재작성이 필요하다. **미리 템플릿을 최소 단위로 모듈화**하여 변경 범위를 제한한다. 각 플랫폼 기능을 독립 템플릿 파일로 분리하여 부분 수정이 가능하게 한다.
- **"스킬이 작동 안 함"**: 플랫폼별 스킬 로딩 메커니즘이 예상과 다르면 전체 워크플로우가 깨진다. **각 플랫폼 E2E 테스트를 반드시 작성**하여 실제 CLI에서 스킬 로딩을 검증한다.
- **"32KB 초과"**: AGENTS.md에 모든 규칙을 인라인할 수 없으면 Codex 사용자의 경험이 저하된다. **핵심 규칙(lore-commit, file-size-limit, subagent-delegation)만 인라인**하고 나머지(branding, context7-docs 등)는 스킬 파일로 분리하는 2단계 전략을 사용한다.
- **"Go 파일 300줄 초과"**: 기능 추가로 codex.go/gemini.go가 더 커지면 리뷰 불통과. **초기부터 관심사별 분리**: `codex_prompts.go`, `codex_agents.go`, `codex_hooks.go`, `codex_settings.go` 등으로 나눈다.

---

## 11. Practitioner Q&A

**Q: 스킬 72+개를 어떻게 변환하나?**
A: 공통 로직은 `templates/shared/` 템플릿으로 통합하고, 플랫폼별 차이만 `templates/codex/`, `templates/gemini/` 템플릿에서 처리한다. 스킬 메타데이터(name, description, trigger)는 공유하고 포맷(Markdown vs TOML)만 분기한다.

**Q: 커스텀 커맨드와 스킬의 경계는?**
A: 커맨드 = 사용자 진입점(`/auto plan`), 스킬 = 상세 지시 내용(`planning.md` 본문). 커맨드는 라우팅만 하고, 실제 동작은 스킬 파일에 정의된다.

**Q: 기존 codex.go 399줄 + 새 기능 = 600줄+?**
A: 반드시 분리한다. 현재 `codex.go`의 `injectMarkerSection`, `renderSkillTemplates`, `prepareFiles`를 별도 파일로 추출하고, 새 기능(prompts, agents, hooks, settings)도 각각 별도 파일로 생성한다. 최종 파일당 200줄 이내를 목표로 한다.

**Q: `SupportsHooks()` 변경 시 기존 동작 영향은?**
A: Codex의 `SupportsHooks()` = false -> true 전환은 `InstallHooks()`가 실제 동작하게 바뀌므로, hooks.json 생성 로직을 `InstallHooks()` 안에 구현한다. Generate()에서는 hooks.json 파일을 직접 생성하고, InstallHooks()는 권한 설정을 담당한다.

**Q: Gemini settings.json 머지 전략은?**
A: `OverwriteMerge` 정책을 사용한다. 기존 settings.json을 JSON 파싱 -> autopus 관련 키만 업데이트 -> 나머지 키 보존 -> 재직렬화. 이미 `adapter.OverwriteMerge` 정책이 정의되어 있으므로 해당 로직을 구현한다.
