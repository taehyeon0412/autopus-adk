package orchestra

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteInputRound_CreatesCorrectJSON verifies S1: WriteInputRound creates
// a JSON input file with correct content and 0o600 permissions.
func TestWriteInputRound_CreatesCorrectJSON(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-write-input-round")
	require.NoError(t, err)
	defer sess.Cleanup()

	// When: WriteInputRound is called for claude round 2
	err = sess.WriteInputRound("claude", 2, "debate prompt round 2")
	require.NoError(t, err)

	// Then: the input file exists with correct name
	expectedName := RoundSignalName("claude", 2, "input.json")
	inputPath := filepath.Join(sess.Dir(), expectedName)
	info, err := os.Stat(inputPath)
	require.NoError(t, err, "input file must exist")

	// Then: file has 0o600 permissions
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// Then: file contains valid JSON with correct fields
	data, err := os.ReadFile(inputPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "debate prompt round 2")
}

// TestWriteInput_Convenience verifies WriteInput is a convenience wrapper
// that delegates to WriteInputRound with the provider name.
func TestWriteInput_Convenience(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-write-input-conv")
	require.NoError(t, err)
	defer sess.Cleanup()

	// When: WriteInput is called (no round parameter)
	err = sess.WriteInput("gemini", "simple prompt")
	require.NoError(t, err)

	// Then: an input file is created in the session directory
	entries, err := os.ReadDir(sess.Dir())
	require.NoError(t, err)
	found := false
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			found = true
			break
		}
	}
	assert.True(t, found, "WriteInput must create a JSON input file")
}

// TestWaitForReadyCtx_DetectsReadyFile verifies S4: WaitForReadyCtx
// returns nil when the ready file appears before timeout.
func TestWaitForReadyCtx_DetectsReadyFile(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-wait-ready-ctx")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Given: a ready file will appear after 100ms
	readyName := RoundSignalName("claude", 1, "ready")
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(sess.Dir(), readyName), []byte("1"), 0o644)
	}()

	// When: WaitForReadyCtx polls for the ready file
	err = sess.WaitForReadyCtx(context.Background(), 2*time.Second, "claude", 1)

	// Then: no error (file detected)
	assert.NoError(t, err)
}

// TestWaitForReadyCtx_Timeout verifies S5: WaitForReadyCtx returns
// an error when the ready file does not appear within the timeout.
func TestWaitForReadyCtx_Timeout(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-wait-ready-timeout")
	require.NoError(t, err)
	defer sess.Cleanup()

	// When: WaitForReadyCtx is called with a short timeout and no ready file
	err = sess.WaitForReadyCtx(context.Background(), 300*time.Millisecond, "claude", 1)

	// Then: timeout error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestWaitForReady_Convenience verifies WaitForReady is a convenience
// wrapper that calls WaitForReadyCtx with context.Background().
func TestWaitForReady_Convenience(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-wait-ready-conv")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Given: a ready file will appear after 100ms
	readyName := RoundSignalName("opencode", 3, "ready")
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(sess.Dir(), readyName), []byte("1"), 0o644)
	}()

	// When: WaitForReady (non-ctx) is called
	err = sess.WaitForReady(2*time.Second, "opencode", 3)

	// Then: no error
	assert.NoError(t, err)
}

// TestWriteAbortSignal_CreatesFile verifies that WriteAbortSignal creates
// an abort signal file for the given provider and round (R5-SAFETY).
func TestWriteAbortSignal_CreatesFile(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-write-abort")
	require.NoError(t, err)
	defer sess.Cleanup()

	// When: WriteAbortSignal is called
	err = sess.WriteAbortSignal("claude", 1)
	require.NoError(t, err)

	// Then: abort file exists in the session directory
	abortName := RoundSignalName("claude", 1, "abort")
	_, err = os.Stat(filepath.Join(sess.Dir(), abortName))
	assert.NoError(t, err, "abort signal file must exist")
}

// TestCleanRoundSignals_Extended_RemovesInputReadyAbort verifies S10:
// CleanRoundSignals removes done, input.json, ready, and abort files.
func TestCleanRoundSignals_Extended_RemovesInputReadyAbort(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-clean-extended")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Given: done, input.json, ready, and abort files exist for round 1
	suffixes := []string{"done", "input.json", "ready", "abort"}
	for _, suffix := range suffixes {
		fname := RoundSignalName("claude", 1, suffix)
		require.NoError(t, os.WriteFile(
			filepath.Join(sess.Dir(), fname), []byte("1"), 0o644,
		))
	}
	// Also create a result file that should be preserved
	resultName := RoundSignalName("claude", 1, "result.json")
	require.NoError(t, os.WriteFile(
		filepath.Join(sess.Dir(), resultName), []byte(`{"output":"keep"}`), 0o644,
	))

	// When: CleanRoundSignals is called
	CleanRoundSignals(sess, 1)

	// Then: done, input.json, ready, abort files are removed
	for _, suffix := range suffixes {
		fname := RoundSignalName("claude", 1, suffix)
		_, statErr := os.Stat(filepath.Join(sess.Dir(), fname))
		assert.True(t, os.IsNotExist(statErr),
			"%s file must be removed by CleanRoundSignals", suffix)
	}

	// Then: result file is preserved
	_, err = os.Stat(filepath.Join(sess.Dir(), resultName))
	assert.NoError(t, err, "result.json must be preserved")
}

// TestMixedMode_FileIPC_And_Fallback verifies S13: when some providers
// have hooks and others don't, file IPC is used for hook providers
// while non-hook providers are handled differently.
func TestMixedMode_FileIPC_And_Fallback(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-mixed-mode")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Given: only claude has a hook, unknown-provider does not
	sess.SetHookProviders(map[string]bool{"claude": true})

	// Then: claude should use file IPC path
	assert.True(t, sess.HasHook("claude"), "claude must have hook")

	// Then: unknown-provider should fall back
	assert.False(t, sess.HasHook("unknown-provider"),
		"unknown-provider must not have hook (triggers fallback)")

	// When: WriteInputRound is called for a hook provider, it succeeds
	err = sess.WriteInputRound("claude", 1, "file IPC prompt")
	require.NoError(t, err)

	// Then: input file exists for claude
	inputName := RoundSignalName("claude", 1, "input.json")
	_, err = os.Stat(filepath.Join(sess.Dir(), inputName))
	assert.NoError(t, err, "input file must exist for hook provider")
}
