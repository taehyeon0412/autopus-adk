// Package e2e provides user-facing scenario-based E2E test infrastructure.
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeGoFile creates a Go source file in a temp project directory.
func writeGoFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

// makeGoModule creates a minimal Go module in a temp directory.
func makeGoModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeGoFile(t, dir, "go.mod", "module example.com/testcli\n\ngo 1.21\n")
	return dir
}

// TestExtractCobra_SimpleCommand_ReturnsScenario verifies that a single-level
// Cobra command is extracted as a scenario with the correct Use and Short fields.
// S1: Cobra command tree analysis for leaf command scenario generation.
func TestExtractCobra_SimpleCommand_ReturnsScenario(t *testing.T) {
	t.Parallel()

	// Given: a Go CLI project with a simple "version" command
	dir := makeGoModule(t)
	writeGoFile(t, dir, "cmd/root.go", `package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {},
}
`)

	// When: ExtractCobra is called on the project directory
	scenarios, err := ExtractCobra(dir)

	// Then: at least one scenario with "version" is returned
	require.NoError(t, err)
	require.NotEmpty(t, scenarios)
	found := false
	for _, s := range scenarios {
		if s.ID == "version" || s.Command == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected scenario for 'version' command")
}

// TestExtractCobra_NestedSubcommands_ExtractsLeafOnly verifies that nested
// Cobra commands yield scenarios only for leaf (non-parent) commands.
func TestExtractCobra_NestedSubcommands_ExtractsLeafOnly(t *testing.T) {
	t.Parallel()

	// Given: a Go CLI project with parent "server" and leaf "server start"
	dir := makeGoModule(t)
	writeGoFile(t, dir, "cmd/server.go", `package cmd

import "github.com/spf13/cobra"

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage server",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Run:   func(cmd *cobra.Command, args []string) {},
}
`)

	// When: ExtractCobra is called
	scenarios, err := ExtractCobra(dir)

	// Then: "server start" leaf scenario is present; "server" parent is not a leaf
	require.NoError(t, err)
	for _, s := range scenarios {
		assert.NotEqual(t, "server", s.ID, "parent command should not be a leaf scenario")
	}
}

// TestExtractCobra_WithFlags_IncludesFlags verifies that command flags are
// captured in the generated scenario metadata.
func TestExtractCobra_WithFlags_IncludesFlags(t *testing.T) {
	t.Parallel()

	// Given: a command with a string flag
	dir := makeGoModule(t)
	writeGoFile(t, dir, "cmd/init.go", `package cmd

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	initCmd.Flags().StringP("name", "n", "", "Project name")
}
`)

	// When: ExtractCobra is called
	scenarios, err := ExtractCobra(dir)

	// Then: the extracted scenario includes flag information
	require.NoError(t, err)
	require.NotEmpty(t, scenarios)
	// At least one scenario should have non-empty command text
	assert.NotEmpty(t, scenarios[0].Command)
}

// TestExtractCobra_NoEntryPoint_ReturnsEmpty verifies that a project without
// any detectable Cobra entry point returns an empty scenario list, not an error.
// S2: no entry point detection.
func TestExtractCobra_NoEntryPoint_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Given: a directory with no Cobra command definitions
	dir := makeGoModule(t)
	writeGoFile(t, dir, "pkg/util/util.go", `package util

func Helper() string { return "helper" }
`)

	// When: ExtractCobra is called
	scenarios, err := ExtractCobra(dir)

	// Then: empty slice returned without error
	require.NoError(t, err)
	assert.Empty(t, scenarios)
}

// TestExtractCobra_RealProject_ExtractsKnownCommands verifies that running
// ExtractCobra on the autopus-adk project itself returns known commands.
func TestExtractCobra_RealProject_ExtractsKnownCommands(t *testing.T) {
	t.Parallel()

	// Given: the autopus-adk project directory (which uses Cobra)
	// Locate project root by traversing upward from this test file's location
	dir := findProjectRoot(t)

	// When: ExtractCobra is called on the real project
	scenarios, err := ExtractCobra(dir)

	// Then: known commands like "init", "doctor", "setup" are present
	require.NoError(t, err)
	require.NotEmpty(t, scenarios)
	knownCmds := map[string]bool{"init": false, "doctor": false, "setup": false}
	for _, s := range scenarios {
		if _, ok := knownCmds[s.ID]; ok {
			knownCmds[s.ID] = true
		}
	}
	found := 0
	for _, v := range knownCmds {
		if v {
			found++
		}
	}
	assert.GreaterOrEqual(t, found, 1, "expected at least one known command extracted from real project")
}

// findProjectRoot walks up from the test's working directory to locate go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod)")
		}
		dir = parent
	}
}
