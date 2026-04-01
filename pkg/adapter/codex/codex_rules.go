package codex

import (
	"fmt"
	"io/fs"
	"strings"

	contentfs "github.com/insajin/autopus-adk/content"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/content"
	"github.com/insajin/autopus-adk/templates"
)

// renderRulesSection renders embedded rules as an inline section for AGENTS.md.
// Codex inlines rules directly in AGENTS.md since it has no separate rules directory.
func (a *Adapter) renderRulesSection(cfg *config.HarnessConfig) (string, error) {
	var sb strings.Builder
	sb.WriteString("\n## Rules\n\n")

	entries, err := contentfs.FS.ReadDir("rules")
	if err != nil {
		return "", fmt.Errorf("rules 디렉터리 읽기 실패: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// Skip branding rules (Codex doesn't use octopus branding natively).
		if entry.Name() == "branding.md" {
			continue
		}
		// Skip file-size-limit.md static copy — rendered from template below.
		if entry.Name() == "file-size-limit.md" {
			continue
		}

		data, err := fs.ReadFile(contentfs.FS, "rules/"+entry.Name())
		if err != nil {
			return "", fmt.Errorf("rule 파일 읽기 실패 %s: %w", entry.Name(), err)
		}
		// Strip YAML frontmatter if present.
		content := stripFrontmatter(string(data))
		sb.WriteString(content)
		sb.WriteString("\n")
	}

	// Render file-size-limit from template (stack/framework-aware).
	fileSizeContent, err := a.renderFileSizeRule(cfg)
	if err != nil {
		return "", err
	}
	sb.WriteString(fileSizeContent)
	sb.WriteString("\n")

	return sb.String(), nil
}

// fileSizeLimitData is the template data for the file-size-limit rule.
type fileSizeLimitData struct {
	Exclusions []content.FileSizeExclusion
}

// renderFileSizeRule renders the file-size-limit rule from template.
func (a *Adapter) renderFileSizeRule(cfg *config.HarnessConfig) (string, error) {
	tmplContent, err := templates.FS.ReadFile("claude/rules/file-size-limit.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("file-size-limit 템플릿 읽기 실패: %w", err)
	}
	exclusions := content.FileSizeExclusions(cfg.Stack, cfg.Framework)
	data := fileSizeLimitData{Exclusions: exclusions}
	rendered, err := a.engine.RenderString(string(tmplContent), data)
	if err != nil {
		return "", fmt.Errorf("file-size-limit 렌더링 실패: %w", err)
	}
	return rendered, nil
}

// stripFrontmatter removes YAML frontmatter (--- ... ---) from content.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return content
	}
	body := rest[idx+4:]
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}
	return body
}
