// Package cli_test contains tests for the learn command group.
package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupLearnDir creates a temp dir with the .autopus/learnings/ structure.
func setupLearnDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	learningsDir := filepath.Join(dir, ".autopus", "learnings")
	require.NoError(t, os.MkdirAll(learningsDir, 0o755))
	return dir
}

// chdir changes to dir and registers cleanup to restore the original wd.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

// --- learn parent command ---

func TestLearnCmd_HasFourSubcommands(t *testing.T) {
	cmd := newTestRootCmd()
	learnCmd, _, err := cmd.Find([]string{"learn"})
	require.NoError(t, err)

	subNames := make([]string, 0, len(learnCmd.Commands()))
	for _, sub := range learnCmd.Commands() {
		subNames = append(subNames, sub.Name())
	}

	assert.Contains(t, subNames, "query")
	assert.Contains(t, subNames, "record")
	assert.Contains(t, subNames, "prune")
	assert.Contains(t, subNames, "summary")
	assert.Len(t, subNames, 4, "learn must have exactly 4 subcommands")
}

func TestLearnCmd_NoSubcommand_PrintsHelp(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "learn")
}

// --- learn query ---

func TestLearnQuery_NoEntries_PrintsNoMatching(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn", "query", "--files", "foo.go"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "No matching entries found.")
}

func TestLearnQuery_FlagParsing_MultipleFlags(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"learn", "query",
		"--files", "a.go,b.go",
		"--packages", "pkg/core",
		"--keywords", "test",
	})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "No matching entries found.")
}

func TestLearnQuery_UnknownFlag_ReturnsError(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn", "query", "--nonexistent", "val"})
	err := cmd.Execute()
	assert.Error(t, err, "unknown flag should produce an error")
}
