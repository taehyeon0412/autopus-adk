package orchestra

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWaitAndCollectHookResults_AllHookProviders verifies that when all
// providers have hooks configured, results are collected via file-based
// hook signal protocol (no ReadScreen fallback).
func TestWaitAndCollectHookResults_AllHookProviders(t *testing.T) {
	t.Parallel()

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			{Name: "claude", Binary: "claude"},
			{Name: "gemini", Binary: "gemini"},
		},
		TimeoutSeconds: 10,
	}

	results, err := WaitAndCollectHookResults(cfg, "test-session-all-hooks")
	require.NoError(t, err)
	assert.Len(t, results, 2, "must collect results from all providers")

	for _, r := range results {
		assert.NotEmpty(t, r.Output, "hook result output must not be empty")
		assert.False(t, r.TimedOut, "provider must not time out")
	}
}

// TestWaitAndCollectHookResults_MixedMode verifies hybrid collection:
// providers with hooks use file-based collection, others fall back to
// ReadScreen-based collection.
func TestWaitAndCollectHookResults_MixedMode(t *testing.T) {
	t.Parallel()

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			{Name: "claude", Binary: "claude"},
			{Name: "unknown-cli", Binary: "unknown-cli"},
		},
		TimeoutSeconds: 10,
	}

	results, err := WaitAndCollectHookResults(cfg, "test-session-mixed")
	require.NoError(t, err)
	assert.Len(t, results, 2, "must collect results from all providers (mixed mode)")
}

// TestWaitAndCollectHookResults_NoHooks verifies that when no providers
// have hooks configured, all results are collected via ReadScreen fallback.
func TestWaitAndCollectHookResults_NoHooks(t *testing.T) {
	t.Parallel()

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			{Name: "unknown-a", Binary: "unknown-a"},
			{Name: "unknown-b", Binary: "unknown-b"},
		},
		TimeoutSeconds: 10,
	}

	results, err := WaitAndCollectHookResults(cfg, "test-session-no-hooks")
	require.NoError(t, err)
	assert.Len(t, results, 2, "must collect results from all providers via fallback")
}

// TestWaitAndCollectHookResults_Timeout verifies that when a hook provider
// times out, the result is marked as timed out and fallback is triggered.
func TestWaitAndCollectHookResults_Timeout(t *testing.T) {
	t.Parallel()

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			{Name: "claude", Binary: "claude"},
		},
		TimeoutSeconds: 0, // immediate timeout
	}

	results, err := WaitAndCollectHookResults(cfg, "test-session-timeout")

	if err == nil {
		require.Len(t, results, 1)
		assert.True(t, results[0].TimedOut,
			"provider must be marked as timed out when hook exceeds timeout")
	} else {
		assert.Contains(t, err.Error(), "timeout",
			"error must indicate timeout")
	}
}

// TestWaitAndCollectHookResults_GracefulDegradation verifies R8: when a
// hook is configured but fails or times out, the system falls back to
// ReadScreen-based collection rather than returning an error.
func TestWaitAndCollectHookResults_GracefulDegradation(t *testing.T) {
	t.Parallel()

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			{Name: "claude", Binary: "claude"},
		},
		TimeoutSeconds: 1,
	}

	results, err := WaitAndCollectHookResults(cfg, "test-session-degradation")

	require.NoError(t, err, "graceful degradation must not return hard error")
	assert.Len(t, results, 1, "must have result even on hook failure")
}

// TestWaitAndCollectHookResults_EmptyProviders verifies behavior when
// no providers are configured.
func TestWaitAndCollectHookResults_EmptyProviders(t *testing.T) {
	t.Parallel()

	cfg := OrchestraConfig{
		Providers:      []ProviderConfig{},
		TimeoutSeconds: 5,
	}

	results, err := WaitAndCollectHookResults(cfg, "test-session-empty")
	require.NoError(t, err)
	assert.Empty(t, results, "empty provider list must return empty results")
}

// TestWaitAndCollectHookResults_HookDoneWithValidResult verifies the
// full success path: done file written + valid result.json = parsed output.
func TestWaitAndCollectHookResults_HookDoneWithValidResult(t *testing.T) {
	t.Parallel()

	// Pre-create session dir and write done + result before calling collect.
	sess, err := NewHookSession("test-session-full-success")
	require.NoError(t, err)

	hr := HookResult{Output: "hook collected output", ExitCode: 0}
	data, err := json.Marshal(hr)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(sess.Dir(), "result.json"), data, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sess.Dir(), "done"), []byte("1"), 0o644))
	sess.Cleanup()

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			{Name: "claude", Binary: "claude"},
		},
		TimeoutSeconds: 5,
	}

	results, err := WaitAndCollectHookResults(cfg, "test-session-full-success")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "claude", results[0].Provider)
}

// TestHookResultToProviderResponse verifies conversion from HookResult
// to ProviderResponse, preserving the Output field.
func TestHookResultToProviderResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		provider     string
		hookOutput   string
		hookExitCode int
		hookDuration time.Duration
		expectEmpty  bool
	}{
		{
			name:         "normal response",
			provider:     "claude",
			hookOutput:   "generated code here",
			hookExitCode: 0,
			hookDuration: 5 * time.Second,
			expectEmpty:  false,
		},
		{
			name:         "empty output",
			provider:     "gemini",
			hookOutput:   "",
			hookExitCode: 0,
			hookDuration: 2 * time.Second,
			expectEmpty:  true,
		},
		{
			name:         "non-zero exit code",
			provider:     "opencode",
			hookOutput:   "error output",
			hookExitCode: 1,
			hookDuration: 1 * time.Second,
			expectEmpty:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hr := HookResult{
				Output:   tt.hookOutput,
				ExitCode: tt.hookExitCode,
			}

			resp := HookResultToProviderResponse(hr, tt.provider, tt.hookDuration)

			assert.Equal(t, tt.provider, resp.Provider)
			assert.Equal(t, tt.hookOutput, resp.Output)
			assert.Equal(t, tt.hookExitCode, resp.ExitCode)
			assert.Equal(t, tt.hookDuration, resp.Duration)
			assert.Equal(t, tt.expectEmpty, resp.EmptyOutput)
		})
	}
}
