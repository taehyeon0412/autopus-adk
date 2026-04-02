package security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupAutonomousMode(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	// Create a policy file.
	policyPath := filepath.Join(workDir, "policy.json")
	require.NoError(t, os.WriteFile(policyPath, []byte(`{"allowed_commands":["go "]}`), 0600))

	err := SetupAutonomousMode(workDir, policyPath)
	require.NoError(t, err)

	// Verify hook config was created.
	settingsPath := filepath.Join(workDir, ".claude", "settings.json")
	_, err = os.Stat(settingsPath)
	assert.NoError(t, err, "settings.json should be created")
}

func TestSetupAutonomousModeFailsOnMissingPolicy(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	err := SetupAutonomousMode(workDir, "/nonexistent/policy.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy file not accessible")
}

func TestCleanupAutonomousMode(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	// Setup first.
	policyPath := filepath.Join(workDir, "policy.json")
	require.NoError(t, os.WriteFile(policyPath, []byte(`{}`), 0600))
	require.NoError(t, SetupAutonomousMode(workDir, policyPath))

	// Cleanup.
	err := CleanupAutonomousMode(workDir)
	require.NoError(t, err)
}

func TestCleanupAutonomousModeNoSettings(t *testing.T) {
	t.Parallel()

	// Cleanup on a dir with no .claude/settings.json should succeed.
	workDir := t.TempDir()
	err := CleanupAutonomousMode(workDir)
	require.NoError(t, err)
}

func TestSetupAutonomousModeWriteFailure(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	policyPath := filepath.Join(workDir, "policy.json")
	require.NoError(t, os.WriteFile(policyPath, []byte(`{}`), 0600))

	// Block .claude dir creation by placing a file there.
	require.NoError(t, os.WriteFile(filepath.Join(workDir, ".claude"), []byte("block"), 0600))

	err := SetupAutonomousMode(workDir, policyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setup autonomous hooks")
}

func TestCleanupAutonomousModeRemoveFailure(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	settingsDir := filepath.Join(workDir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))

	// Write invalid JSON so RemoveHookConfig fails.
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte("{bad"), 0644))

	err := CleanupAutonomousMode(workDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cleanup autonomous hooks")
}

func TestAutonomousFlags(t *testing.T) {
	t.Parallel()

	flags := AutonomousFlags()
	require.Len(t, flags, 1)
	assert.Equal(t, "--dangerously-skip-permissions", flags[0])
}
