// Package cli tests internal check_rules functions via white-box testing.
package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHasValidLoreType_AllKnownTypes verifies that every registered Lore type
// prefix is recognised correctly.
func TestHasValidLoreType_AllKnownTypes(t *testing.T) {
	t.Parallel()

	validPrefixes := []string{
		"feat(cli): add something",
		"fix(config): correct typo",
		"refactor(pkg): clean up logic",
		"test(api): add unit tests",
		"docs(readme): update guide",
		"chore(deps): bump version",
		"perf(cache): reduce allocations",
	}

	for _, msg := range validPrefixes {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()
			assert.True(t, hasValidLoreType(msg), "expected valid Lore type for: %q", msg)
		})
	}
}

// TestHasValidLoreType_InvalidMessages verifies that non-Lore messages return false.
func TestHasValidLoreType_InvalidMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  string
	}{
		{"empty string", ""},
		{"plain english", "Update README"},
		{"lowercase type without parens", "feat: missing parens"},
		{"wrong type", "update(cli): something"},
		{"typo in type", "fixt(api): typo"},
		{"mixed case", "Feat(scope): capitalized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.False(t, hasValidLoreType(tt.msg), "expected invalid Lore type for: %q", tt.msg)
		})
	}
}

// TestCheckLore_ValidLoreCommit verifies checkLore returns true for a commit
// that has both a valid type prefix and the Lore sign-off.
func TestCheckLore_ValidLoreCommit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Initialize git repo on main branch.
	gitCmds := [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	}
	for _, args := range gitCmds {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v failed: %s", args, out)
	}

	// Create a file and commit with valid Lore format.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dummy.go"), []byte("package dummy\n"), 0o644))

	commitMsg := "feat(cli): add feature\n\nWhy: testing\nDecision: yes\nAlternatives: none\n\n🐙 Autopus <noreply@autopus.co>"
	addCmd := exec.Command("git", "add", "dummy.go")
	addCmd.Dir = dir
	require.NoError(t, addCmd.Run())

	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = dir
	out, err := commitCmd.CombinedOutput()
	require.NoError(t, err, "git commit failed: %s", out)

	var buf bytes.Buffer
	result := checkLore(dir, &buf, false)
	assert.True(t, result, "checkLore must return true for a valid Lore commit")
	assert.Contains(t, buf.String(), "OK", "output must indicate success")
}

// TestCheckLore_InvalidLoreCommit verifies checkLore returns false when both
// the type prefix and sign-off are missing.
func TestCheckLore_InvalidLoreCommit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	gitCmds := [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	}
	for _, args := range gitCmds {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v failed: %s", args, out)
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "dummy.go"), []byte("package dummy\n"), 0o644))

	addCmd := exec.Command("git", "add", "dummy.go")
	addCmd.Dir = dir
	require.NoError(t, addCmd.Run())

	// Commit without Lore format (no type prefix, no sign-off).
	commitCmd := exec.Command("git", "commit", "-m", "Update something without lore format")
	commitCmd.Dir = dir
	out, err := commitCmd.CombinedOutput()
	require.NoError(t, err, "git commit failed: %s", out)

	var buf bytes.Buffer
	result := checkLore(dir, &buf, false)
	assert.False(t, result, "checkLore must return false for an invalid commit")
}

// TestCheckLore_MissingSignOffOnly verifies checkLore returns false when
// the type prefix is valid but the sign-off line is missing.
func TestCheckLore_MissingSignOffOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	gitCmds := [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	}
	for _, args := range gitCmds {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v failed: %s", args, out)
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "dummy.go"), []byte("package dummy\n"), 0o644))

	addCmd := exec.Command("git", "add", "dummy.go")
	addCmd.Dir = dir
	require.NoError(t, addCmd.Run())

	// Has valid type prefix but no sign-off.
	commitCmd := exec.Command("git", "commit", "-m", "fix(api): correct logic\n\nWhy: needed")
	commitCmd.Dir = dir
	out, err := commitCmd.CombinedOutput()
	require.NoError(t, err, "git commit failed: %s", out)

	var buf bytes.Buffer
	result := checkLore(dir, &buf, false)
	assert.False(t, result, "checkLore must return false when sign-off is missing")
	assert.Contains(t, buf.String(), "sign-off", "output must mention missing sign-off")
}

// TestCheckLore_QuietModeNoOutput verifies that quiet mode suppresses section headers.
func TestCheckLore_QuietModeNoOutput(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // no git repo, so lastCommitMessage will fail → returns true quietly

	var buf bytes.Buffer
	result := checkLore(dir, &buf, true)
	assert.True(t, result, "checkLore must return true when no commits exist")
	assert.Empty(t, buf.String(), "quiet mode must produce no output for the no-commit case")
}

// TestCheckArch_WarnRangeFile verifies that files between 201-300 lines produce
// a SKIP/warn message in non-quiet mode but do not fail.
func TestCheckArch_WarnRangeFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Write a file with exactly 201 lines (warn zone: 200 < n <= 300).
	var sb strings.Builder
	sb.WriteString("package dummy\n")
	for i := 0; i < 200; i++ {
		sb.WriteString("// line\n")
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "warn.go"), []byte(sb.String()), 0o644))

	var buf bytes.Buffer
	result := checkArch(dir, &buf, false)
	assert.True(t, result, "warn-range file must not fail arch check")
	// The output should mention "consider splitting".
	assert.Contains(t, buf.String(), "consider splitting")
}

// TestCheckArch_WarnRangeQuiet verifies that warn-range files produce no output
// in quiet mode and still pass.
func TestCheckArch_WarnRangeQuiet(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var sb strings.Builder
	sb.WriteString("package dummy\n")
	for i := 0; i < 200; i++ {
		sb.WriteString("// line\n")
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "warn.go"), []byte(sb.String()), 0o644))

	var buf bytes.Buffer
	result := checkArch(dir, &buf, true) // quiet=true
	assert.True(t, result, "warn-range file must pass in quiet mode")
	assert.Empty(t, buf.String(), "quiet mode must suppress warn-range output")
}

// TestCountLines_ReturnsCorrectCount verifies countLines returns the correct
// number of lines for a known file.
func TestCountLines_ReturnsCorrectCount(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "line1\nline2\nline3\n"
	path := filepath.Join(dir, "test.go")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	n, err := countLines(path)
	require.NoError(t, err)
	assert.Equal(t, 3, n)
}

// TestCountLines_NonExistentFile verifies countLines returns an error for missing files.
func TestCountLines_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := countLines("/nonexistent/path/file.go")
	assert.Error(t, err)
}
