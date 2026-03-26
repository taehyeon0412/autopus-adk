package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// Compile-time interface compliance check.
var _ PipelineMonitor = (*TeamMonitorSession)(nil)

// TeamMonitorSession manages terminal panes for team pipeline monitoring.
// Each teammate gets a dedicated pane with real-time log streaming.
// The initial pane serves as the dashboard.
type TeamMonitorSession struct {
	specID     string
	term       terminal.Terminal
	teammates  []string
	panes      []TeammatePaneInfo
	dashPaneID terminal.PaneID
	logPath    string // primary log path (dashboard)
	state      MonitorState

	mu sync.Mutex // protects state for concurrent UpdateAgent calls
}

// NewTeamMonitorSession creates a TeamMonitorSession for the given spec and team.
func NewTeamMonitorSession(specID string, term terminal.Terminal, teammates []string) *TeamMonitorSession {
	return &TeamMonitorSession{
		specID:    specID,
		term:      term,
		teammates: teammates,
		state: MonitorState{
			Agents: make(map[string]string),
		},
	}
}

// @AX:NOTE [AUTO] @AX:REASON: design choice — uses name != "plain" instead of explicit cmux/tmux check to support future multiplexer adapters without code changes
// isMultiplexer returns true if the terminal supports pane splitting (cmux or tmux).
func (t *TeamMonitorSession) isMultiplexer() bool {
	return t.term != nil && t.term.Name() != "plain"
}

// Start creates the team pane layout and begins log streaming.
// For plain terminals, returns nil without creating panes (graceful degradation).
func (t *TeamMonitorSession) Start(ctx context.Context) error {
	if err := ValidateSpecID(t.specID); err != nil {
		return err
	}

	if !t.isMultiplexer() {
		return nil
	}

	// Plan and apply the sequential vertical split layout.
	plan := planLayout(t.teammates)
	result, err := applyLayout(ctx, t.term, plan)
	if err != nil {
		return fmt.Errorf("team layout: %w", err)
	}

	// Create log files and start tail -f streaming in each pane.
	panes, err := createTeammatePanes(ctx, t.term, t.specID, result.TeammatePaneIDs, t.teammates)
	if err != nil {
		// Cleanup layout panes on pane creation failure.
		cleanupLayoutPanes(ctx, t.term, result.TeammatePaneIDs)
		return fmt.Errorf("create teammate panes: %w", err)
	}

	t.panes = panes
	t.dashPaneID = result.DashboardPaneID

	// Set primary log path from the first teammate (lead) for LogPath() compatibility.
	if len(panes) > 0 {
		t.logPath = panes[0].LogPath
	}

	// Initialize teammate statuses in state.
	for _, role := range t.teammates {
		t.state.Agents[role] = "pending"
	}

	// Render initial dashboard.
	t.refreshDashboard(ctx)

	return nil
}

// UpdateAgent updates a teammate's status and refreshes the dashboard.
func (t *TeamMonitorSession) UpdateAgent(name, status string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.state.Agents[name] = status

	// Best-effort dashboard refresh (ignore errors on non-multiplexer).
	if t.isMultiplexer() {
		t.refreshDashboard(context.Background())
	}
}

// Close cleans up all panes, removes log files, and closes the terminal session.
func (t *TeamMonitorSession) Close(ctx context.Context) error {
	cleanupTeammatePanes(t.panes)
	t.panes = nil

	if t.term != nil {
		if err := t.term.Close(ctx, t.specID); err != nil {
			return fmt.Errorf("close terminal: %w", err)
		}
	}

	return nil
}

// LogPath returns the primary log file path for this monitor session.
func (t *TeamMonitorSession) LogPath() string {
	return t.logPath
}

// State returns the current monitor state.
func (t *TeamMonitorSession) State() MonitorState {
	return t.state
}

// Panes returns the current teammate pane info for external inspection.
func (t *TeamMonitorSession) Panes() []TeammatePaneInfo {
	return t.panes
}

// FailTeammate sends a failure message to a specific teammate's pane.
func (t *TeamMonitorSession) FailTeammate(ctx context.Context, role, errMsg string) {
	t.mu.Lock()
	t.state.Agents[role] = "failed"
	t.mu.Unlock()

	for _, p := range t.panes {
		if p.Role == role {
			sendFailureMessage(ctx, t.term, p.PaneID, role, errMsg)
			break
		}
	}

	if t.isMultiplexer() {
		t.refreshDashboard(ctx)
	}
}

// @AX:NOTE [AUTO] @AX:REASON: design choice — best-effort dashboard update; errors silently ignored because dashboard rendering must not block pipeline execution
// refreshDashboard renders and sends the team dashboard to the dashboard pane.
func (t *TeamMonitorSession) refreshDashboard(ctx context.Context) {
	teammates := make([]TeammateStatus, 0, len(t.teammates))
	for _, role := range t.teammates {
		status := t.state.Agents[role]
		phase := t.state.Phase
		teammates = append(teammates, NewTeammateStatus(role, phase, status))
	}

	data := TeamDashboardData{
		DashboardData: DashboardData{
			Phases:  make(map[string]PhaseStatus),
			Agents:  t.state.Agents,
			Elapsed: time.Duration(0),
		},
		Teammates: teammates,
	}

	rendered := RenderTeamDashboard(data, 0)
	if t.dashPaneID != "" {
		_ = t.term.SendCommand(ctx, t.dashPaneID, fmt.Sprintf("echo %s", teamShellEscape(rendered)))
	}
}
