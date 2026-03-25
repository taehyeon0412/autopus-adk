// Package pipeline provides pipeline state management types and persistence.
package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// specIDPattern allows alphanumeric characters, hyphens, and underscores (max 64 chars).
var specIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

// ValidateSpecID checks that specID contains only safe characters to prevent
// shell injection when used in file paths or commands.
func ValidateSpecID(id string) error {
	if !specIDPattern.MatchString(id) {
		return fmt.Errorf("invalid spec ID: %q", id)
	}
	return nil
}

// MonitorState tracks the current phase and agent statuses for the dashboard.
type MonitorState struct {
	Phase  string
	Agents map[string]string // agent name -> status
}

// MonitorSession manages terminal panes and log files for pipeline monitoring.
type MonitorSession struct {
	specID  string
	term    terminal.Terminal
	panes   []terminal.PaneID
	logPath string
	state   MonitorState
}

// NewMonitorSession creates a MonitorSession for the given spec ID and terminal.
func NewMonitorSession(specID string, term terminal.Terminal) *MonitorSession {
	logPath := filepath.Join(os.TempDir(), fmt.Sprintf("autopus-pipeline-%s.log", specID))
	return &MonitorSession{
		specID:  specID,
		term:    term,
		logPath: logPath,
		state: MonitorState{
			Agents: make(map[string]string),
		},
	}
}

// @AX:NOTE [AUTO] @AX:REASON: magic constant — "cmux" string must match terminal.Name() return value in cmux adapter
// isCmux returns true if the terminal is a cmux multiplexer.
func (m *MonitorSession) isCmux() bool {
	return m.term != nil && m.term.Name() == "cmux"
}

// Start initializes panes in the terminal for monitoring.
// For cmux terminals, creates 2 panes (dashboard + log tail).
// For non-cmux terminals, skips pane creation gracefully.
func (m *MonitorSession) Start(ctx context.Context) error {
	if err := ValidateSpecID(m.specID); err != nil {
		return err
	}

	if !m.isCmux() {
		return nil
	}

	// Create log tail pane.
	logPane, err := m.term.SplitPane(ctx, terminal.Vertical)
	if err != nil {
		return fmt.Errorf("create log pane: %w", err)
	}
	m.panes = append(m.panes, logPane)

	tailCmd := fmt.Sprintf("tail -f %s", m.logPath)
	// @AX:NOTE [AUTO] @AX:REASON: design choice — SendCommand errors silently ignored; pane commands are best-effort and must not block pipeline execution
	_ = m.term.SendCommand(ctx, logPane, tailCmd)

	// Create dashboard pane.
	dashPane, err := m.term.SplitPane(ctx, terminal.Horizontal)
	if err != nil {
		return fmt.Errorf("create dashboard pane: %w", err)
	}
	m.panes = append(m.panes, dashPane)

	dashCmd := fmt.Sprintf("auto pipeline dashboard %s", m.specID)
	_ = m.term.SendCommand(ctx, dashPane, dashCmd)

	return nil
}

// Close closes all panes and cleans up resources.
func (m *MonitorSession) Close(ctx context.Context) error {
	m.cleanupPanes()

	// Remove temporary log file if it exists.
	if m.logPath != "" {
		_ = os.Remove(m.logPath)
	}

	if m.term != nil {
		if err := m.term.Close(ctx, m.specID); err != nil {
			return fmt.Errorf("close terminal: %w", err)
		}
	}

	return nil
}

// cleanupPanes resets the tracked pane list.
func (m *MonitorSession) cleanupPanes() {
	m.panes = nil
}

// LogPath returns the log file path for this monitor session.
func (m *MonitorSession) LogPath() string {
	return m.logPath
}

// State returns the current monitor state.
func (m *MonitorSession) State() MonitorState {
	return m.state
}
