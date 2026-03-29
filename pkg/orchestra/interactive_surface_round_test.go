package orchestra

import (
	"context"
	"testing"
	"time"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// TestExecuteRound_Round2_SurfaceValidation verifies that executeRound in round 2
// checks surface validity and recreates stale panes.
func TestExecuteRound_Round2_SurfaceValidation(t *testing.T) {
	mock := &surfaceMock{
		mockTerminal: mockTerminal{name: "cmux", readScreenOutput: "Ask anything"},
		stalePane:    map[terminal.PaneID]bool{"stale-pane": true},
	}

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			{Name: "opencode", Binary: "opencode"},
		},
		Strategy:       StrategyDebate,
		Prompt:         "round 2 test",
		TimeoutSeconds: 5,
		Terminal:       mock,
		Interactive:    true,
		InitialDelay:   time.Millisecond,
	}
	panes := []paneInfo{{
		provider: cfg.Providers[0],
		paneID:   "stale-pane",
	}}
	prevResponses := []ProviderResponse{{Provider: "claude", Output: "round 1 answer"}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = executeRound(ctx, cfg, panes, nil, 2, prevResponses)

	// After surface recreation, pane ID should have changed from "stale-pane".
	if panes[0].paneID == "stale-pane" {
		t.Error("expected pane to be recreated after stale surface detection")
	}
}

// TestExecuteRound_Round2_ClaudeSkipsSurfaceCheck verifies that claude providers
// skip surface validation even in round > 1 (R5).
func TestExecuteRound_Round2_ClaudeSkipsSurfaceCheck(t *testing.T) {
	mock := &surfaceMock{
		mockTerminal: mockTerminal{name: "cmux", readScreenOutput: "❯\n"},
		stalePane:    map[terminal.PaneID]bool{"claude-pane": true},
	}

	cfg := OrchestraConfig{
		Providers: []ProviderConfig{
			{Name: "claude", Binary: "claude"},
		},
		Strategy:       StrategyDebate,
		Prompt:         "round 2 claude test",
		TimeoutSeconds: 5,
		Terminal:       mock,
		Interactive:    true,
		InitialDelay:   time.Millisecond,
	}
	panes := []paneInfo{{
		provider: cfg.Providers[0],
		paneID:   "claude-pane",
	}}
	prevResponses := []ProviderResponse{{Provider: "opencode", Output: "round 1 answer"}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = executeRound(ctx, cfg, panes, nil, 2, prevResponses)

	// Claude should NOT have been recreated despite stale surface.
	if panes[0].paneID != "claude-pane" {
		t.Error("claude should skip surface check — paneID should be unchanged")
	}
}
