package codex

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/content"
	"github.com/insajin/autopus-adk/templates"
)

// generateHooks renders hooks.json template and writes to .codex/hooks.json.
func (a *Adapter) generateHooks(cfg *config.HarnessConfig) ([]adapter.FileMapping, error) {
	tmplContent, err := templates.FS.ReadFile("codex/hooks.json.tmpl")
	if err != nil {
		return nil, fmt.Errorf("codex hooks 템플릿 읽기 실패: %w", err)
	}

	rendered, err := a.engine.RenderString(string(tmplContent), cfg)
	if err != nil {
		return nil, fmt.Errorf("codex hooks 템플릿 렌더링 실패: %w", err)
	}

	targetPath := filepath.Join(a.root, ".codex", "hooks.json")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return nil, fmt.Errorf(".codex 디렉터리 생성 실패: %w", err)
	}
	if err := os.WriteFile(targetPath, []byte(rendered), 0644); err != nil {
		return nil, fmt.Errorf("codex hooks.json 쓰기 실패: %w", err)
	}

	return []adapter.FileMapping{{
		TargetPath:      filepath.Join(".codex", "hooks.json"),
		OverwritePolicy: adapter.OverwriteAlways,
		Checksum:        checksum(rendered),
		Content:         []byte(rendered),
	}}, nil
}

// prepareHooksFile returns hooks.json file mapping without writing to disk.
func (a *Adapter) prepareHooksFile(cfg *config.HarnessConfig) ([]adapter.FileMapping, error) {
	tmplContent, err := templates.FS.ReadFile("codex/hooks.json.tmpl")
	if err != nil {
		return nil, fmt.Errorf("codex hooks 템플릿 읽기 실패: %w", err)
	}

	rendered, err := a.engine.RenderString(string(tmplContent), cfg)
	if err != nil {
		return nil, fmt.Errorf("codex hooks 템플릿 렌더링 실패: %w", err)
	}

	return []adapter.FileMapping{{
		TargetPath:      filepath.Join(".codex", "hooks.json"),
		OverwritePolicy: adapter.OverwriteAlways,
		Checksum:        checksum(rendered),
		Content:         []byte(rendered),
	}}, nil
}

// installGitHooks generates and writes git hooks as fallback.
func (a *Adapter) installGitHooks(cfg *config.HarnessConfig) error {
	_, gitHooks, _ := content.GenerateHookConfigs(cfg.Hooks, adapterName, false)

	for _, gh := range gitHooks {
		ghPath := filepath.Join(a.root, gh.Path)
		if err := os.MkdirAll(filepath.Dir(ghPath), 0755); err != nil {
			return fmt.Errorf("git hook 디렉터리 생성 실패: %w", err)
		}
		if err := os.WriteFile(ghPath, []byte(gh.Content), 0755); err != nil {
			return fmt.Errorf("git hook 쓰기 실패 %s: %w", gh.Path, err)
		}
	}
	return nil
}
