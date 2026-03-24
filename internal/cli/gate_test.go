package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/internal/cli"
)

// TestGateCheck_MandatoryMode_TestsMissing_ExitNonZero verifies that GateCheck
// returns a non-zero exit code in mandatory mode when required tests are absent.
func TestGateCheck_MandatoryMode_TestsMissing_ExitNonZero(t *testing.T) {
	t.Parallel()

	// Given: a directory with no test files and mandatory gate mode
	dir := t.TempDir()
	cfg := cli.GateConfig{
		GateName: "phase2",
		Mode:     cli.GateModeMandatory,
		Dir:      dir,
	}

	// When: GateCheck is called
	result := cli.GateCheck(cfg)

	// Then: the gate fails (exit non-zero)
	assert.False(t, result.Passed)
	assert.NotEmpty(t, result.Message)
}

// TestGateCheck_AdvisoryMode_TestsMissing_ExitZero verifies that GateCheck
// returns exit 0 (warning only) in advisory mode even when tests are missing.
func TestGateCheck_AdvisoryMode_TestsMissing_ExitZero(t *testing.T) {
	t.Parallel()

	// Given: a directory with no test files and advisory gate mode
	dir := t.TempDir()
	cfg := cli.GateConfig{
		GateName: "phase2",
		Mode:     cli.GateModeAdvisory,
		Dir:      dir,
	}

	// When: GateCheck is called
	result := cli.GateCheck(cfg)

	// Then: the gate passes with a warning (exit zero)
	assert.True(t, result.Passed)
	assert.NotEmpty(t, result.Warning)
}

// TestGateCheck_Phase2_PassesWhenTestsExist verifies that GateCheck passes when
// the required test files are present in the target directory.
func TestGateCheck_Phase2_PassesWhenTestsExist(t *testing.T) {
	t.Parallel()

	// Given: a directory containing a test file
	dir := t.TempDir()
	testFile := filepath.Join(dir, "foo_test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package foo_test\n"), 0o644))

	cfg := cli.GateConfig{
		GateName: "phase2",
		Mode:     cli.GateModeMandatory,
		Dir:      dir,
	}

	// When: GateCheck is called
	result := cli.GateCheck(cfg)

	// Then: the gate passes
	assert.True(t, result.Passed)
}

// TestGateCheck_InvalidGateName_ReturnsError verifies that GateCheck returns an
// error result when an unknown gate name is provided.
func TestGateCheck_InvalidGateName_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: an invalid gate name
	dir := t.TempDir()
	cfg := cli.GateConfig{
		GateName: "nonexistent-gate-xyz",
		Mode:     cli.GateModeMandatory,
		Dir:      dir,
	}

	// When: GateCheck is called
	result := cli.GateCheck(cfg)

	// Then: an error is returned
	assert.False(t, result.Passed)
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "unknown gate")
}

// TestGateCheck_Phase2_NonExistentDir_Fails verifies that GateCheck in
// mandatory mode returns Passed=false when the directory does not exist.
func TestGateCheck_Phase2_NonExistentDir_Fails(t *testing.T) {
	t.Parallel()

	// Given: a non-existent directory
	dir := filepath.Join(t.TempDir(), "no-such-dir")
	cfg := cli.GateConfig{
		GateName: "phase2",
		Mode:     cli.GateModeMandatory,
		Dir:      dir,
	}

	// When: GateCheck is called
	result := cli.GateCheck(cfg)

	// Then: the gate fails (cannot read directory)
	assert.False(t, result.Passed)
	assert.NotEmpty(t, result.Message)
}

// TestGateCheck_Phase2_Advisory_NonExistentDir_Passes verifies that advisory
// mode returns Passed=true even when the directory cannot be read.
func TestGateCheck_Phase2_Advisory_NonExistentDir_Passes(t *testing.T) {
	t.Parallel()

	// Given: a non-existent directory and advisory mode
	dir := filepath.Join(t.TempDir(), "no-such-dir")
	cfg := cli.GateConfig{
		GateName: "phase2",
		Mode:     cli.GateModeAdvisory,
		Dir:      dir,
	}

	// When: GateCheck is called
	result := cli.GateCheck(cfg)

	// Then: passes (advisory) with a warning
	assert.True(t, result.Passed)
	assert.NotEmpty(t, result.Warning)
}
