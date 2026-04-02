package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ApprovalRequest describes an action that requires user approval.
type ApprovalRequest struct {
	TaskID    string
	Action    string
	RiskLevel string // "low", "medium", "high", "critical"
	Context   string
}

// ApprovalResult represents the user's decision on an approval request.
type ApprovalResult int

const (
	ApproveResult ApprovalResult = iota
	DenyResult
	ViewDiffResult
	SkipResult
)

var (
	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorWarning).
			Padding(1, 2).
			Align(lipgloss.Center)

	riskLow = lipgloss.NewStyle().
		Foreground(colorSuccess).Bold(true)
	riskMedium = lipgloss.NewStyle().
			Foreground(colorWarning).Bold(true)
	riskHigh = lipgloss.NewStyle().
			Foreground(colorError).Bold(true)
	riskCritical = lipgloss.NewStyle().
			Foreground(colorWhite).Background(colorError).Bold(true)
)

// renderApprovalDialog renders a modal overlay for approval requests.
func renderApprovalDialog(req ApprovalRequest, width int) string {
	dialogWidth := width - 10
	if dialogWidth < 40 {
		dialogWidth = 40
	}
	if dialogWidth > 80 {
		dialogWidth = 80
	}

	title := brandStyle.Render("⚠ Approval Required")
	riskBadge := renderRiskBadge(req.RiskLevel)

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Action:  %s", req.Action))
	lines = append(lines, fmt.Sprintf("Risk:    %s", riskBadge))

	if req.Context != "" {
		lines = append(lines, "")
		lines = append(lines, mutedStyle.Render("Context:"))
		for _, line := range wrapText(req.Context, dialogWidth-6) {
			lines = append(lines, "  "+line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, renderApprovalKeys())

	content := strings.Join(lines, "\n")
	return dialogStyle.Width(dialogWidth).Render(content)
}

func renderRiskBadge(level string) string {
	switch level {
	case "low":
		return riskLow.Render(" LOW ")
	case "medium":
		return riskMedium.Render(" MEDIUM ")
	case "high":
		return riskHigh.Render(" HIGH ")
	case "critical":
		return riskCritical.Render(" CRITICAL ")
	default:
		return mutedStyle.Render(" UNKNOWN ")
	}
}

func renderApprovalKeys() string {
	return fmt.Sprintf("%s  %s  %s  %s",
		successStyle.Render("[a]pprove"),
		errorStyle.Render("[d]eny"),
		mutedStyle.Render("[v]iew diff"),
		mutedStyle.Render("[s]kip"),
	)
}

func wrapText(text string, maxWidth int) []string {
	if maxWidth < 10 {
		maxWidth = 10
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]

	for _, w := range words[1:] {
		if len(current)+1+len(w) > maxWidth {
			lines = append(lines, current)
			current = w
		} else {
			current += " " + w
		}
	}
	lines = append(lines, current)
	return lines
}
