package orchestra

import (
	"context"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// waitForCompletion is a compatibility wrapper that delegates to CompletionDetector.
// Callers (waitAndCollectResults, etc.) continue using the same signature.
// @AX:NOTE [AUTO] thin wrapper — preserves old call signature; round=0 means auto-detect
func waitForCompletion(ctx context.Context, term terminal.Terminal, pi paneInfo, patterns []CompletionPattern, baseline string) bool {
	detector := NewCompletionDetector(term)
	completed, _ := detector.WaitForCompletion(ctx, pi, patterns, baseline, 0)
	return completed
}
