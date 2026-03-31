package orchestra

import (
	"context"
	"log"
	"time"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// SignalDetector uses cmux wait-for signals for instant completion detection.
// PRIMARY detector when cmux terminal is available.
// Integrates with SignalEmitter: on WaitForCompletion, a background emitter
// goroutine polls the pane and sends the cmux signal when completion is
// detected, allowing WaitForSignal to return immediately.
type SignalDetector struct {
	term    terminal.Terminal
	signal  terminal.SignalCapable
	emitter *SignalEmitter
}

// signalFallbackCap is the maximum time reserved for poll fallback if
// the emitter/signal path fails unexpectedly. Kept short because the
// emitter covers the normal path; this is a safety net only.
const signalFallbackCap = 10 * time.Second

// WaitForCompletion waits for the "done-{provider}" signal via cmux wait-for.
// Signal name format: "done-{provider}" for round 1, "done-{provider}-round{N}" for N > 1.
//
// A SignalEmitter goroutine polls the pane for completion patterns and sends
// the cmux signal when detected, which unblocks WaitForSignal instantly.
// The signal wait uses the parent context's remaining time (minus a small
// reserve for poll fallback), so it works even for slow providers.
//
// Falls back to ScreenPollDetector if the signal is not received in time
// (safety net for emitter failures).
func (d *SignalDetector) WaitForCompletion(ctx context.Context, pi paneInfo, patterns []CompletionPattern, baseline string, round int) (bool, error) {
	signalName := buildSignalName(pi.provider.Name, round)

	// Start emitter goroutine that polls screen and sends the signal.
	emitter := d.getOrCreateEmitter()
	emitter.Start(ctx, pi, patterns, baseline, round)
	defer emitter.Stop(pi.provider.Name, round)

	// Use the parent context's remaining time for signal wait, reserving
	// signalFallbackCap for poll fallback in case the emitter fails.
	signalTimeout := signalWaitTimeout(ctx)
	signalCtx, signalCancel := context.WithTimeout(ctx, signalTimeout)
	defer signalCancel()

	err := d.signal.WaitForSignal(signalCtx, signalName, signalTimeout)
	if err == nil {
		return true, nil // Signal received from emitter — instant completion
	}

	// If parent context is already done, no point in polling.
	if ctx.Err() != nil {
		return false, nil
	}

	// Fallback to screen polling with remaining time (safety net).
	log.Printf("[SignalDetector] signal wait failed for %s: %v -- falling back to poll", pi.provider.Name, err)
	fallback := &ScreenPollDetector{term: d.term}
	return fallback.WaitForCompletion(ctx, pi, patterns, baseline, round)
}

// signalWaitTimeout computes how long the signal wait should block.
// Uses the parent context's remaining time minus signalFallbackCap.
// Returns a minimum of 5s to avoid instant timeout.
func signalWaitTimeout(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 5 * time.Minute // No deadline — generous default
	}
	remaining := time.Until(deadline) - signalFallbackCap
	if remaining < 5*time.Second {
		remaining = 5 * time.Second
	}
	return remaining
}

// getOrCreateEmitter returns the detector's emitter, creating one if needed.
func (d *SignalDetector) getOrCreateEmitter() *SignalEmitter {
	if d.emitter == nil {
		d.emitter = NewSignalEmitter(d.term, d.signal)
	}
	return d.emitter
}
