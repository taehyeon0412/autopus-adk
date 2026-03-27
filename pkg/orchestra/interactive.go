package orchestra

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// RunInteractivePaneOrchestra runs orchestration using interactive CLI sessions in terminal panes.
// Each provider's CLI binary is launched as an interactive session, prompts are sent via SendCommand,
// and completion is detected via ReadScreen polling and pipe-pane idle detection.
// Falls back to sentinel-based RunPaneOrchestra if interactive mode fails (R8).
// @AX:NOTE [AUTO] interactive orchestration entry point — fan_in=1 (pane_runner.go only); downgraded from ANCHOR
func RunInteractivePaneOrchestra(ctx context.Context, cfg OrchestraConfig) (*OrchestraResult, error) {
	// R8: plain terminal -> fallback to sentinel mode
	if cfg.Terminal == nil || cfg.Terminal.Name() == "plain" {
		cfg.Interactive = false
		return RunPaneOrchestra(ctx, cfg)
	}

	start := time.Now()
	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 120
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Hook mode: create session for file-based result collection
	var hookSession *HookSession
	if cfg.HookMode {
		var hsErr error
		hookSession, hsErr = NewHookSession(cfg.SessionID)
		if hsErr != nil {
			// R8: fallback to non-hook mode
			cfg.HookMode = false
		} else {
			defer hookSession.Cleanup()
			_ = os.Setenv("AUTOPUS_SESSION_ID", cfg.SessionID)
		}
	}

	// Step 1: Split panes (reuse existing splitProviderPanes)
	panes, failed, err := splitProviderPanes(timeoutCtx, cfg)
	if err != nil {
		// R8: fallback on split failure
		cfg.Interactive = false
		return RunPaneOrchestra(ctx, cfg)
	}
	defer cleanupInteractivePanes(cfg.Terminal, panes)

	// Step 2: Start pipe capture for each pane
	if err := startPipeCapture(timeoutCtx, cfg.Terminal, panes); err != nil {
		cfg.Interactive = false
		return RunPaneOrchestra(ctx, cfg)
	}

	// Step 3: Launch interactive sessions (send binary name to each pane)
	launchFailed := launchInteractiveSessions(timeoutCtx, cfg, panes)
	failed = append(failed, launchFailed...)

	// Step 4: Wait for sessions to be ready (prompt visible)
	waitForSessionReady(timeoutCtx, cfg.Terminal, panes)

	// Step 5: Send prompts to each session
	promptFailed := sendPrompts(timeoutCtx, cfg, panes)
	failed = append(failed, promptFailed...)

	// Step 5.5: Wait for AI to start processing before completion detection.
	// Without this delay, the prompt pattern on the current screen triggers
	// immediate "completion" before the AI even begins responding.
	time.Sleep(15 * time.Second)

	// Step 6-7: Wait for completion and collect results
	patterns := DefaultCompletionPatterns()
	var responses []ProviderResponse
	if cfg.HookMode && hookSession != nil {
		// R5: Hook-based collection
		var hookErr error
		responses, hookErr = WaitAndCollectHookResults(cfg, cfg.SessionID)
		if hookErr != nil {
			// R8: fallback to ReadScreen-based collection
			responses = waitAndCollectResults(timeoutCtx, cfg, panes, patterns, start)
		}
	} else {
		// Original ReadScreen-based collection
		responses = waitAndCollectResults(timeoutCtx, cfg, panes, patterns, start)
	}

	// Step 8: Merge by strategy (reuse existing mergeByStrategy)
	total := time.Since(start)
	merged, summary := mergeByStrategy(cfg.Strategy, responses, cfg)
	if merged == "" {
		merged = fmt.Sprintf("[interactive mode] %d providers executed", len(responses))
	}

	return &OrchestraResult{
		Strategy:        cfg.Strategy,
		Responses:       responses,
		Merged:          merged,
		Duration:        total,
		Summary:         summary,
		FailedProviders: failed,
	}, nil
}

// startPipeCapture starts pipe-pane output streaming for each pane.
func startPipeCapture(ctx context.Context, term terminal.Terminal, panes []paneInfo) error {
	for _, pi := range panes {
		if err := term.PipePaneStart(ctx, pi.paneID, pi.outputFile); err != nil {
			return fmt.Errorf("pipe-pane start for %s: %w", pi.provider.Name, err)
		}
	}
	return nil
}

// launchInteractiveSessions sends the provider binary name to each pane to start an interactive session.
// In interactive mode, we launch the CLI binary with model flags to get a real CLI session.
// The user prompt will be sent separately via sendPrompts() after the session is ready.
func launchInteractiveSessions(ctx context.Context, cfg OrchestraConfig, panes []paneInfo) []FailedProvider {
	var failed []FailedProvider
	for i, pi := range panes {
		// Build launch command: binary + interactive args (model flags, etc.)
		// PaneArgs without print/pipe flags are used; prompt is sent separately
		cmd := buildInteractiveLaunchCmd(pi.provider) + "\n"
		if err := cfg.Terminal.SendCommand(ctx, pi.paneID, cmd); err != nil {
			failed = append(failed, FailedProvider{
				Name:  pi.provider.Name,
				Error: fmt.Sprintf("launch session failed: %v", err),
			})
			panes[i].skipWait = true
		}
	}
	return failed
}

