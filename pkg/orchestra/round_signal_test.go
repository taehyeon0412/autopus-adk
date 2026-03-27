package orchestra

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRoundSignalName_Format verifies that RoundSignalName produces
// the correct "{provider}-round{N}-{suffix}" format for various inputs.
func TestRoundSignalName_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		round    int
		suffix   string
		expected string
	}{
		{"round 1 done for claude", "claude", 1, "done", "claude-round1-done"},
		{"round 2 done for gemini", "gemini", 2, "done", "gemini-round2-done"},
		{"round 1 result for claude", "claude", 1, "result", "claude-round1-result"},
		{"round 10 done boundary", "opencode", 10, "done", "opencode-round10-done"},
		{"result.json suffix", "claude", 3, "result.json", "claude-round3-result.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RoundSignalName(tt.provider, tt.round, tt.suffix)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestRoundSignalName_ZeroRound verifies behavior for round 0 (edge case).
func TestRoundSignalName_ZeroRound(t *testing.T) {
	t.Parallel()
	got := RoundSignalName("claude", 0, "done")
	assert.Equal(t, "claude-round0-done", got)
}

// TestCleanRoundSignals_RemovesDoneFiles verifies that CleanRoundSignals
// removes "{provider}-round{N}-done" files but preserves result files.
func TestCleanRoundSignals_RemovesDoneFiles(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-clean-round-signals")
	require.NoError(t, err)
	defer sess.Cleanup()

	doneFile := "claude-round1-done"
	resultFile := "claude-round1-result"
	require.NoError(t, os.WriteFile(filepath.Join(sess.Dir(), doneFile), []byte("1"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sess.Dir(), resultFile), []byte("data"), 0o644))

	CleanRoundSignals(sess, 1)

	_, err = os.Stat(filepath.Join(sess.Dir(), doneFile))
	assert.True(t, os.IsNotExist(err), "done file must be removed")
	_, err = os.Stat(filepath.Join(sess.Dir(), resultFile))
	assert.NoError(t, err, "result file must be preserved")
}

// TestCleanRoundSignals_MultipleProviders verifies that done files
// for all providers are cleaned in a single call.
func TestCleanRoundSignals_MultipleProviders(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-clean-multi-provider")
	require.NoError(t, err)
	defer sess.Cleanup()

	providers := []string{"claude", "gemini", "opencode"}
	for _, p := range providers {
		fname := p + "-round2-done"
		require.NoError(t, os.WriteFile(filepath.Join(sess.Dir(), fname), []byte("1"), 0o644))
	}

	CleanRoundSignals(sess, 2)

	for _, p := range providers {
		fname := p + "-round2-done"
		_, statErr := os.Stat(filepath.Join(sess.Dir(), fname))
		assert.True(t, os.IsNotExist(statErr), "done file for %s must be removed", p)
	}
}

// TestCleanRoundSignals_NoDoneFiles verifies no error when no done files exist.
func TestCleanRoundSignals_NoDoneFiles(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-clean-no-done")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Should not panic or error when no matching files exist.
	CleanRoundSignals(sess, 99)
}

// TestSetRoundEnv_SetsEnvVariable verifies that SetRoundEnv sets
// the AUTOPUS_ROUND environment variable to the current round number.
func TestSetRoundEnv_SetsEnvVariable(t *testing.T) {
	tests := []struct {
		name     string
		round    int
		expected string
	}{
		{"round 1", 1, "1"},
		{"round 5", 5, "5"},
		{"round 10 boundary", 10, "10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetRoundEnv(tt.round)
			got := os.Getenv("AUTOPUS_ROUND")
			assert.Equal(t, tt.expected, got)
		})
	}
	_ = os.Unsetenv("AUTOPUS_ROUND")
}

// TestSetRoundEnv_ZeroRound verifies that round 0 is set correctly.
func TestSetRoundEnv_ZeroRound(t *testing.T) {
	SetRoundEnv(0)
	got := os.Getenv("AUTOPUS_ROUND")
	assert.Equal(t, "0", got)
	_ = os.Unsetenv("AUTOPUS_ROUND")
}

// TestRoundSignalName_EmptyProvider verifies empty provider name is sanitized.
func TestRoundSignalName_EmptyProvider(t *testing.T) {
	t.Parallel()
	got := RoundSignalName("", 1, "done")
	// sanitizeProviderName replaces empty with "unknown".
	assert.Equal(t, "unknown-round1-done", got)
}

// TestRoundSignalName_NegativeRound verifies negative round number produces valid string.
func TestRoundSignalName_NegativeRound(t *testing.T) {
	t.Parallel()
	got := RoundSignalName("claude", -1, "done")
	assert.Equal(t, "claude-round-1-done", got)
}

// TestRoundSignalName_SpecialChars verifies provider names with special characters
// are sanitized through the sanitizeProviderName path.
func TestRoundSignalName_SpecialChars(t *testing.T) {
	t.Parallel()
	got := RoundSignalName("../evil", 1, "done")
	assert.NotContains(t, got, "..")
}

// TestSetRoundEnv_NegativeRound verifies negative round sets correctly.
func TestSetRoundEnv_NegativeRound(t *testing.T) {
	SetRoundEnv(-1)
	got := os.Getenv("AUTOPUS_ROUND")
	assert.Equal(t, "-1", got)
	_ = os.Unsetenv("AUTOPUS_ROUND")
}

// TestCleanRoundSignals_ConcurrentAccess verifies concurrent cleanup is safe.
func TestCleanRoundSignals_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	sess, err := NewHookSession("test-clean-concurrent")
	require.NoError(t, err)
	defer sess.Cleanup()

	// Write done files for round 1.
	for _, p := range []string{"claude", "gemini"} {
		fname := p + "-round1-done"
		require.NoError(t, os.WriteFile(filepath.Join(sess.Dir(), fname), []byte("1"), 0o644))
	}

	// Run cleanup concurrently — should not panic.
	done := make(chan struct{}, 5)
	for i := 0; i < 5; i++ {
		go func() {
			CleanRoundSignals(sess, 1)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 5; i++ {
		<-done
	}
}
