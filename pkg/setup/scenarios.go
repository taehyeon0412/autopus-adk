package setup

import (
	"os"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/e2e"
)

// @AX:NOTE [AUTO] @AX:REASON: design choice — extraction failures are non-fatal; writes minimal empty ScenarioSet on error to avoid blocking setup flow; fan_in=2 (engine.go:Generate and engine.go:Update)
// generateScenarios extracts and writes scenarios.md from the project codebase.
func generateScenarios(projectDir string, info *ProjectInfo) error {
	absDir, _ := filepath.Abs(projectDir)

	// Extract scenarios from project codebase.
	scenarios, err := e2e.ExtractCobra(absDir)
	if err != nil {
		// Non-fatal: if extraction fails, write a minimal file.
		scenarios = []e2e.Scenario{}
	}

	set := &e2e.ScenarioSet{
		ProjectName: info.Name,
		ProjectType: "Library",
		Binary:      "N/A",
		Build:       "N/A",
		Scenarios:   scenarios,
	}

	content, _ := e2e.RenderScenarios(set)

	// Ensure .autopus/project directory exists.
	scenariosDir := filepath.Join(absDir, ".autopus", "project")
	if err := os.MkdirAll(scenariosDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(scenariosDir, "scenarios.md"), content, 0644)
}