// waitForSessionReady polls ReadScreen until a prompt is visible or timeout.
func waitForSessionReady(ctx context.Context, term terminal.Terminal, panes []paneInfo) {
	patterns := DefaultCompletionPatterns()
	for _, pi := range panes {
		if pi.skipWait {
			continue
		}
		pollUntilPrompt(ctx, term, pi.paneID, patterns, 30*time.Second)
	}
}

// pollUntilPrompt polls ReadScreen at 500ms intervals until a prompt pattern is detected or timeout.
func pollUntilPrompt(ctx context.Context, term terminal.Terminal, paneID terminal.PaneID, patterns []CompletionPattern, timeout time.Duration) bool {
	deadline := time.After(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-deadline:
			return false
		case <-ticker.C:
			screen, err := term.ReadScreen(ctx, paneID, terminal.ReadScreenOpts{})
			if err != nil {
				continue
			}
			if isPromptVisible(screen, patterns) {
				return true
			}
		}
	}
}

// sendPrompts sends the user prompt to each interactive session.
// Sends prompt text first, then a separate Enter to submit (handles paste-mode CLIs).
func sendPrompts(ctx context.Context, cfg OrchestraConfig, panes []paneInfo) []FailedProvider {
	var failed []FailedProvider
	for i, pi := range panes {
		if pi.skipWait {
			continue
		}
		// Send prompt text (may be shown as "[Pasted text]" in some CLIs)
		if err := cfg.Terminal.SendCommand(ctx, pi.paneID, cfg.Prompt); err != nil {
			failed = append(failed, FailedProvider{
				Name:  pi.provider.Name,
				Error: fmt.Sprintf("send prompt failed: %v", err),
			})
			panes[i].skipWait = true
			continue
		}
		// Small delay to let the CLI register the pasted text
		time.Sleep(500 * time.Millisecond)
		// Send Enter separately to submit the prompt
		if err := cfg.Terminal.SendCommand(ctx, pi.paneID, "\n"); err != nil {
			failed = append(failed, FailedProvider{
				Name:  pi.provider.Name,
				Error: fmt.Sprintf("send enter failed: %v", err),
			})
			panes[i].skipWait = true
		}
	}
	return failed
}

// waitAndCollectResults waits for each provider to complete and collects cleaned results.
// @AX:WARN [AUTO] concurrent goroutine writes to shared responses slice — guarded by mu sync.Mutex
func waitAndCollectResults(ctx context.Context, cfg OrchestraConfig, panes []paneInfo, patterns []CompletionPattern, start time.Time) []ProviderResponse {
	var (
		responses []ProviderResponse
		mu        sync.Mutex
		wg        sync.WaitGroup
	)

	for _, pi := range panes {
		if pi.skipWait {
			responses = append(responses, ProviderResponse{
				Provider: pi.provider.Name,
				Duration: time.Since(start),
				TimedOut: true,
			})
			continue
		}
		wg.Add(1)
		go func(pi paneInfo) {
			defer wg.Done()
			timedOut := !waitForCompletion(ctx, cfg.Terminal, pi, patterns)
			// R9: collect partial results even on timeout
			screen, _ := cfg.Terminal.ReadScreen(ctx, pi.paneID, terminal.ReadScreenOpts{Scrollback: true})
			// R10: clean the output
			output := cleanScreenOutput(screen)

			mu.Lock()
			defer mu.Unlock()
			responses = append(responses, ProviderResponse{
				Provider: pi.provider.Name,
				Output:   output,
				Duration: time.Since(start),
				TimedOut: timedOut,
			})
		}(pi)
	}
	wg.Wait()
	return responses
}

// waitForCompletion polls for completion using ReadScreen prompt pattern detection only.
// Idle detection is disabled because it cannot distinguish "AI thinking" from "AI done".
// @AX:NOTE [AUTO] prompt pattern only — idle detection disabled to avoid false positives during AI thinking
func waitForCompletion(ctx context.Context, term terminal.Terminal, pi paneInfo, patterns []CompletionPattern) bool {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false // R9: timeout reached
		case <-ticker.C:
			// ReadScreen prompt pattern detection — wait for CLI to show a new input prompt
			screen, err := term.ReadScreen(ctx, pi.paneID, terminal.ReadScreenOpts{})
			if err == nil && isPromptVisible(screen, patterns) {
				return true
			}
		}
	}
}

// buildInteractiveLaunchCmd constructs the launch command for interactive mode.
// Uses the binary name plus model/variant flags from PaneArgs, excluding print/pipe flags.
// For claude: "claude --model opus --effort high"
// For opencode: "opencode -m openai/gpt-5.4"
// For gemini: "gemini -m gemini-3.1-pro-preview"
func buildInteractiveLaunchCmd(p ProviderConfig) string {
	cmd := p.Binary
	for _, arg := range paneArgs(p) {
		// Skip non-interactive flags that conflict with TUI mode
		if arg == "--print" || arg == "-p" || arg == "--quiet" || arg == "-q" || arg == "run" {
			continue
		}
		cmd += " " + arg
	}
	return cmd
}

// cleanupInteractivePanes stops pipe capture and closes panes.
func cleanupInteractivePanes(term terminal.Terminal, panes []paneInfo) {
	ctx := context.Background()
	for _, pi := range panes {
		_ = term.PipePaneStop(ctx, pi.paneID)
		_ = term.Close(ctx, string(pi.paneID))
		_ = os.Remove(pi.outputFile)
	}
}
