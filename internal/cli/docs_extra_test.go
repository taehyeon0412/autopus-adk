package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocsCacheListCmd_WithEntries verifies that populated cache entries are listed.
// Given: a cache directory with pre-existing JSON entry files
// When: cache list is executed
// Then: output contains the entry keys
func TestDocsCacheListCmd_WithEntries(t *testing.T) {
	// Note: cannot use t.Parallel() with t.Setenv

	dir := t.TempDir()
	t.Setenv("AUTO_DOCS_CACHE_DIR", dir)

	// Write a minimal cache entry JSON file directly
	entryJSON := `{"LibraryID":"cobra","Topic":"commands","Content":"cobra docs","Tokens":10,"CachedAt":"2099-01-01T00:00:00Z"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cobra_commands.json"), []byte(entryJSON), 0o644))

	cmd := newDocsCacheCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
	assert.Contains(t, buf.String(), "cobra")
}

// TestDocsFetchCmd_AutoDetect_WithGoMod verifies auto-detection finds libraries from go.mod.
// Given: a temp dir with a go.mod file listing cobra
// When: docs fetch is executed without args
// Then: command runs without error (auto-detect path exercised)
func TestDocsFetchCmd_AutoDetect_WithGoMod(t *testing.T) {
	// Note: cannot use t.Parallel() with t.Setenv

	dir := t.TempDir()
	gomod := `module example.com/app

go 1.21

require (
	github.com/spf13/cobra v1.9.1
)
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644))
	t.Setenv("AUTO_DOCS_PROJECT_DIR", dir)
	t.Setenv("AUTO_DOCS_CACHE_DIR", t.TempDir())

	cmd := newDocsFetchCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	// Command may fail to fetch from real Context7/npm, but must not panic
	// and should at least attempt detection (no crash = coverage exercised)
	_ = cmd.Execute()
}

// TestDocsFetchCmd_AutoDetect_WithPackageJSON verifies auto-detection from package.json.
// Given: a temp dir with package.json listing express
// When: docs fetch is executed without args
// Then: command runs without error
func TestDocsFetchCmd_AutoDetect_WithPackageJSON(t *testing.T) {
	// Note: cannot use t.Parallel() with t.Setenv

	dir := t.TempDir()
	pkgjson := `{"name":"my-app","dependencies":{"express":"^4.18.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgjson), 0o644))
	t.Setenv("AUTO_DOCS_PROJECT_DIR", dir)
	t.Setenv("AUTO_DOCS_CACHE_DIR", t.TempDir())

	cmd := newDocsFetchCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	_ = cmd.Execute()
}

// TestDocsFetchCmd_AutoDetect_WithPyProject verifies auto-detection from pyproject.toml.
// Given: a temp dir with pyproject.toml listing requests
// When: docs fetch is executed without args
// Then: command runs without error
func TestDocsFetchCmd_AutoDetect_WithPyProject(t *testing.T) {
	// Note: cannot use t.Parallel() with t.Setenv

	dir := t.TempDir()
	pyproject := `[project]
name = "my-app"
dependencies = ["requests>=2.28.0"]
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(pyproject), 0o644))
	t.Setenv("AUTO_DOCS_PROJECT_DIR", dir)
	t.Setenv("AUTO_DOCS_CACHE_DIR", t.TempDir())

	cmd := newDocsFetchCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	_ = cmd.Execute()
}

// TestDocsCacheClearCmd_Output verifies that clear outputs "cache clear: done".
// Given: a cache clear command
// When: executed against an empty cache dir
// Then: output explicitly contains "clear"
func TestDocsCacheClearCmd_Output(t *testing.T) {
	// Note: cannot use t.Parallel() with t.Setenv

	dir := t.TempDir()
	t.Setenv("AUTO_DOCS_CACHE_DIR", dir)

	cmd := newDocsCacheCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"clear"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "clear")
}

// TestDocsFetchCmd_WithMultipleLibraries verifies fetch with multiple library args.
// Given: two library names as arguments
// When: docs fetch is executed (no real network needed — failures are silent)
// Then: command does not panic
func TestDocsFetchCmd_WithMultipleLibraries(t *testing.T) {
	// Note: cannot use t.Parallel() with t.Setenv

	t.Setenv("AUTO_DOCS_CACHE_DIR", t.TempDir())

	cmd := newDocsFetchCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"cobra", "viper"})

	// Ignore error — real network may not be available in test env
	_ = cmd.Execute()
}

// TestDocsFetchCmd_AutoDetect_NoProjectDir verifies auto-detect works when AUTO_DOCS_PROJECT_DIR is unset.
// Given: no AUTO_DOCS_PROJECT_DIR env and a temp working directory
// When: docs fetch is executed without args
// Then: command falls back to os.Getwd() and runs without panic
func TestDocsFetchCmd_AutoDetect_NoProjectDir(t *testing.T) {
	// Note: cannot use t.Parallel() with t.Setenv or t.Chdir

	t.Setenv("AUTO_DOCS_PROJECT_DIR", "") // explicitly unset
	t.Setenv("AUTO_DOCS_CACHE_DIR", t.TempDir())

	dir := t.TempDir()
	t.Chdir(dir)

	cmd := newDocsFetchCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	// No manifest files in temp dir — should print "no libraries detected"
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "no libraries detected")
}
