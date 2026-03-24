package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateScenarios_GoProject_WritesFile verifies that generateScenarios
// creates scenarios.md in .autopus/project/ for a minimal Go project.
func TestGenerateScenarios_GoProject_WritesFile(t *testing.T) {
	t.Parallel()

	// Given: a minimal Go project
	dir := setupGoProject(t)
	info := &ProjectInfo{Name: "testproject"}

	// When: generateScenarios is called
	err := generateScenarios(dir, info)

	// Then: scenarios.md is written without error
	require.NoError(t, err)
	scenariosPath := filepath.Join(dir, ".autopus", "project", "scenarios.md")
	assert.FileExists(t, scenariosPath)

	data, err := os.ReadFile(scenariosPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "# E2E Scenarios — testproject")
}

// TestGenerateScenarios_EmptyDir_WritesMinimalFile verifies that when no
// Cobra commands are found, a minimal file (0 scenarios) is still written.
func TestGenerateScenarios_EmptyDir_WritesMinimalFile(t *testing.T) {
	t.Parallel()

	// Given: a project directory with no Go files (no Cobra commands)
	dir := t.TempDir()
	info := &ProjectInfo{Name: "emptyproject"}

	// When: generateScenarios is called
	err := generateScenarios(dir, info)

	// Then: a minimal scenarios.md is written
	require.NoError(t, err)
	scenariosPath := filepath.Join(dir, ".autopus", "project", "scenarios.md")
	assert.FileExists(t, scenariosPath)

	data, err := os.ReadFile(scenariosPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "emptyproject")
}

// TestGenerateScenarios_WithCobraCommands_ScenariosIncluded verifies that
// a project with Cobra commands results in scenario entries in scenarios.md.
func TestGenerateScenarios_WithCobraCommands_ScenariosIncluded(t *testing.T) {
	t.Parallel()

	// Given: a project with a Cobra command definition
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/testcli\n\ngo 1.21\n")
	writeFile(t, dir, "cmd/root.go", `package cmd

import "github.com/spf13/cobra"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run:   func(cmd *cobra.Command, args []string) {},
}
`)
	info := &ProjectInfo{Name: "cobraproject"}

	// When: generateScenarios is called
	err := generateScenarios(dir, info)

	// Then: scenarios.md contains scenario entries
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(dir, ".autopus", "project", "scenarios.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "S1:")
}

// TestGenerateScenarios_ReadonlyDir_ReturnsError verifies that when the
// output directory cannot be created, an error is returned.
func TestGenerateScenarios_ReadonlyDir_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: a read-only parent directory
	parent := t.TempDir()
	readonlyDir := filepath.Join(parent, "readonly")
	require.NoError(t, os.MkdirAll(readonlyDir, 0o555)) // r-xr-xr-x
	t.Cleanup(func() { _ = os.Chmod(readonlyDir, 0o755) })

	info := &ProjectInfo{Name: "readonly-project"}

	// When: generateScenarios attempts to create .autopus/project under readonly dir
	err := generateScenarios(readonlyDir, info)

	// Then: an error is returned (cannot create directory)
	assert.Error(t, err)
}

// TestGenerateScenarios_ProjectNameInHeader verifies the project name appears
// in the generated scenarios.md header.
func TestGenerateScenarios_ProjectNameInHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		projectName string
	}{
		{"simple name", "myapp"},
		{"hyphenated name", "my-cool-app"},
		{"name with spaces", "My Project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			info := &ProjectInfo{Name: tt.projectName}

			err := generateScenarios(dir, info)
			require.NoError(t, err)

			data, err := os.ReadFile(filepath.Join(dir, ".autopus", "project", "scenarios.md"))
			require.NoError(t, err)
			assert.Contains(t, string(data), tt.projectName)
		})
	}
}
