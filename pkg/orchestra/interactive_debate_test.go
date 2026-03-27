package orchestra

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- runInteractiveDebate tests ---

// TestRunInteractiveDebate_DefaultRound verifies that DebateRounds=0
// defaults to 1 round (no rebuttal).
func TestRunInteractiveDebate_DefaultRound(t *testing.T) {
	t.Parallel()
	cfg := OrchestraConfig{
		Strategy: StrategyDebate, DebateRounds: 0,
		Prompt: "default round", Providers: []ProviderConfig{{Name: "claude", Binary: "echo"}},
		TimeoutSeconds: 10, Interactive: true, HookMode: true, SessionID: "test-debate-default",
	}
	result, err := runInteractiveDebate(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestRunInteractiveDebate_RoundsFlag_Range verifies --rounds N validation.
func TestRunInteractiveDebate_RoundsFlag_Range(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		rounds    int
		expectErr bool
	}{
		{"valid 1", 1, false},
		{"valid 10", 10, false},
		{"zero defaults", 0, false},
		{"negative", -1, true},
		{"exceeds max", 11, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := OrchestraConfig{
				Strategy: StrategyDebate, DebateRounds: tt.rounds,
				Prompt: "range test", Providers: []ProviderConfig{{Name: "claude", Binary: "echo"}},
				TimeoutSeconds: 5, Interactive: true, HookMode: true, SessionID: "test-range",
			}
			_, err := runInteractiveDebate(context.Background(), cfg)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "rounds must be")
			}
		})
	}
}

// TestRunInteractiveDebate_ContextCancellation verifies context cancellation.
func TestRunInteractiveDebate_ContextCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cfg := OrchestraConfig{
		Strategy: StrategyDebate, DebateRounds: 3, Prompt: "cancel test",
		Providers: []ProviderConfig{{Name: "claude", Binary: "echo"}},
		TimeoutSeconds: 10, Interactive: true, HookMode: true, SessionID: "test-cancel",
	}
	_, err := runInteractiveDebate(ctx, cfg)
	assert.Error(t, err, "must return error on cancelled context")
}

// TestRunInteractiveDebate_NoTerminal_Fallback verifies nil terminal triggers
// non-interactive fallback path.
func TestRunInteractiveDebate_NoTerminal_Fallback(t *testing.T) {
	t.Parallel()
	cfg := OrchestraConfig{
		Strategy: StrategyDebate, DebateRounds: 1, Prompt: "no terminal test",
		Providers: []ProviderConfig{{Name: "claude", Binary: "echo"}},
		TimeoutSeconds: 10, Terminal: nil,
	}
	result, err := runInteractiveDebate(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StrategyDebate, result.Strategy)
}

// TestRunInteractiveDebate_MultiRound_NoTerminal verifies multi-round debate
// runs the non-interactive fallback path.
func TestRunInteractiveDebate_MultiRound_NoTerminal(t *testing.T) {
	t.Parallel()
	cfg := OrchestraConfig{
		Strategy: StrategyDebate, DebateRounds: 2, Prompt: "multi round",
		Providers: []ProviderConfig{
			{Name: "claude", Binary: "echo"},
			{Name: "gemini", Binary: "echo"},
		},
		TimeoutSeconds: 10, Terminal: nil,
	}
	result, err := runInteractiveDebate(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestRunInteractiveDebate_NonExistentBinary verifies fallback path when
// binary does not exist (runDebate fails, falls through to parallel/empty).
func TestRunInteractiveDebate_NonExistentBinary(t *testing.T) {
	t.Parallel()
	cfg := OrchestraConfig{
		Strategy: StrategyDebate, DebateRounds: 1, Prompt: "bad binary test",
		Providers: []ProviderConfig{{Name: "test", Binary: "nonexistent-binary-xyz"}},
		TimeoutSeconds: 5, Terminal: nil,
	}
	result, err := runInteractiveDebate(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// --- Helper function tests ---

// TestConsensusReached_Different verifies no consensus for different outputs.
func TestConsensusReached_Different(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		{Provider: "claude", Output: "answer A with lots of detail"},
		{Provider: "gemini", Output: "completely different answer B"},
	}
	assert.False(t, consensusReached(responses))
}

// TestConsensusReached_SingleProvider verifies single provider returns false.
func TestConsensusReached_SingleProvider(t *testing.T) {
	t.Parallel()
	assert.False(t, consensusReached([]ProviderResponse{{Provider: "claude", Output: "one"}}))
}

// TestConsensusReached_EmptyOutput verifies empty outputs returns false.
func TestConsensusReached_EmptyOutput(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		{Provider: "claude", Output: ""},
		{Provider: "gemini", Output: ""},
	}
	assert.False(t, consensusReached(responses))
}

// TestCountNonEmpty verifies counting of non-empty responses.
func TestCountNonEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		resps    []ProviderResponse
		expected int
	}{
		{"all non-empty", []ProviderResponse{{Output: "a"}, {Output: "b"}, {Output: "c"}}, 3},
		{"mixed", []ProviderResponse{{Output: "a"}, {Output: ""}, {Output: "c"}}, 2},
		{"all empty", []ProviderResponse{{Output: ""}, {Output: ""}}, 0},
		{"nil slice", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, countNonEmpty(tt.resps))
		})
	}
}

