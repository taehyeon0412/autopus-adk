package codex

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/insajin/autopus-adk/pkg/config"
)

const (
	markerBegin = "<!-- AUTOPUS:BEGIN -->"
	markerEnd   = "<!-- AUTOPUS:END -->"
)

var markerRe = regexp.MustCompile(`(?s)` + regexp.QuoteMeta(markerBegin) + `.*?` + regexp.QuoteMeta(markerEnd))

// injectMarkerSection creates or updates the AUTOPUS marker section in AGENTS.md.
func (a *Adapter) injectMarkerSection(cfg *config.HarnessConfig) (string, error) {
	agentsPath := filepath.Join(a.root, "AGENTS.md")
	cfg = ensurePlatformInRootDoc(cfg, "codex")

	var existing string
	if data, err := os.ReadFile(agentsPath); err == nil {
		existing = string(data)
	}

	sectionContent, err := a.engine.RenderString(agentsMDTemplate, cfg)
	if err != nil {
		return "", fmt.Errorf("AGENTS.md 템플릿 렌더링 실패: %w", err)
	}

	// Append inline agents section.
	agentsSection, err := renderAgentsSection()
	if err != nil {
		return "", fmt.Errorf("agents 섹션 렌더링 실패: %w", err)
	}
	sectionContent += agentsSection

	// Reference separate rule files instead of inlining.
	sectionContent += "\n## Rules\n\n"
	if containsPlatform(cfg.Platforms, "codex") {
		sectionContent += "See .codex/rules/autopus/ for Codex guidance.\n"
	}
	if containsPlatform(cfg.Platforms, "opencode") {
		sectionContent += "See .opencode/rules/autopus/ for OpenCode guidance.\n"
	}

	newSection := markerBegin + "\n" + sectionContent + "\n" + markerEnd

	if strings.Contains(existing, markerBegin) && strings.Contains(existing, markerEnd) {
		return replaceMarkerSection(existing, newSection), nil
	}

	if existing == "" {
		return newSection + "\n", nil
	}
	return existing + "\n\n" + newSection + "\n", nil
}

func replaceMarkerSection(content, newSection string) string {
	return markerRe.ReplaceAllString(content, newSection)
}

func removeMarkerSection(content string) string {
	return strings.TrimSpace(markerRe.ReplaceAllString(content, "")) + "\n"
}

func containsPlatform(platforms []string, target string) bool {
	for _, platform := range platforms {
		if platform == target {
			return true
		}
	}
	return false
}

func ensurePlatformInRootDoc(cfg *config.HarnessConfig, platform string) *config.HarnessConfig {
	if cfg == nil {
		return nil
	}
	if containsPlatform(cfg.Platforms, platform) {
		return cfg
	}
	cloned := *cfg
	cloned.Platforms = append(append([]string{}, cfg.Platforms...), platform)
	return &cloned
}
