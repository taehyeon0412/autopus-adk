// Package gemini provides rule template rendering for Gemini CLI.
package gemini

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/content"
	"github.com/insajin/autopus-adk/templates"
)

const rulesTemplateDir = "gemini/rules/autopus"

// fileSizeLimitData is the template data for file-size-limit.md.
type fileSizeLimitData struct {
	Exclusions []content.FileSizeExclusion
}

// renderRuleTemplates reads Gemini rule templates from embedded FS,
// renders them, and writes to .gemini/rules/autopus/.
func (a *Adapter) renderRuleTemplates(cfg *config.HarnessConfig) ([]adapter.FileMapping, error) {
	mappings, err := a.prepareRuleMappings(cfg)
	if err != nil {
		return nil, err
	}

	for _, m := range mappings {
		destPath := filepath.Join(a.root, m.TargetPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, fmt.Errorf("gemini rules 디렉터리 생성 실패: %w", err)
		}
		if err := os.WriteFile(destPath, m.Content, 0644); err != nil {
			return nil, fmt.Errorf("gemini rule 파일 쓰기 실패 %s: %w", destPath, err)
		}
	}

	return mappings, nil
}

// prepareRuleMappings renders rule templates and returns file mappings
// without writing to disk.
func (a *Adapter) prepareRuleMappings(cfg *config.HarnessConfig) ([]adapter.FileMapping, error) {
	var files []adapter.FileMapping

	entries, err := templates.FS.ReadDir(rulesTemplateDir)
	if err != nil {
		return nil, fmt.Errorf("gemini rule 템플릿 디렉터리 읽기 실패: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}

		name := entry.Name()
		outFile := strings.TrimSuffix(name, ".tmpl")

		tmplContent, err := templates.FS.ReadFile(rulesTemplateDir + "/" + name)
		if err != nil {
			return nil, fmt.Errorf("gemini rule 템플릿 읽기 실패 %s: %w", name, err)
		}

		// file-size-limit uses a special data struct with exclusions
		var rendered string
		if outFile == "file-size-limit.md" {
			exclusions := content.FileSizeExclusions(cfg.Stack, cfg.Framework)
			data := fileSizeLimitData{Exclusions: exclusions}
			rendered, err = a.engine.RenderString(string(tmplContent), data)
		} else {
			rendered, err = a.engine.RenderString(string(tmplContent), cfg)
		}
		if err != nil {
			return nil, fmt.Errorf("gemini rule 템플릿 렌더링 실패 %s: %w", name, err)
		}

		relPath := filepath.Join(".gemini", "rules", "autopus", outFile)
		files = append(files, adapter.FileMapping{
			TargetPath:      relPath,
			OverwritePolicy: adapter.OverwriteAlways,
			Checksum:        checksum(rendered),
			Content:         []byte(rendered),
		})
	}

	return files, nil
}
