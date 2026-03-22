package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/sigmap"
)

const signaturesDir = ".autopus/context"
const signaturesFile = "signatures.md"

// generateSignatureMap creates the signature map file if enabled in config.
// A nil config is treated as enabled (default behavior).
func generateSignatureMap(projectDir string, cfg *config.HarnessConfig) error {
	if cfg != nil && !cfg.Context.SignatureMap {
		return nil
	}

	sm, err := sigmap.Extract(projectDir)
	if err != nil {
		return fmt.Errorf("extract signatures: %w", err)
	}

	if err := sigmap.CalculateFanIn(projectDir, sm); err != nil {
		return fmt.Errorf("calculate fan-in: %w", err)
	}

	content := sigmap.Render(sm)

	outDir := filepath.Join(projectDir, signaturesDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create context directory: %w", err)
	}

	outPath := filepath.Join(outDir, signaturesFile)
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write signatures: %w", err)
	}

	return nil
}

// updateSignatureMap regenerates the signature map and reports whether the
// file content changed. A nil config is treated as enabled.
func updateSignatureMap(projectDir string, cfg *config.HarnessConfig) (bool, error) {
	if cfg != nil && !cfg.Context.SignatureMap {
		return false, nil
	}

	outPath := filepath.Join(projectDir, signaturesDir, signaturesFile)

	// Read existing content for comparison; ignore errors (file may not exist yet).
	oldContent, _ := os.ReadFile(outPath)

	sm, err := sigmap.Extract(projectDir)
	if err != nil {
		return false, fmt.Errorf("extract signatures: %w", err)
	}

	if err := sigmap.CalculateFanIn(projectDir, sm); err != nil {
		return false, fmt.Errorf("calculate fan-in: %w", err)
	}

	newContent := sigmap.Render(sm)

	if string(oldContent) == newContent {
		return false, nil // no changes
	}

	outDir := filepath.Join(projectDir, signaturesDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return false, fmt.Errorf("create context directory: %w", err)
	}

	if err := os.WriteFile(outPath, []byte(newContent), 0644); err != nil {
		return false, fmt.Errorf("write signatures: %w", err)
	}

	return true, nil
}
