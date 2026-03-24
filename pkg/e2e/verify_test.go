// Package e2e provides user-facing scenario-based E2E test infrastructure.
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExitCode_Matches_ReturnsPass verifies that exit_code(N) returns PASS
// when the actual exit code matches N.
// S9: verification primitive success.
func TestExitCode_Matches_ReturnsPass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected int
		actual   int
	}{
		{"zero exit code", 0, 0},
		{"nonzero exit code match", 1, 1},
		{"exit code 2 match", 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Given: a VerifyResult with matching exit codes
			// When: ExitCode primitive is evaluated
			result := CheckExitCode(tt.expected, tt.actual)

			// Then: result is PASS
			assert.True(t, result.Pass)
			assert.Empty(t, result.Message)
		})
	}
}

// TestExitCode_Mismatch_ReturnsFail verifies that exit_code(N) returns FAIL
// with a descriptive message when actual exit code differs from N.
// S10: verification primitive failure.
func TestExitCode_Mismatch_ReturnsFail(t *testing.T) {
	t.Parallel()

	// Given: expected exit code 0 but actual is 1
	// When: ExitCode primitive is evaluated
	result := CheckExitCode(0, 1)

	// Then: result is FAIL with expected/got message
	assert.False(t, result.Pass)
	assert.Contains(t, result.Message, "expected exit_code(0)")
	assert.Contains(t, result.Message, "got 1")
}

// TestStdoutContains_Found_ReturnsPass verifies that stdout_contains(str)
// returns PASS when the string is present in stdout.
// S9: stdout_contains primitive success.
func TestStdoutContains_Found_ReturnsPass(t *testing.T) {
	t.Parallel()

	// Given: stdout output containing target string
	// When: StdoutContains primitive is evaluated
	result := CheckStdoutContains("version", "auto version 1.2.3")

	// Then: result is PASS
	assert.True(t, result.Pass)
}

// TestStdoutContains_NotFound_ReturnsFail verifies that stdout_contains(str)
// returns FAIL when the string is absent from stdout.
func TestStdoutContains_NotFound_ReturnsFail(t *testing.T) {
	t.Parallel()

	// Given: stdout that does NOT contain the target string
	// When: StdoutContains primitive is evaluated
	result := CheckStdoutContains("version", "error: unknown command")

	// Then: result is FAIL
	assert.False(t, result.Pass)
	assert.Contains(t, result.Message, "version")
}

// TestStderrEmpty_Empty_ReturnsPass verifies that stderr_empty() returns PASS
// when stderr output is empty.
func TestStderrEmpty_Empty_ReturnsPass(t *testing.T) {
	t.Parallel()

	// Given: empty stderr
	// When: StderrEmpty primitive is evaluated
	result := CheckStderrEmpty("")

	// Then: result is PASS
	assert.True(t, result.Pass)
}

// TestStderrEmpty_NotEmpty_ReturnsFail verifies that stderr_empty() returns FAIL
// when stderr has content.
func TestStderrEmpty_NotEmpty_ReturnsFail(t *testing.T) {
	t.Parallel()

	// Given: non-empty stderr
	// When: StderrEmpty primitive is evaluated
	result := CheckStderrEmpty("error: something went wrong")

	// Then: result is FAIL
	assert.False(t, result.Pass)
	assert.NotEmpty(t, result.Message)
}

// TestFileExists_Exists_ReturnsPass verifies that file_exists(path) returns PASS
// when the file is present on disk.
func TestFileExists_Exists_ReturnsPass(t *testing.T) {
	t.Parallel()

	// Given: a file that exists
	dir := t.TempDir()
	path := filepath.Join(dir, "output.txt")
	require.NoError(t, os.WriteFile(path, []byte("data"), 0o644))

	// When: FileExists primitive is evaluated
	result := CheckFileExists(path)

	// Then: result is PASS
	assert.True(t, result.Pass)
}

// TestFileExists_NotExists_ReturnsFail verifies that file_exists(path) returns FAIL
// when the file is absent.
func TestFileExists_NotExists_ReturnsFail(t *testing.T) {
	t.Parallel()

	// Given: a path that does not exist
	// When: FileExists primitive is evaluated
	result := CheckFileExists("/tmp/this-file-does-not-exist-e2e-test")

	// Then: result is FAIL
	assert.False(t, result.Pass)
}

// TestFileContains_Found_ReturnsPass verifies that file_contains(path, str)
// returns PASS when the file content includes the expected string.
func TestFileContains_Found_ReturnsPass(t *testing.T) {
	t.Parallel()

	// Given: a file containing the target string
	dir := t.TempDir()
	path := filepath.Join(dir, "result.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello world output"), 0o644))

	// When: FileContains primitive is evaluated
	result := CheckFileContains(path, "hello world")

	// Then: result is PASS
	assert.True(t, result.Pass)
}

// TestFileContains_NotFound_ReturnsFail verifies that file_contains(path, str)
// returns FAIL when the file content does not include the expected string.
func TestFileContains_NotFound_ReturnsFail(t *testing.T) {
	t.Parallel()

	// Given: a file that does NOT contain the target string
	dir := t.TempDir()
	path := filepath.Join(dir, "result.txt")
	require.NoError(t, os.WriteFile(path, []byte("unrelated content"), 0o644))

	// When: FileContains primitive is evaluated
	result := CheckFileContains(path, "expected_string")

	// Then: result is FAIL
	assert.False(t, result.Pass)
}
