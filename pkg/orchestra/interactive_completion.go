package orchestra

import (
	"context"
	"time"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// waitForCompletion polls for completion using 2-phase consecutive match.
// R2: baseline prevents false positives from previous round's prompt.
// If baseline is non-empty, prompt matching is skipped until the screen content
// changes from the baseline (indicating new AI output has appeared).
// @AX:NOTE [AUTO] REQ-3 — 2-phase consecutive match; idle detection disabled
func waitForCompletion(ctx context.Context, term terminal.Terminal, pi paneInfo, patterns []CompletionPattern, baseline string) bool {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	candidateDetected := false
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			screen, err := term.ReadScreen(ctx, pi.paneID, terminal.ReadScreenOpts{})
			if err != nil {
				candidateDetected = false
				continue
			}
			// R2: Screen unchanged from baseline — skip prompt matching to avoid
			// false positives from previous round's leftover prompt.
			if baseline != "" && screen == baseline {
				candidateDetected = false
				continue
			}
			if isPromptVisible(screen, patterns) {
				if candidateDetected {
					return true // Two consecutive matches — confirmed completion
				}
				candidateDetected = true // First match — wait for confirmation
			} else {
				candidateDetected = false // Reset — AI resumed output
			}
		}
	}
}
