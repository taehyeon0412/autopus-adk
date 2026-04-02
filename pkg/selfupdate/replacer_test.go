package selfupdate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReplace_Success verifies that an existing binary is atomically replaced
// with a new binary, preserving the original file permissions.
// R4: atomic replace via os.Rename. R5: preserve file permissions.
func TestReplace_Success(t *testing.T) {
	t.Parallel()

	// Given: an existing binary with specific content and permissions
	destDir := t.TempDir()
	targetPath := filepath.Join(destDir, "autopus-adk")
	originalContent := []byte("original binary content")
	require.NoError(t, os.WriteFile(targetPath, originalContent, 0755))

	newBinaryPath := filepath.Join(t.TempDir(), "autopus-adk-new")
	newContent := []byte("new binary content v0.7.0")
	require.NoError(t, os.WriteFile(newBinaryPath, newContent, 0644))

	// When: Replace is called
	r := NewReplacer()
	err := r.Replace(newBinaryPath, targetPath)

	// Then: target contains new content and permissions are preserved
	require.NoError(t, err)

	gotContent, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, newContent, gotContent)

	info, err := os.Stat(targetPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

// TestReplace_TargetNotFound verifies that replacing a non-existent target
// returns an error from os.Stat.
func TestReplace_TargetNotFound(t *testing.T) {
	t.Parallel()

	// Given: a new binary and a target path that does not exist
	newBinaryPath := filepath.Join(t.TempDir(), "autopus-adk-new")
	require.NoError(t, os.WriteFile(newBinaryPath, []byte("new content"), 0755))
	nonExistentTarget := filepath.Join(t.TempDir(), "does-not-exist")

	// When: Replace is called with a non-existent target
	r := NewReplacer()
	err := r.Replace(newBinaryPath, nonExistentTarget)

	// Then: error is returned
	require.Error(t, err)
}

// TestReplace_PermissionError verifies that attempting to replace a binary in
// a read-only directory returns an error with actionable guidance.
// R13: on permission error, print guidance message.
func TestReplace_PermissionError(t *testing.T) {
	t.Parallel()

	// Given: a target path inside a read-only directory
	readOnlyDir := t.TempDir()
	targetPath := filepath.Join(readOnlyDir, "autopus-adk")
	require.NoError(t, os.WriteFile(targetPath, []byte("original"), 0755))
	require.NoError(t, os.Chmod(readOnlyDir, 0555))
	t.Cleanup(func() { _ = os.Chmod(readOnlyDir, 0755) })

	newBinaryPath := filepath.Join(t.TempDir(), "autopus-adk-new")
	require.NoError(t, os.WriteFile(newBinaryPath, []byte("new content"), 0755))

	// When: Replace is called on a read-only directory
	r := NewReplacer()
	err := r.Replace(newBinaryPath, targetPath)

	// Then: error is returned containing guidance
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission")
}

// TestReplaceWindows_MoveAside verifies the Windows fallback path:
// old binary is renamed to .old, new binary is placed at target.
func TestReplaceWindows_MoveAside(t *testing.T) {
	t.Parallel()

	r := NewReplacer()

	destDir := t.TempDir()
	targetPath := filepath.Join(destDir, "auto.exe")
	require.NoError(t, os.WriteFile(targetPath, []byte("old"), 0755))

	newPath := filepath.Join(t.TempDir(), "auto-new.exe")
	require.NoError(t, os.WriteFile(newPath, []byte("new"), 0755))

	err := r.replaceWindows(newPath, targetPath)
	require.NoError(t, err)

	got, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, []byte("new"), got)

	// .old should be cleaned up (or may still exist on real Windows).
	_, err = os.Stat(targetPath + ".old")
	assert.True(t, os.IsNotExist(err), ".old file should be removed")
}

// TestReplaceWindows_RestoreOnFailure verifies that if the new binary rename
// fails, the old binary is restored from .old.
func TestReplaceWindows_RestoreOnFailure(t *testing.T) {
	t.Parallel()

	r := NewReplacer()

	destDir := t.TempDir()
	targetPath := filepath.Join(destDir, "auto.exe")
	require.NoError(t, os.WriteFile(targetPath, []byte("old"), 0755))

	// Use a non-existent new binary path to trigger rename failure.
	err := r.replaceWindows("/nonexistent/path/auto.exe", targetPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rename new binary")

	// Old binary should be restored.
	got, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, []byte("old"), got)
}
