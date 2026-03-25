package pipeline

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRenderDashboard_PendingState verifies dashboard renders correctly
// when all phases are pending.
func TestRenderDashboard_PendingState(t *testing.T) {
	t.Parallel()

	data := DashboardData{
		Phases: map[string]PhaseStatus{
			"phase1":   PhasePending,
			"phase1.5": PhasePending,
			"phase2":   PhasePending,
			"phase3":   PhasePending,
		},
		Agents: map[string]string{},
	}

	output := RenderDashboard(data)
	assert.NotEmpty(t, output,
		"RenderDashboard should produce non-empty output for pending state")
	assert.Contains(t, output, "pending",
		"output should indicate pending status")
}

// TestRenderDashboard_RunningState verifies dashboard renders correctly
// when Phase 2 is running with active executors.
func TestRenderDashboard_RunningState(t *testing.T) {
	t.Parallel()

	data := DashboardData{
		Phases: map[string]PhaseStatus{
			"phase1":   PhaseDone,
			"phase1.5": PhaseDone,
			"phase2":   PhaseRunning,
			"phase3":   PhasePending,
		},
		Agents: map[string]string{
			"executor-1": "running",
			"executor-2": "running",
		},
		Elapsed: 45 * time.Second,
	}

	output := RenderDashboard(data)
	assert.NotEmpty(t, output,
		"RenderDashboard should produce non-empty output for running state")

	// Should show running phase.
	lower := strings.ToLower(output)
	assert.True(t,
		strings.Contains(lower, "running") || strings.Contains(lower, "phase2"),
		"output should indicate phase2 is running")
}

// TestRenderDashboard_CompletedState verifies dashboard renders correctly
// when all phases are done.
func TestRenderDashboard_CompletedState(t *testing.T) {
	t.Parallel()

	data := DashboardData{
		Phases: map[string]PhaseStatus{
			"phase1":   PhaseDone,
			"phase1.5": PhaseDone,
			"phase2":   PhaseDone,
			"phase3":   PhaseDone,
		},
		Agents:  map[string]string{},
		Elapsed: 5 * time.Minute,
	}

	output := RenderDashboard(data)
	assert.NotEmpty(t, output,
		"RenderDashboard should produce non-empty output for completed state")
	assert.Contains(t, strings.ToLower(output), "done",
		"output should indicate completion")
}

// TestRenderDashboard_WithBlocker verifies dashboard shows blocker info.
func TestRenderDashboard_WithBlocker(t *testing.T) {
	t.Parallel()

	data := DashboardData{
		Phases: map[string]PhaseStatus{
			"phase2": PhaseRunning,
		},
		Agents:  map[string]string{"executor-1": "blocked"},
		Blocker: "merge conflict in pkg/pipeline/types.go",
	}

	output := RenderDashboard(data)
	assert.NotEmpty(t, output,
		"RenderDashboard should produce non-empty output with blocker")
	assert.Contains(t, output, "merge conflict",
		"output should display the blocker message")
}

// TestRenderDashboard_ElapsedTime verifies elapsed time formatting.
func TestRenderDashboard_ElapsedTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		elapsed  time.Duration
		contains string
	}{
		{"seconds", 30 * time.Second, "30s"},
		{"minutes_and_seconds", 2*time.Minute + 30*time.Second, "2m30s"},
		{"hours", 1*time.Hour + 5*time.Minute, "1h5m"},
		{"zero", 0, "0s"},
		{"negative", -1 * time.Second, "0s"},
		{"sub_second", 500 * time.Millisecond, "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			formatted := FormatElapsed(tt.elapsed)
			assert.NotEmpty(t, formatted,
				"FormatElapsed should return non-empty string")
			assert.Contains(t, formatted, tt.contains,
				"FormatElapsed(%v) should contain %q", tt.elapsed, tt.contains)
		})
	}
}

// TestStatusIcon_AllStatuses verifies statusIcon returns correct icons for all PhaseStatus values.
func TestStatusIcon_AllStatuses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   PhaseStatus
		contains string
	}{
		{"done", PhaseDone, "done"},
		{"running", PhaseRunning, "running"},
		{"failed", PhaseFailed, "failed"},
		{"pending", PhasePending, "pending"},
		{"unknown", PhaseStatus("unknown"), "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			icon := statusIcon(tt.status)
			assert.Contains(t, icon, tt.contains,
				"statusIcon(%q) should contain %q", tt.status, tt.contains)
		})
	}
}

// TestRenderDashboard_FailedPhase verifies dashboard renders failed phase with correct icon.
func TestRenderDashboard_FailedPhase(t *testing.T) {
	t.Parallel()

	data := DashboardData{
		Phases: map[string]PhaseStatus{
			"phase1": PhaseDone,
			"phase2": PhaseFailed,
		},
		Agents:  map[string]string{},
		Elapsed: 10 * time.Second,
	}

	output := RenderDashboard(data)
	assert.Contains(t, output, "failed",
		"output should show failed status")
}

// TestRenderDashboard_EmptyPhases verifies dashboard handles empty phases map.
func TestRenderDashboard_EmptyPhases(t *testing.T) {
	t.Parallel()

	data := DashboardData{
		Phases: map[string]PhaseStatus{},
		Agents: map[string]string{},
	}

	output := RenderDashboard(data)
	assert.NotEmpty(t, output,
		"RenderDashboard should produce output even with empty phases")
	assert.Contains(t, output, "Dashboard",
		"output should contain header")
}
