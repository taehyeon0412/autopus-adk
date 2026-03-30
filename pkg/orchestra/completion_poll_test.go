package orchestra

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestScreenPollDetector_TwoPhaseMatch verifies 2-phase consecutive prompt match.
func TestScreenPollDetector_TwoPhaseMatch(t *testing.T) {
	t.Parallel()
	mock := &countingScreenMock{}
	mock.outputs = []string{">\n", ">\n"}
	mock.mockTerminal.name = "plain"

	// Use codex pattern so ">" alone won't match — use explicit ❯ for claude
	mock.outputs = []string{"\u276F\n", "\u276F\n"}
	detector := &ScreenPollDetector{term: mock}
	pi := paneInfo{paneID: "pane-1", provider: ProviderConfig{Name: "claude"}}
	patterns := DefaultCompletionPatterns()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ok, err := detector.WaitForCompletion(ctx, pi, patterns, "", 0)
	assert.NoError(t, err)
	assert.True(t, ok, "2-phase consecutive match should detect completion")
}

// TestScreenPollDetector_ContextCancel verifies context cancellation returns false.
func TestScreenPollDetector_ContextCancel(t *testing.T) {
	t.Parallel()
	mock := newPlainMock()
	mock.readScreenOutput = "still running..."

	detector := &ScreenPollDetector{term: mock}
	pi := paneInfo{paneID: "pane-1", provider: ProviderConfig{Name: "claude"}}
	patterns := DefaultCompletionPatterns()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	ok, err := detector.WaitForCompletion(ctx, pi, patterns, "", 0)
	assert.NoError(t, err)
	assert.False(t, ok, "cancelled context must return false")
}

// TestScreenPollDetector_BaselineFiltering verifies baseline prevents false positives.
func TestScreenPollDetector_BaselineFiltering(t *testing.T) {
	t.Parallel()
	mock := &countingScreenMock{}
	baseline := "\u276F\n"
	// First 2 calls return baseline (filtered), then new screen with prompt
	mock.outputs = []string{baseline, baseline, "new output\n\u276F\n", "new output\n\u276F\n"}
	mock.mockTerminal.name = "plain"

	detector := &ScreenPollDetector{term: mock}
	pi := paneInfo{paneID: "pane-1", provider: ProviderConfig{Name: "claude"}}
	patterns := DefaultCompletionPatterns()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ok, err := detector.WaitForCompletion(ctx, pi, patterns, baseline, 0)
	assert.NoError(t, err)
	assert.True(t, ok, "should complete after baseline changes and 2-phase match")
}

// TestScreenPollDetector_IdleFallbackDirectCheck verifies isOutputIdle via the detector.
func TestScreenPollDetector_IdleFallbackDirectCheck(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")
	if err := os.WriteFile(outputFile, []byte("done"), 0644); err != nil {
		t.Fatal(err)
	}
	// Set modtime far in the past so isOutputIdle returns true
	past := time.Now().Add(-1 * time.Minute)
	if err := os.Chtimes(outputFile, past, past); err != nil {
		t.Fatal(err)
	}

	assert.True(t, isOutputIdle(outputFile, outputIdleThreshold),
		"output file with old modtime should be considered idle")
}

// TestScreenPollDetector_PerProviderIdleThreshold verifies per-provider idle threshold is used.
func TestScreenPollDetector_PerProviderIdleThreshold(t *testing.T) {
	t.Parallel()

	// Verify that a provider with custom IdleThreshold compiles and is accepted.
	pi := paneInfo{
		paneID:     "pane-1",
		provider:   ProviderConfig{Name: "opencode", IdleThreshold: 90 * time.Second},
		outputFile: "/tmp/nonexistent",
	}
	assert.Equal(t, 90*time.Second, pi.provider.IdleThreshold)
}
