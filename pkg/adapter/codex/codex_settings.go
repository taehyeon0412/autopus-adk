package codex

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/templates"
)

// generateConfig renders config.toml template and writes to project root.
func (a *Adapter) generateConfig(cfg *config.HarnessConfig) ([]adapter.FileMapping, error) {
	tmplContent, err := templates.FS.ReadFile("codex/config.toml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("codex config 템플릿 읽기 실패: %w", err)
	}

	rendered, err := a.engine.RenderString(string(tmplContent), cfg)
	if err != nil {
		return nil, fmt.Errorf("codex config 템플릿 렌더링 실패: %w", err)
	}

	targetPath := filepath.Join(a.root, "config.toml")
	if err := os.WriteFile(targetPath, []byte(rendered), 0644); err != nil {
		return nil, fmt.Errorf("codex config.toml 쓰기 실패: %w", err)
	}

	return []adapter.FileMapping{{
		TargetPath:      "config.toml",
		OverwritePolicy: adapter.OverwriteMerge,
		Checksum:        checksum(rendered),
		Content:         []byte(rendered),
	}}, nil
}

// prepareConfigFile returns config.toml file mapping without writing to disk.
func (a *Adapter) prepareConfigFile(cfg *config.HarnessConfig) ([]adapter.FileMapping, error) {
	tmplContent, err := templates.FS.ReadFile("codex/config.toml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("codex config 템플릿 읽기 실패: %w", err)
	}

	rendered, err := a.engine.RenderString(string(tmplContent), cfg)
	if err != nil {
		return nil, fmt.Errorf("codex config 템플릿 렌더링 실패: %w", err)
	}

	return []adapter.FileMapping{{
		TargetPath:      "config.toml",
		OverwritePolicy: adapter.OverwriteMerge,
		Checksum:        checksum(rendered),
		Content:         []byte(rendered),
	}}, nil
}
