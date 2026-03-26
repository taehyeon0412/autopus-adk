package pipeline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTeammateIcon(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status string
		icon   string
	}{
		{"running", "\033[33m▶\033[0m"},
		{"done", "\033[32m✓\033[0m"},
		{"failed", "\033[31m✗\033[0m"},
		{"pending", "\033[2m○\033[0m"},
		{"unknown", "\033[2m○\033[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.icon, teammateIcon(tt.status))
		})
	}
}

func TestNewTeammateStatus(t *testing.T) {
	t.Parallel()
	ts := NewTeammateStatus("builder", "phase2", "running")
	assert.Equal(t, "builder", ts.Role)
	assert.Equal(t, "phase2", ts.Phase)
	assert.Equal(t, "running", ts.Status)
	assert.Equal(t, teammateIcon("running"), ts.Icon)
}

// S13: compact dashboard rendering
func TestRenderTeamDashboard_CompactMode(t *testing.T) {
	t.Parallel()
	data := TeamDashboardData{
		DashboardData: DashboardData{
			Phases: map[string]PhaseStatus{},
			Agents: map[string]string{},
		},
		Teammates: []TeammateStatus{
			NewTeammateStatus("lead", "", "running"),
			NewTeammateStatus("builder", "", "pending"),
		},
	}

	rendered := RenderTeamDashboard(data, 30)
	assert.Contains(t, rendered, "Team")
	assert.Contains(t, rendered, "lead")
	assert.Contains(t, rendered, "builder")
	// Compact mode: should NOT contain "Team Members" (that's normal mode).
	assert.NotContains(t, rendered, "Team Members")
}

func TestRenderTeamDashboard_NormalMode(t *testing.T) {
	t.Parallel()
	data := TeamDashboardData{
		DashboardData: DashboardData{
			Phases: map[string]PhaseStatus{},
			Agents: map[string]string{},
		},
		Teammates: []TeammateStatus{
			NewTeammateStatus("lead", "phase2", "running"),
			NewTeammateStatus("builder", "phase2", "done"),
		},
	}

	rendered := RenderTeamDashboard(data, 80)
	assert.Contains(t, rendered, "Team Members")
	assert.Contains(t, rendered, "lead")
	assert.Contains(t, rendered, "builder")
	assert.Contains(t, rendered, "phase2")
}

func TestRenderTeamDashboard_DefaultWidth(t *testing.T) {
	t.Parallel()
	data := TeamDashboardData{
		DashboardData: DashboardData{
			Phases: map[string]PhaseStatus{},
			Agents: map[string]string{},
		},
		Teammates: []TeammateStatus{
			NewTeammateStatus("lead", "", "done"),
		},
	}

	// maxWidth=0 uses default boxWidth.
	rendered := RenderTeamDashboard(data, 0)
	assert.Contains(t, rendered, "Team Members")
}

func TestRenderTeamDashboard_NoTeammates(t *testing.T) {
	t.Parallel()
	data := TeamDashboardData{
		DashboardData: DashboardData{
			Phases: map[string]PhaseStatus{},
			Agents: map[string]string{},
		},
	}

	rendered := RenderTeamDashboard(data, 0)
	assert.Contains(t, rendered, "Team Pipeline Dashboard")
	assert.NotContains(t, rendered, "Team Members")
}

func TestRenderTeamDashboard_WithBlocker(t *testing.T) {
	t.Parallel()
	data := TeamDashboardData{
		DashboardData: DashboardData{
			Phases:  map[string]PhaseStatus{},
			Agents:  map[string]string{},
			Blocker: "waiting for approval",
		},
		Teammates: []TeammateStatus{
			NewTeammateStatus("lead", "", "pending"),
		},
	}

	rendered := RenderTeamDashboard(data, 0)
	assert.Contains(t, rendered, "Blocker:")
	assert.Contains(t, rendered, "waiting for approval")
}

func TestRenderTeamDashboard_WithPhases(t *testing.T) {
	t.Parallel()
	data := TeamDashboardData{
		DashboardData: DashboardData{
			Phases: map[string]PhaseStatus{
				"phase1": PhaseDone,
				"phase2": PhaseRunning,
			},
			Agents: map[string]string{},
		},
		Teammates: []TeammateStatus{
			NewTeammateStatus("lead", "phase2", "running"),
		},
	}

	rendered := RenderTeamDashboard(data, 0)
	assert.Contains(t, rendered, "Planning")
	assert.Contains(t, rendered, "Implementation")
}

func TestRenderTeamDashboard_BoxDrawing(t *testing.T) {
	t.Parallel()
	data := TeamDashboardData{
		DashboardData: DashboardData{
			Phases: map[string]PhaseStatus{},
			Agents: map[string]string{},
		},
	}

	rendered := RenderTeamDashboard(data, 0)
	lines := strings.Split(rendered, "\n")
	assert.True(t, strings.HasPrefix(lines[0], "\u2554"), "should start with top-left corner")
}

func TestRenderTeamDashboard_CompactEmptyPhase(t *testing.T) {
	t.Parallel()
	data := TeamDashboardData{
		DashboardData: DashboardData{
			Phases: map[string]PhaseStatus{},
			Agents: map[string]string{},
		},
		Teammates: []TeammateStatus{
			NewTeammateStatus("lead", "", "running"),
		},
	}

	// Compact: empty phase should show "-" in normal mode but compact just shows icon+role.
	rendered := RenderTeamDashboard(data, 30)
	assert.Contains(t, rendered, "lead")
}
