package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDir creates a temp directory with basic platform directory structure.
func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create platform-specific directories.
	platformDirs := []string{
		".codex",
		".gemini",
		".claude",
		".autopus",
	}
	for _, d := range platformDirs {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, d), 0755))
	}
	return dir
}

// assertFileExists asserts that a file exists at the given path.
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, "expected file to exist: %s", path)
}

// assertFileContains asserts that the file at path contains the given substring.
func assertFileContains(t *testing.T, path, substring string) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %s", path)
	assert.Contains(t, string(data), substring,
		"file %s should contain %q", path, substring)
}

// assertFileNotContains asserts that the file at path does NOT contain the given substring.
func assertFileNotContains(t *testing.T, path, substring string) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %s", path)
	assert.NotContains(t, string(data), substring,
		"file %s should not contain %q", path, substring)
}

// assertLineCount asserts that the file at path has at most maxLines lines.
func assertLineCount(t *testing.T, path string, maxLines int) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %s", path)
	lines := strings.Count(string(data), "\n")
	// Account for content that doesn't end with newline.
	if len(data) > 0 && data[len(data)-1] != '\n' {
		lines++
	}
	assert.LessOrEqual(t, lines, maxLines,
		"file %s has %d lines, expected at most %d", path, lines, maxLines)
}

// --- Tests for the test utilities themselves ---

func TestSetupTestDir(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)
	assert.DirExists(t, filepath.Join(dir, ".codex"))
	assert.DirExists(t, filepath.Join(dir, ".gemini"))
	assert.DirExists(t, filepath.Join(dir, ".claude"))
	assert.DirExists(t, filepath.Join(dir, ".autopus"))
}

func TestAssertFileExists_Pass(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.txt")
	require.NoError(t, os.WriteFile(path, []byte("data"), 0644))
	assertFileExists(t, path)
}

func TestAssertFileContains_Pass(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "content.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello world"), 0644))
	assertFileContains(t, path, "world")
}

func TestAssertFileNotContains_Pass(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "content.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello world"), 0644))
	assertFileNotContains(t, path, "foobar")
}

func TestAssertLineCount_UnderLimit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "small.txt")
	content := "line1\nline2\nline3\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	assertLineCount(t, path, 3)
}

func TestAssertLineCount_ExactLimit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "exact.txt")
	content := "a\nb\nc\nd\ne\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	assertLineCount(t, path, 5)
}

func TestAssertLineCount_NoTrailingNewline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "notail.txt")
	content := "line1\nline2" // 2 lines, no trailing newline
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	assertLineCount(t, path, 2)
}

func TestAssertFileContains_MultipleSubstrings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.txt")
	content := "alpha\nbeta\ngamma\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	assertFileContains(t, path, "alpha")
	assertFileContains(t, path, "gamma")
	assertFileNotContains(t, path, "delta")
}
