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

// signalMaxWait caps how long the signal detector waits before falling back to polling.
// This prevents the signal detector from consuming the entire round timeout when
// provider hooks are not configured to send completion signals.
const signalMaxWait = 15 * time.Second

// WaitForCompletion waits for the "done-{provider}" signal via cmux wait-for.
// Signal name format: "done-{provider}" for round 1, "done-{provider}-round{N}" for N > 1.
// The signal wait is capped at signalMaxWait (15s) to preserve time for poll fallback.
// Falls back to ScreenPollDetector if signal wait fails or times out.
// @AX:WARN [AUTO] silent fallback — signal failure degrades to polling without caller notification
func (d *SignalDetector) WaitForCompletion(ctx context.Context, pi paneInfo, patterns []CompletionPattern, baseline string, round int) (bool, error) {
	// Build signal name
	signalName := fmt.Sprintf("done-%s", sanitizeProviderName(pi.provider.Name))
	if round > 1 {
		signalName = fmt.Sprintf("done-%s-round%d", sanitizeProviderName(pi.provider.Name), round)
	}

	// Cap signal wait at signalMaxWait so poll fallback gets the remaining time.
	// Without this cap, signal detection consumes the entire round timeout and
	// the poll detector gets only milliseconds — causing false positives.
	signalCtx, signalCancel := context.WithTimeout(ctx, signalMaxWait)
	defer signalCancel()

	// Try signal-based detection first
	err := d.signal.WaitForSignal(signalCtx, signalName, signalMaxWait)
	if err == nil {
		return true, nil // Signal received -- immediate completion
	}

	// Fallback to screen polling on signal failure
	log.Printf("[SignalDetector] signal wait failed for %s: %v -- falling back to poll", pi.provider.Name, err)
	fallback := &ScreenPollDetector{term: d.term}
	return fallback.WaitForCompletion(ctx, pi, patterns, baseline, round)
}
