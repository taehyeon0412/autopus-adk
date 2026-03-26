package browse

import (
	"context"
	"os/exec"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// execCommand is a package-level variable for exec.Command, mockable in tests.
// @AX:WARN [AUTO] global state mutation — mutable package-level variable used for test injection
// @AX:REASON: required for unit testing CLI wrappers; do not use t.Parallel() in tests that mock this var
var execCommand = exec.Command

// SessionID identifies a browser session.
type SessionID string

// BrowserBackend abstracts browser automation across terminal environments.
type BrowserBackend interface {
	Open(ctx context.Context, url string) (SessionID, error)
	Snapshot(ctx context.Context) (string, error)
	Click(ctx context.Context, selector string) error
	Fill(ctx context.Context, selector string, text string) error
	Screenshot(ctx context.Context, outPath string) error
	Close(ctx context.Context) error
	Name() string
}

// NewBackend returns the appropriate BrowserBackend for the given terminal.
// cmux → CmuxBrowserBackend, otherwise → AgentBrowserBackend.
func NewBackend(term terminal.Terminal) BrowserBackend {
	if term != nil && term.Name() == "cmux" {
		return &CmuxBrowserBackend{}
	}
	return &AgentBrowserBackend{}
}
