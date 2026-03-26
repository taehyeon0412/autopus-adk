package browse

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

// CmuxBrowserBackend implements BrowserBackend for cmux terminal environments.
type CmuxBrowserBackend struct {
	surfaceRef string
}

// Name returns the backend identifier.
func (c *CmuxBrowserBackend) Name() string {
	return "cmux"
}

// Open opens a URL in a cmux-managed browser session.
func (c *CmuxBrowserBackend) Open(ctx context.Context, url string) (SessionID, error) {
	cmd := execCommand("cmux", "browser", "open", url)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cmux browser open: %w", err)
	}
	surfaceRef := strings.TrimSpace(out.String())
	c.surfaceRef = surfaceRef
	return SessionID(surfaceRef), nil
}

// Snapshot captures the current page state.
func (c *CmuxBrowserBackend) Snapshot(ctx context.Context) (string, error) {
	cmd := execCommand("cmux", "browser", "--surface", c.surfaceRef, "snapshot")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cmux browser snapshot: %w", err)
	}
	return strings.TrimSpace(out.String()), nil
}

// Click performs a click action on the specified selector.
func (c *CmuxBrowserBackend) Click(ctx context.Context, selector string) error {
	cmd := execCommand("cmux", "browser", "--surface", c.surfaceRef, "click", selector)
	return cmd.Run()
}

// Fill fills a form field with text.
func (c *CmuxBrowserBackend) Fill(ctx context.Context, selector string, text string) error {
	cmd := execCommand("cmux", "browser", "--surface", c.surfaceRef, "fill", selector, text)
	return cmd.Run()
}

// Screenshot captures a screenshot and saves it to the specified path.
func (c *CmuxBrowserBackend) Screenshot(ctx context.Context, outPath string) error {
	cmd := execCommand("cmux", "browser", "--surface", c.surfaceRef, "screenshot", "--out", outPath)
	return cmd.Run()
}

// Close closes the browser session.
func (c *CmuxBrowserBackend) Close(ctx context.Context) error {
	cmd := execCommand("cmux", "close-surface", "--surface", c.surfaceRef)
	return cmd.Run()
}
