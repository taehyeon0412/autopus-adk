package browse

import (
	"context"
	"fmt"
	"strings"
)

// AgentBrowserBackend implements BrowserBackend for agent-based browser automation.
type AgentBrowserBackend struct{}

// Name returns the backend identifier.
func (a *AgentBrowserBackend) Name() string {
	return "agent-browser"
}

// Open opens a URL in an agent-managed browser session and returns a SessionID.
func (a *AgentBrowserBackend) Open(ctx context.Context, url string) (SessionID, error) {
	cmd := execCommand("agent-browser", "open", url)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("agent-browser open: %w", err)
	}
	ref := strings.TrimSpace(string(out))
	return SessionID(ref), nil
}

// Snapshot captures the current page state and returns its content.
func (a *AgentBrowserBackend) Snapshot(ctx context.Context) (string, error) {
	cmd := execCommand("agent-browser", "snapshot")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("agent-browser snapshot: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Click performs a click action on the specified selector.
func (a *AgentBrowserBackend) Click(ctx context.Context, selector string) error {
	cmd := execCommand("agent-browser", "click", selector)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("agent-browser click: %w", err)
	}
	return nil
}

// Fill fills a form field identified by selector with the given text.
func (a *AgentBrowserBackend) Fill(ctx context.Context, selector string, text string) error {
	cmd := execCommand("agent-browser", "fill", selector, text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("agent-browser fill: %w", err)
	}
	return nil
}

// Screenshot captures a screenshot and saves it to outPath.
func (a *AgentBrowserBackend) Screenshot(ctx context.Context, outPath string) error {
	cmd := execCommand("agent-browser", "screenshot", outPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("agent-browser screenshot: %w", err)
	}
	return nil
}

// Close closes the browser session. No-op for agent backend.
func (a *AgentBrowserBackend) Close(ctx context.Context) error {
	return nil
}
