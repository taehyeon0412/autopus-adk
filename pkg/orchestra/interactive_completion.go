package orchestra

import (
	"context"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// waitForCompletion delegates to the best CompletionDetector for the terminal.
// The round parameter enables round-scoped signal names in debate mode
// (e.g., "done-gemini-round2"). Pass 0 for non-debate strategies.
func waitForCompletion(ctx context.Context, term terminal.Terminal, pi paneInfo, patterns []CompletionPattern, baseline string, round int) bool {
	detector := NewCompletionDetector(term)
	completed, _ := detector.WaitForCompletion(ctx, pi, patterns, baseline, round)
	return completed
}
