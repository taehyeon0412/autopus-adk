// Package terminal provides input validation for terminal adapter parameters.
package terminal

import (
	"fmt"
	"regexp"
)

// validName matches safe workspace/session names: alphanumeric, hyphens, underscores, dots.
var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

// validCmuxRef matches cmux reference format returned by CLI output (e.g., surface:7, pane:6, workspace:5).
var validCmuxRef = regexp.MustCompile(`^(surface|pane|workspace):\d+$`)

// validateWorkspaceName checks that name is safe for use as a tmux session or cmux workspace name.
func validateWorkspaceName(name string) error {
	if name == "" {
		return fmt.Errorf("workspace name must not be empty")
	}
	if len(name) > 256 {
		return fmt.Errorf("workspace name too long: %d characters (max 256)", len(name))
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid workspace name %q: must be alphanumeric with hyphens, underscores, or dots", name)
	}
	return nil
}

// validatePaneID checks that paneID is safe for use as a tmux/cmux pane target.
// Accepts both simple names (alphanumeric with hyphens/underscores/dots) and cmux refs (surface:N, pane:N).
func validatePaneID(id PaneID) error {
	if id == "" {
		return fmt.Errorf("pane ID must not be empty")
	}
	s := string(id)
	if validName.MatchString(s) || validCmuxRef.MatchString(s) {
		return nil
	}
	return fmt.Errorf("invalid pane ID %q: must be alphanumeric with hyphens/underscores/dots, or a cmux ref (surface:N, pane:N)", id)
}
