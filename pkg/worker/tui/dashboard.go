package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Brand colors — autopus purple/octopus theme.
var (
	colorViolet  = lipgloss.Color("#7c3aed")
	colorPurple  = lipgloss.Color("#A855F7")
	colorSuccess = lipgloss.Color("#2ECC71")
	colorError   = lipgloss.Color("#E74C3C")
	colorWarning = lipgloss.Color("#F39C12")
	colorMuted   = lipgloss.Color("#95A5A6")
	colorWhite   = lipgloss.Color("#FFFFFF")
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorViolet).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(0, 1)

	mutedStyle   = lipgloss.NewStyle().Foreground(colorMuted)
	successStyle = lipgloss.NewStyle().Foreground(colorSuccess)
	errorStyle   = lipgloss.NewStyle().Foreground(colorError)
	warningStyle = lipgloss.NewStyle().Foreground(colorWarning)
	brandStyle   = lipgloss.NewStyle().Foreground(colorViolet).Bold(true)
)

// renderHeader renders the top banner with connection status and providers.
func renderHeader(status ConnStatus, providers []ProviderInfo, width int) string {
	var statusText string
	switch status {
	case ConnConnected:
		statusText = successStyle.Render("● connected")
	case ConnDisconnected:
		statusText = errorStyle.Render("○ disconnected")
	case ConnReconnecting:
		statusText = warningStyle.Render("◌ reconnecting")
	}

	title := headerStyle.Render(" 🐙 Autopus Worker ")
	providerList := renderProviders(providers)

	return fmt.Sprintf("%s  %s\n%s", title, statusText, providerList)
}

func renderProviders(providers []ProviderInfo) string {
	if len(providers) == 0 {
		return mutedStyle.Render("  no providers registered")
	}
	var parts []string
	for _, p := range providers {
		if p.Available {
			parts = append(parts, successStyle.Render("✓ "+p.Name))
		} else {
			parts = append(parts, errorStyle.Render("✗ "+p.Name))
		}
	}
	return "  " + strings.Join(parts, "  ")
}

// renderTaskQueue renders the queued tasks section.
func renderTaskQueue(tasks []TaskInfo, width int) string {
	title := brandStyle.Render("Task Queue")

	if len(tasks) == 0 {
		content := fmt.Sprintf("%s\n%s", title, mutedStyle.Render("  idle — waiting for tasks"))
		return sectionStyle.Width(clampWidth(width)).Render(content)
	}

	var lines []string
	lines = append(lines, title)

	maxDisplay := 5
	for i, t := range tasks {
		if i >= maxDisplay {
			lines = append(lines, mutedStyle.Render(
				fmt.Sprintf("  ... +%d more", len(tasks)-maxDisplay)))
			break
		}
		statusIcon := taskStatusIcon(t.Status)
		lines = append(lines, fmt.Sprintf("  %s %s %s",
			statusIcon, mutedStyle.Render(t.ID), t.Description))
	}

	return sectionStyle.Width(clampWidth(width)).Render(strings.Join(lines, "\n"))
}

func taskStatusIcon(status string) string {
	switch status {
	case "running":
		return warningStyle.Render("▶")
	case "completed":
		return successStyle.Render("✓")
	case "failed":
		return errorStyle.Render("✗")
	default:
		return mutedStyle.Render("○")
	}
}

// renderCurrentTask renders the active task section with optional detail.
func renderCurrentTask(task *CurrentTask, showDetail bool, width int) string {
	title := brandStyle.Render("Current Task")
	w := clampWidth(width)

	lines := []string{
		title,
		fmt.Sprintf("  ID:    %s", task.ID),
		fmt.Sprintf("  Phase: %s", task.Phase),
		"  " + renderProgressBar(task.Progress, w-6),
	}

	if showDetail && task.SubagentTree != "" {
		lines = append(lines, "")
		lines = append(lines, brandStyle.Render("  Subagent Tree"))
		lines = append(lines, task.SubagentTree)
	}

	return sectionStyle.Width(w).Render(strings.Join(lines, "\n"))
}

// renderCostFooter renders the cost tracker at the bottom.
func renderCostFooter(tracker CostTracker, width int) string {
	current := fmt.Sprintf("Task: $%.4f", tracker.CurrentTaskCost)
	daily := fmt.Sprintf("Daily: $%.2f", tracker.DailyCost)

	return mutedStyle.Render(fmt.Sprintf("  💰 %s  │  %s", current, daily))
}

// renderProgressBar renders an ASCII progress bar.
func renderProgressBar(progress float64, width int) string {
	if width < 10 {
		width = 10
	}
	barWidth := width - 8 // room for brackets and percentage
	if barWidth < 5 {
		barWidth = 5
	}

	filled := int(progress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	pct := fmt.Sprintf("%3.0f%%", progress*100)

	return fmt.Sprintf("[%s] %s", brandStyle.Render(bar), pct)
}

func clampWidth(width int) int {
	w := width - 4
	if w < 40 {
		return 40
	}
	if w > 120 {
		return 120
	}
	return w
}
