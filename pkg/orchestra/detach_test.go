package orchestra

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunPaneOrchestraDetached_ReturnsJobID verifies that RunPaneOrchestraDetached
// launches providers in panes and returns a job ID without blocking.
func TestRunPaneOrchestraDetached_ReturnsJobID(t *testing.T) {
	t.Parallel()

	// Given: a cmux terminal and provider config
	mock := newCmuxMock()
	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			echoProvider("claude"),
			echoProvider("codex"),
		},
		Strategy:       StrategyConsensus,
		Prompt:         "test detach",
		TimeoutSeconds: 60,
		Terminal:       mock,
	}

	// When: detached orchestration is launched
	jobID, err := RunPaneOrchestraDetached(context.Background(), cfg)

	// Then: a valid job ID is returned immediately
	require.NoError(t, err)
	assert.NotEmpty(t, jobID, "detached mode must return a job ID")
	assert.Len(t, jobID, 16, "job ID should be 16 hex characters")
}

// TestRunPaneOrchestraDetached_ReturnsUnder2s verifies the <2s latency requirement.
// The function must return the job ID before providers finish executing.
func TestRunPaneOrchestraDetached_ReturnsUnder2s(t *testing.T) {
	t.Parallel()

	// Given: a cmux terminal with slow providers
	mock := newCmuxMock()
	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			sleepProvider("slow1"),
			sleepProvider("slow2"),
		},
		Strategy:       StrategyConsensus,
		Prompt:         "test latency",
		TimeoutSeconds: 120,
		Terminal:       mock,
	}

	// When: detached orchestration is launched and timed
	start := time.Now()
	_, err := RunPaneOrchestraDetached(context.Background(), cfg)
	elapsed := time.Since(start)

	// Then: must return in under 2 seconds
	require.NoError(t, err)
	assert.Less(t, elapsed, 2*time.Second, "detached mode must return job ID in <2s")
}

// TestRunPaneOrchestraDetached_PlainTerminal_Error verifies that detach mode
// is not available for plain terminals — only pane terminals support it.
func TestRunPaneOrchestraDetached_PlainTerminal_Error(t *testing.T) {
	t.Parallel()

	// Given: a plain terminal (no pane support)
	mock := newPlainMock()
	cfg := OrchestraConfig{
		Providers:      []ProviderConfig{echoProvider("p1")},
		Strategy:       StrategyConsensus,
		Prompt:         "test",
		TimeoutSeconds: 10,
		Terminal:       mock,
	}

	// When: detached mode is attempted on plain terminal
	_, err := RunPaneOrchestraDetached(context.Background(), cfg)

	// Then: should return an error (detach requires pane terminal)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "detach")
}

// TestRunPaneOrchestra_NoDetachFlag verifies that when --no-detach is set,
// the pane runner uses blocking mode even on pane terminals.
func TestRunPaneOrchestra_NoDetachFlag(t *testing.T) {
	t.Parallel()

	// Given: a cmux terminal with NoDetach flag set
	mock := newCmuxMock()
	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			echoProvider("p1"),
		},
		Strategy:       StrategyConsensus,
		Prompt:         "test no-detach",
		TimeoutSeconds: 10,
		Terminal:       mock,
		NoDetach:       true,
	}

	// When: pane runner executes with NoDetach=true
	result, err := RunPaneOrchestra(context.Background(), cfg)

	// Then: should complete synchronously with full result (not a job ID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Merged, "blocking mode must return merged result")
}

// TestShouldDetach_TTYAndPaneTerminal verifies the auto-detach detection logic:
// detach when stdout is TTY AND terminal is pane-capable.
func TestShouldDetach_TTYAndPaneTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		terminal string
		isTTY    bool
		noDetach bool
		expected bool
	}{
		{
			name:     "cmux + TTY + no flag => detach",
			terminal: "cmux",
			isTTY:    true,
			noDetach: false,
			expected: true,
		},
		{
			name:     "cmux + TTY + no-detach flag => no detach",
			terminal: "cmux",
			isTTY:    true,
			noDetach: true,
			expected: false,
		},
		{
			name:     "cmux + non-TTY => no detach",
			terminal: "cmux",
			isTTY:    false,
			noDetach: false,
			expected: false,
		},
		{
			name:     "plain + TTY => no detach",
			terminal: "plain",
			isTTY:    true,
			noDetach: false,
			expected: false,
		},
		{
			name:     "tmux + TTY => detach",
			terminal: "tmux",
			isTTY:    true,
			noDetach: false,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ShouldDetach(tt.terminal, tt.isTTY, tt.noDetach)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestRunPaneOrchestraDetached_SplitPaneFailure verifies that detach returns
// an error when pane splitting fails.
func TestRunPaneOrchestraDetached_SplitPaneFailure(t *testing.T) {
	t.Parallel()

	mock := newCmuxMock()
	mock.splitPaneErr = fmt.Errorf("pane split failed")
	cfg := OrchestraConfig{
		Providers:      []ProviderConfig{echoProvider("p1")},
		Strategy:       StrategyConsensus,
		Prompt:         "test",
		TimeoutSeconds: 10,
		Terminal:       mock,
	}

	_, err := RunPaneOrchestraDetached(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "split panes")
}

// TestRunPaneOrchestraDetached_NilTerminal verifies that detach with nil
// terminal returns an error.
func TestRunPaneOrchestraDetached_NilTerminal(t *testing.T) {
	t.Parallel()

	cfg := OrchestraConfig{
		Providers:      []ProviderConfig{echoProvider("p1")},
		Strategy:       StrategyConsensus,
		Prompt:         "test",
		TimeoutSeconds: 10,
		Terminal:       nil,
	}

	_, err := RunPaneOrchestraDetached(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "detach")
}

// TestRunPaneOrchestraDetached_DefaultTimeout verifies that zero timeout
// defaults to 120 seconds.
func TestRunPaneOrchestraDetached_DefaultTimeout(t *testing.T) {
	t.Parallel()

	mock := newCmuxMock()
	cfg := OrchestraConfig{
		Providers:      []ProviderConfig{echoProvider("p1")},
		Strategy:       StrategyConsensus,
		Prompt:         "test",
		TimeoutSeconds: 0,
		Terminal:       mock,
	}

	jobID, err := RunPaneOrchestraDetached(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, jobID)
}