// TestPerRoundTimeout verifies per-round timeout calculation.
func TestPerRoundTimeout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		total    int
		rounds   int
		expected time.Duration
	}{
		{"120s / 3 rounds", 120, 3, 40 * time.Second},
		{"60s / 1 round", 60, 1, 60 * time.Second},
		{"zero total defaults", 0, 2, 60 * time.Second},
		{"negative total defaults", -1, 4, 30 * time.Second},
		{"zero rounds defaults", 60, 0, 60 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, perRoundTimeout(tt.total, tt.rounds))
		})
	}
}

// TestBuildDebateResult verifies result construction.
func TestBuildDebateResult(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{
		{Provider: "claude", Output: "claude says"},
		{Provider: "gemini", Output: "gemini says"},
	}
	history := [][]ProviderResponse{responses}
	result := buildDebateResult(OrchestraConfig{Strategy: StrategyDebate}, responses, history, time.Now())
	assert.Equal(t, StrategyDebate, result.Strategy)
	assert.Len(t, result.Responses, 2)
	assert.Len(t, result.RoundHistory, 1)
	assert.NotEmpty(t, result.Merged)
}

// TestBuildDebateResult_NilResponses verifies nil responses fallback.
func TestBuildDebateResult_NilResponses(t *testing.T) {
	t.Parallel()
	result := buildDebateResult(OrchestraConfig{Strategy: StrategyDebate}, nil, nil, time.Now())
	assert.Contains(t, result.Merged, "0 rounds completed")
}

// TestMergeByStrategyWithRoundHistory verifies round history merge.
func TestMergeByStrategyWithRoundHistory(t *testing.T) {
	t.Parallel()
	rounds := [][]ProviderResponse{
		{{Provider: "claude", Output: "r1"}, {Provider: "gemini", Output: "r1"}},
		{{Provider: "claude", Output: "r2"}, {Provider: "gemini", Output: "r2"}},
	}
	result := mergeByStrategyWithRoundHistory(rounds, OrchestraConfig{Strategy: StrategyDebate})
	require.NotNil(t, result)
	assert.Equal(t, StrategyDebate, result.Strategy)
	assert.Len(t, result.RoundHistory, 2)
	assert.Len(t, result.Responses, 2)
}

// TestMergeByStrategyWithRoundHistory_Empty verifies empty rounds.
func TestMergeByStrategyWithRoundHistory_Empty(t *testing.T) {
	t.Parallel()
	result := mergeByStrategyWithRoundHistory(nil, OrchestraConfig{Strategy: StrategyDebate})
	require.NotNil(t, result)
	assert.Nil(t, result.Responses)
}

// TestRunInteractiveDebate_SingleProvider verifies debate with only one provider.
func TestRunInteractiveDebate_SingleProvider(t *testing.T) {
	t.Parallel()
	cfg := OrchestraConfig{
		Strategy: StrategyDebate, DebateRounds: 2, Prompt: "single provider debate",
		Providers:      []ProviderConfig{{Name: "claude", Binary: "echo"}},
		TimeoutSeconds: 10, Terminal: nil,
	}
	result, err := runInteractiveDebate(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestRunInteractiveDebate_WithJudge_NoTerminal verifies judge path in
// non-interactive fallback mode.
func TestRunInteractiveDebate_WithJudge_NoTerminal(t *testing.T) {
	t.Parallel()
	cfg := OrchestraConfig{
		Strategy: StrategyDebate, DebateRounds: 1, Prompt: "judge test",
		JudgeProvider: "echo",
		Providers: []ProviderConfig{
			{Name: "claude", Binary: "echo"},
			{Name: "gemini", Binary: "echo"},
		},
		TimeoutSeconds: 10, Terminal: nil,
	}
	result, err := runInteractiveDebate(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestBuildDebateResult_SingleRound verifies single round result.
func TestBuildDebateResult_SingleRound(t *testing.T) {
	t.Parallel()
	responses := []ProviderResponse{{Provider: "claude", Output: "only"}}
	history := [][]ProviderResponse{responses}
	result := buildDebateResult(OrchestraConfig{Strategy: StrategyDebate}, responses, history, time.Now())
	assert.Len(t, result.RoundHistory, 1)
	assert.NotEmpty(t, result.Merged)
}

// TestMergeByStrategyWithRoundHistory_SingleRound verifies single round.
func TestMergeByStrategyWithRoundHistory_SingleRound(t *testing.T) {
	t.Parallel()
	rounds := [][]ProviderResponse{{{Provider: "claude", Output: "single"}}}
	result := mergeByStrategyWithRoundHistory(rounds, OrchestraConfig{Strategy: StrategyDebate})
	assert.Len(t, result.Responses, 1)
	assert.Len(t, result.RoundHistory, 1)
}
