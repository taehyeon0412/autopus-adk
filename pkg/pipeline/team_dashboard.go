// Package pipeline — team dashboard extends the base dashboard with teammate status rendering.
package pipeline

import (
	"fmt"
	"strings"
)

// TeammateStatus holds the display status of a single teammate.
type TeammateStatus struct {
	Role   string
	Phase  string
	Status string // running, done, failed, pending
	Icon   string // computed from Status
}

// TeamDashboardData extends DashboardData with teammate status information.
type TeamDashboardData struct {
	DashboardData
	Teammates []TeammateStatus
}

// teammateIcon returns a display icon for the given teammate status.
func teammateIcon(status string) string {
	switch status {
	case "done":
		return "\033[32m✓\033[0m"
	case "running":
		return "\033[33m▶\033[0m"
	case "failed":
		return "\033[31m✗\033[0m"
	default:
		return "\033[2m○\033[0m"
	}
}

// RenderTeamDashboard produces a box-drawing dashboard string with teammate statuses.
// When maxWidth < 38, it switches to compact mode with narrower box width.
func RenderTeamDashboard(data TeamDashboardData, maxWidth int) string {
	width := boxWidth
	if maxWidth > 0 && maxWidth < boxWidth {
		width = maxWidth
	}

	var b strings.Builder

	top := "╔" + strings.Repeat("═", width) + "╗"
	sep := "╠" + strings.Repeat("═", width) + "╣"
	bot := "╚" + strings.Repeat("═", width) + "╝"

	line := func(text string) string {
		return fmt.Sprintf("║  %-*s║", width-2, text)
	}

	// Header
	b.WriteString(top + "\n")
	b.WriteString(line("Team Pipeline Dashboard") + "\n")
	b.WriteString(sep + "\n")

	// Phase section (reuse phaseOrder from dashboard.go)
	for _, p := range phaseOrder {
		status, ok := data.Phases[p.key]
		if !ok {
			continue
		}
		label := fmt.Sprintf("%s: %s", p.name, statusIcon(status))
		b.WriteString("║ " + label + "\n")
	}

	// Teammate section
	if len(data.Teammates) > 0 {
		b.WriteString(sep + "\n")
		if width < boxWidth {
			// Compact mode: role + icon only
			b.WriteString(line("Team") + "\n")
			for _, tm := range data.Teammates {
				icon := teammateIcon(tm.Status)
				label := fmt.Sprintf(" %s %s", icon, tm.Role)
				b.WriteString("║" + label + "\n")
			}
		} else {
			// Normal mode: role + phase + status
			b.WriteString(line("Team Members") + "\n")
			for _, tm := range data.Teammates {
				icon := teammateIcon(tm.Status)
				phase := tm.Phase
				if phase == "" {
					phase = "-"
				}
				label := fmt.Sprintf(" %s %-12s %s", icon, tm.Role, phase)
				b.WriteString("║" + label + "\n")
			}
		}
	}

	// Blocker
	if data.Blocker != "" {
		b.WriteString(sep + "\n")
		b.WriteString(line("Blocker: "+data.Blocker) + "\n")
	}

	// Elapsed
	b.WriteString(sep + "\n")
	b.WriteString(line("Elapsed: "+FormatElapsed(data.Elapsed)) + "\n")
	b.WriteString(bot + "\n")

	return b.String()
}

// NewTeammateStatus creates a TeammateStatus with the computed icon.
func NewTeammateStatus(role, phase, status string) TeammateStatus {
	return TeammateStatus{
		Role:   role,
		Phase:  phase,
		Status: status,
		Icon:   teammateIcon(status),
	}
}
