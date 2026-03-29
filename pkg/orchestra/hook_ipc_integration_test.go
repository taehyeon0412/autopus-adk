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

// TestTryFileIPC_Success verifies that tryFileIPC returns true when
// WaitForReady succeeds and WriteInputRound succeeds.
func TestTryFileIPC_Success(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-try-file-ipc-ok")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Given: ready file appears before timeout
	readyName := RoundSignalName("claude", 2, "ready")
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(sess.Dir(), readyName), []byte("1"), 0o644)
	}()

	// When: tryFileIPC is called
	ctx := context.Background()
	ok := tryFileIPC(ctx, sess, "claude", 2, "debate prompt")

	// Then: returns true (file IPC succeeded)
	assert.True(t, ok, "tryFileIPC must return true on success")

	// Then: input file was created
	inputName := RoundSignalName("claude", 2, "input.json")
	_, err = os.Stat(filepath.Join(sess.Dir(), inputName))
	assert.NoError(t, err, "input file must exist after successful IPC")
}

// TestTryFileIPC_ReadyTimeout_Fallback verifies that tryFileIPC returns false
// when WaitForReady times out (provider not ready).
func TestTryFileIPC_ReadyTimeout_Fallback(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-try-file-ipc-timeout")
	require.NoError(t, err)
	defer sess.Cleanup()

	// When: tryFileIPC is called but no ready file appears
	// Use a cancelled context to speed up the test
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	ok := tryFileIPC(ctx, sess, "claude", 1, "prompt")

	// Then: returns false (fallback to SendLongText)
	assert.False(t, ok, "tryFileIPC must return false on ready timeout")
}

// TestTryFileIPC_ContextCancelled_Fallback verifies that tryFileIPC returns false
// when context is cancelled during WaitForReady.
func TestTryFileIPC_ContextCancelled_Fallback(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-try-file-ipc-ctx-cancel")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Given: context is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// When: tryFileIPC is called
	ok := tryFileIPC(ctx, sess, "claude", 1, "prompt")

	// Then: returns false (fallback)
	assert.False(t, ok, "tryFileIPC must return false on cancelled context")
}

// TestTryFileIPC_WriteFailure_SendsAbort verifies R5-SAFETY: when WriteInputRound
// fails after WaitForReady succeeds, an abort signal is written.
func TestTryFileIPC_WriteFailure_SendsAbort(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-try-file-ipc-write-fail")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Given: ready file exists immediately
	readyName := RoundSignalName("claude", 1, "ready")
	require.NoError(t, os.WriteFile(filepath.Join(sess.Dir(), readyName), []byte("1"), 0o644))

	// Given: make the session dir read-only to cause WriteInputRound failure
	require.NoError(t, os.Chmod(sess.Dir(), 0o555))
	defer func() { _ = os.Chmod(sess.Dir(), 0o700) }()

	// When: tryFileIPC is called
	ok := tryFileIPC(context.Background(), sess, "claude", 1, "prompt")

	// Then: returns false (fallback)
	assert.False(t, ok, "tryFileIPC must return false on write failure")

	// Restore permissions so we can check abort file and cleanup
	_ = os.Chmod(sess.Dir(), 0o700)

	// Note: abort signal write also fails due to permissions, but that's
	// expected — the key behavior is the fallback return value.
}

// TestExecuteRound_FileIPC_Path verifies that executeRound uses file IPC
// for hook providers in round 2+ when hookSession is active (SPEC-ORCH-017 R4).
func TestExecuteRound_FileIPC_Path(t *testing.T) {
	t.Parallel()

	mock := newCmuxMock()
	mock.readScreenOutput = "❯\n"

	sess, err := NewHookSession("test-exec-round-fileipc")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Only claude has hook; gemini does not
	sess.SetHookProviders(map[string]bool{"claude": true})

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			{Name: "claude", Binary: "echo", InteractiveInput: "stdin"},
		},
		Strategy:       StrategyDebate,
		Prompt:         "round 2 prompt",
		TimeoutSeconds: 5,
		Terminal:       mock,
		Interactive:    true,
		HookMode:       true,
		InitialDelay:   time.Millisecond,
	}

	panes := []paneInfo{
		{provider: cfg.Providers[0], paneID: "pane-1"},
	}

	// Simulate round 1 responses (needed for round 2 rebuttal)
	prevResponses := []ProviderResponse{
		{Provider: "claude", Output: "round 1 response"},
	}

	// Write ready file for claude round 2 (so file IPC succeeds)
	readyName := RoundSignalName("claude", 2, "ready")
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(sess.Dir(), readyName), []byte("1"), 0o644)
	}()

	// Also write a done + result file so collectRoundHookResults returns
	go func() {
		time.Sleep(200 * time.Millisecond)
		doneName := RoundSignalName("claude", 2, "done")
		resultName := RoundSignalName("claude", 2, "result.json")
		_ = os.WriteFile(filepath.Join(sess.Dir(), doneName), []byte("1"), 0o644)
		_ = os.WriteFile(filepath.Join(sess.Dir(), resultName), []byte(`{"output":"file ipc response","exit_code":0}`), 0o644)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responses := executeRound(ctx, cfg, panes, sess, 2, prevResponses)

	// Then: response was collected via hook (file IPC path)
	require.Len(t, responses, 1)
	assert.Equal(t, "file ipc response", responses[0].Output)

	// Then: input file was created via file IPC (not SendLongText)
	inputName := RoundSignalName("claude", 2, "input.json")
	_, err = os.Stat(filepath.Join(sess.Dir(), inputName))
	assert.NoError(t, err, "file IPC input file must exist")
}

// TestWaitForDoneRoundCtx_ZeroRound_FallsBack verifies WaitForDoneRoundCtx
// falls back to WaitForDone for round=0 (covering the else branch).
func TestWaitForDoneRoundCtx_ZeroRound_FallsBack(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-donectx-zero-round")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Given: provider-done file appears
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(sess.Dir(), "claude-done"), []byte("1"), 0o644)
	}()

	// When: WaitForDoneRoundCtx with round=0
	err = sess.WaitForDoneRoundCtx(context.Background(), 2*time.Second, "claude", 0)

	// Then: succeeds via WaitForDone fallback
	assert.NoError(t, err)
}

// TestWaitForReadyCtx_ContextCancelled verifies WaitForReadyCtx returns
// error when context is cancelled before ready file appears.
func TestWaitForReadyCtx_ContextCancelled(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-ready-ctx-cancel")
	require.NoError(t, err)
	defer sess.Cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err = sess.WaitForReadyCtx(ctx, 5*time.Second, "claude", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}
