package orchestra

import (
	"context"
	"time"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

const (
	// idleFallbackThreshold is how long 2-phase match must fail before trying idle fallback (R7).
	idleFallbackThreshold = 60 * time.Second
	// outputIdleThreshold is how long the output file must be unchanged to trigger idle completion (R7).
	outputIdleThreshold = 30 * time.Second
)

// waitForCompletion polls for completion using 2-phase consecutive match.
// R2: baseline prevents false positives from previous round's prompt.
// R7: When 2-phase match fails for idleFallbackThreshold, falls back to
// pipe-pane output file idle detection using isOutputIdle.
// Detects and auto-approves interactive tool permission prompts (gemini "Action Required").
// Suppresses idle fallback when provider shows working indicators ("Generating", "Working").
func waitForCompletion(ctx context.Context, term terminal.Terminal, pi paneInfo, patterns []CompletionPattern, baseline string) bool {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	candidateDetected := false
	// R7: Track when idle fallback becomes eligible
	idleFallbackStart := time.Now()

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
			// Auto-approve provider tool permission prompts (e.g., gemini "Action Required")
			if needsToolApproval(screen) {
				_ = term.SendCommand(ctx, pi.paneID, "1")
				_ = term.SendCommand(ctx, pi.paneID, "\n")
				candidateDetected = false
				continue
			}
			// R2: Screen unchanged from baseline — skip prompt matching to avoid
			// false positives from previous round's leftover prompt.
			// Still allow idle fallback to proceed (no continue).
			baselineMatch := baseline != "" && screen == baseline
			if baselineMatch {
				candidateDetected = false
			}
			if !baselineMatch && isPromptVisible(screen, patterns) {
				if candidateDetected {
					return true // Two consecutive matches — confirmed completion
				}
				candidateDetected = true // First match — wait for confirmation
			} else if !baselineMatch {
				candidateDetected = false // Reset — AI resumed output
			}
			// R7: Idle fallback — if 2-phase match hasn't succeeded within threshold,
			// check if output file is idle (no modifications for outputIdleThreshold).
			// Skip idle check if provider is actively working (thinking/generating).
			if pi.outputFile != "" && time.Since(idleFallbackStart) >= idleFallbackThreshold {
				if isOutputIdle(pi.outputFile, outputIdleThreshold) && !isProviderWorking(screen) {
					return true
				}
			}
		}
	}
}
