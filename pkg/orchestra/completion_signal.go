package orchestra

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// SignalDetector uses cmux wait-for signals for instant completion detection.
// PRIMARY detector when cmux terminal is available.
type SignalDetector struct {
	term   terminal.Terminal
	signal terminal.SignalCapable
}

// WaitForCompletion waits for the "done-{provider}" signal via cmux wait-for.
// Signal name format: "done-{provider}" for round 1, "done-{provider}-round{N}" for N > 1.
// Falls back to ScreenPollDetector if signal wait fails or times out.
// @AX:WARN [AUTO] silent fallback — signal failure degrades to polling without caller notification
func (d *SignalDetector) WaitForCompletion(ctx context.Context, pi paneInfo, patterns []CompletionPattern, baseline string, round int) (bool, error) {
	// Build signal name
	signalName := fmt.Sprintf("done-%s", sanitizeProviderName(pi.provider.Name))
	if round > 1 {
		signalName = fmt.Sprintf("done-%s-round%d", sanitizeProviderName(pi.provider.Name), round)
	}

	// Calculate timeout from context
	deadline, ok := ctx.Deadline()
	timeout := 60 * time.Second
	if ok {
		timeout = time.Until(deadline)
	}

	// Try signal-based detection first
	err := d.signal.WaitForSignal(ctx, signalName, timeout)
	if err == nil {
		return true, nil // Signal received -- immediate completion
	}

	// Fallback to screen polling on signal failure
	log.Printf("[SignalDetector] signal wait failed for %s: %v -- falling back to poll", pi.provider.Name, err)
	fallback := &ScreenPollDetector{term: d.term}
	return fallback.WaitForCompletion(ctx, pi, patterns, baseline, round)
}
