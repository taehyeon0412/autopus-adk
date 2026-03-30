package orchestra

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// SurfaceManager monitors surface health in the background and provides
// proactive stale detection. Replaces the reactive validateSurface() approach.
// @AX:ANCHOR [AUTO] coordinator — owns background goroutine, health cache, and pane recovery; used by runPaneDebate and executeRound
type SurfaceManager struct {
	term     terminal.Terminal
	signal   terminal.SignalCapable // nil if terminal doesn't support signals
	interval time.Duration          // health check interval (default 5s)

	mu     sync.RWMutex
	health map[string]terminal.SurfaceStatus // paneID -> last known health
	cancel context.CancelFunc
}

// NewSurfaceManager creates a SurfaceManager. If the terminal supports
// SignalCapable, uses surface-health checks; otherwise falls back to no-op.
func NewSurfaceManager(term terminal.Terminal) *SurfaceManager {
	sm := &SurfaceManager{
		term:     term,
		interval: 5 * time.Second,
		health:   make(map[string]terminal.SurfaceStatus),
	}
	if sc, ok := term.(terminal.SignalCapable); ok {
		sm.signal = sc
	}
	return sm
}

// Start begins background health monitoring for the given panes.
// Call Stop() when the debate ends to clean up the goroutine.
func (sm *SurfaceManager) Start(ctx context.Context, panes []paneInfo) {
	if sm.signal == nil {
		return // No signal support -- skip monitoring
	}
	monCtx, cancel := context.WithCancel(ctx)
	sm.cancel = cancel

	go sm.monitorLoop(monCtx, panes)
}

// Stop stops the background monitoring goroutine.
func (sm *SurfaceManager) Stop() {
	if sm.cancel != nil {
		sm.cancel()
	}
}

// @AX:TODO [AUTO] P1 warm pool (R9) — pre-create spare panes for instant recovery instead of on-demand recreatePane
// IsHealthy returns true if the pane's last known health status is valid.
// Returns true by default (optimistic) if no health data is available yet.
func (sm *SurfaceManager) IsHealthy(paneID terminal.PaneID) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	status, exists := sm.health[string(paneID)]
	if !exists {
		return true // Optimistic: no data yet means assume healthy
	}
	return status.Valid
}

// monitorLoop polls surface health at the configured interval.
func (sm *SurfaceManager) monitorLoop(ctx context.Context, panes []paneInfo) {
	ticker := time.NewTicker(sm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sm.checkAll(ctx, panes)
		}
	}
}

// checkAll checks health of all panes and updates the cache.
func (sm *SurfaceManager) checkAll(ctx context.Context, panes []paneInfo) {
	for _, pi := range panes {
		if pi.skipWait {
			continue
		}
		status, err := sm.signal.SurfaceHealth(ctx, pi.paneID)
		if err != nil {
			// On error, mark as unhealthy
			status = terminal.SurfaceStatus{Valid: false}
		}
		sm.mu.Lock()
		sm.health[string(pi.paneID)] = status
		sm.mu.Unlock()
	}
}

// ValidateAndRecover checks surface health and recreates stale panes.
// Returns updated paneInfo and whether recovery occurred.
// This replaces the inline surface validation logic in executeRound().
func (sm *SurfaceManager) ValidateAndRecover(ctx context.Context, cfg OrchestraConfig, pi paneInfo, round int) (paneInfo, bool, error) {
	healthy := sm.IsHealthy(pi.paneID)
	if healthy {
		// Double-check with live ReadScreen if signal says healthy
		if !validateSurface(ctx, cfg.Terminal, pi.paneID) {
			healthy = false
		}
	}
	if healthy {
		return pi, false, nil // No recovery needed
	}

	// Surface is stale -- recreate
	log.Printf("[SurfaceManager] %s surface stale, recreating", pi.provider.Name)
	newPI, err := recreatePane(ctx, cfg, pi, round)
	if err != nil {
		return pi, false, err
	}

	// Update health cache with new pane
	sm.mu.Lock()
	delete(sm.health, string(pi.paneID))
	sm.health[string(newPI.paneID)] = terminal.SurfaceStatus{Valid: true}
	sm.mu.Unlock()

	return newPI, true, nil
}

// captureBaselines reads the current screen content for all active panes.
// Used to establish baseline for false-positive completion detection prevention (R7).
func captureBaselines(ctx context.Context, term terminal.Terminal, panes []paneInfo) map[string]string {
	baselines := make(map[string]string)
	for _, pi := range panes {
		if pi.skipWait {
			continue
		}
		screen, _ := term.ReadScreen(ctx, pi.paneID, terminal.ReadScreenOpts{})
		baselines[pi.provider.Name] = screen
	}
	return baselines
}

// sendPromptWithRetry sends a prompt to a pane, recreating the pane on failure.
// Returns updated paneInfo and whether recreation occurred.
func sendPromptWithRetry(ctx context.Context, cfg OrchestraConfig, pi paneInfo, prompt string, round int, baselines map[string]string) (paneInfo, bool, error) {
	// Try SendLongText
	if err := cfg.Terminal.SendLongText(ctx, pi.paneID, prompt); err == nil {
		return pi, false, nil
	}

	// Recreation + retry logic (moved from executeRound)
	log.Printf("[Round %d] %s SendLongText failed -- attempting recreation", round, pi.provider.Name)
	newPI, err := recreatePane(ctx, cfg, pi, round)
	if err != nil {
		return pi, false, fmt.Errorf("recreatePane failed: %w", err)
	}

	// Refresh baseline for the new pane
	if screen, rerr := cfg.Terminal.ReadScreen(ctx, newPI.paneID, terminal.ReadScreenOpts{}); rerr == nil {
		baselines[pi.provider.Name] = screen
	}

	// @AX:NOTE [AUTO] magic constant 3 — max retry attempts; paired with 2s exponential backoff
	// Retry with exponential backoff — longer delays give cmux surfaces time to stabilize
	for attempt := 1; attempt <= 3; attempt++ {
		if retryErr := cfg.Terminal.SendLongText(ctx, newPI.paneID, prompt); retryErr == nil {
			return newPI, true, nil
		}
		wait := time.Duration(attempt) * 2 * time.Second
		log.Printf("[Round %d] %s SendLongText attempt %d failed, waiting %v...", round, pi.provider.Name, attempt, wait)
		time.Sleep(wait)
	}
	return newPI, true, fmt.Errorf("SendLongText failed after 3 attempts")
}
