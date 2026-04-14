package codex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/templates"
)

// renderSkillTemplates reads Codex skill templates from embedded FS,
// renders them, and writes to .codex/skills/.
func (a *Adapter) renderSkillTemplates(cfg *config.HarnessConfig) ([]adapter.FileMapping, error) {
	var files []adapter.FileMapping

	entries, err := templates.FS.ReadDir("codex/skills")
	if err != nil {
		return nil, fmt.Errorf("코덱스 스킬 템플릿 디렉터리 읽기 실패: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}

		name := entry.Name()
		skillFile := strings.TrimSuffix(name, ".tmpl")

		tmplContent, err := templates.FS.ReadFile("codex/skills/" + name)
		if err != nil {
			return nil, fmt.Errorf("코덱스 스킬 템플릿 읽기 실패 %s: %w", name, err)
		}

		rendered, err := a.engine.RenderString(string(tmplContent), cfg)
		if err != nil {
			return nil, fmt.Errorf("코덱스 스킬 템플릿 렌더링 실패 %s: %w", name, err)
		}
		rendered = normalizeCodexInvocationBody(rendered)
		rendered = normalizeCodexHelperPaths(rendered)
		rendered = normalizeCodexToolingBody(rendered)

		targetPath := filepath.Join(a.root, ".codex", "skills", skillFile)
		if err := os.WriteFile(targetPath, []byte(rendered), 0644); err != nil {
			return nil, fmt.Errorf("코덱스 스킬 파일 쓰기 실패 %s: %w", targetPath, err)
		}

		files = append(files, adapter.FileMapping{
			TargetPath:      filepath.Join(".codex", "skills", skillFile),
			OverwritePolicy: adapter.OverwriteAlways,
			Checksum:        checksum(rendered),
			Content:         []byte(rendered),
		})
	}

	// Extended skills from content/skills/ via transformer
	extFiles, err := a.renderExtendedSkills()
	if err != nil {
		return nil, fmt.Errorf("extended skill rendering failed: %w", err)
	}
	for _, ef := range extFiles {
		targetPath := filepath.Join(a.root, ef.TargetPath)
		if err := os.WriteFile(targetPath, ef.Content, 0644); err != nil {
			return nil, fmt.Errorf("extended skill write failed %s: %w", targetPath, err)
		}
	}
	files = append(files, extFiles...)

	return files, nil
}

// agentsMDTemplate is the AGENTS.md AUTOPUS section template.
// Kept slim — detailed rules and agent definitions live in separate files.
const agentsMDTemplate = `# Autopus-ADK Harness

> 이 섹션은 Autopus-ADK에 의해 자동 생성됩니다. 수동으로 편집하지 마세요.

- **프로젝트**: {{.ProjectName}}
- **모드**: {{.Mode}}
- **플랫폼**: {{join ", " .Platforms}}

## Installed Components

{{if contains (join ", " .Platforms) "codex"}}- Codex Rules: .codex/rules/autopus/
- Codex Skills: .codex/skills/
- Codex Agents: .codex/agents/
- Codex Config: config.toml
{{end}}{{if contains (join ", " .Platforms) "opencode"}}- OpenCode Rules: .opencode/rules/autopus/
- OpenCode Commands: .opencode/commands/
- OpenCode Agents: .opencode/agents/
- OpenCode Plugins: .opencode/plugins/
{{end}}{{if contains (join ", " .Platforms) "codex"}}- Shared Agent Skills: .agents/skills/
- Plugin Marketplace: .agents/plugins/marketplace.json
{{else if contains (join ", " .Platforms) "opencode"}}- Shared Skills: .agents/skills/
{{end}}

## Language Policy

IMPORTANT: Follow these language settings strictly for all work in this project.

- **Code comments**: {{.Language.Comments}}
- **Commit messages**: {{.Language.Commits}}
- **AI responses**: {{.Language.AIResponses}}

## Execution Model

{{if contains (join ", " .Platforms) "codex"}}- **Codex**: 하네스 기본값은 spawn_agent(...) 기반 subagent-first 입니다.
- **Codex --auto**: @auto ... --auto 가 포함되면, 기본 subagent pipeline 진행에 대한 명시적 승인으로 해석합니다.
- **Codex Runtime Caveat**: 현재 세션의 Codex 런타임 정책이 암묵적 spawn_agent(...) 호출을 제한하면, 조용히 단일 세션으로 폴백하지 말고 그 제약을 명시적으로 알린 뒤 사용자의 서브에이전트 opt-in 또는 --solo 선택을 받으세요.
- **Codex --team**: 미래의 native multi-agent surface를 위한 reserved compatibility flag입니다.
{{end}}{{if contains (join ", " .Platforms) "opencode"}}- **OpenCode**: 기본 실행 모델은 task(...) 기반 subagent-first 입니다.
- **OpenCode Invocation**: /auto <subcommand> ... 또는 /auto-<subcommand> ... alias를 사용합니다.
{{end}}

## Core Guidelines

{{if contains (join ", " .Platforms) "codex"}}See .codex/rules/autopus/ for Codex rule definitions.
See .codex/agents/ for Codex agent definitions.
{{end}}{{if contains (join ", " .Platforms) "opencode"}}See .opencode/rules/autopus/ for OpenCode rule definitions.
{{end}}
`
