// Package tui provides Autopus brand styling for terminal output.
package tui

import "github.com/charmbracelet/lipgloss"

// Brand colors from autopus.co
var (
	// Primary violet gradient endpoints
	ColorViolet     = lipgloss.Color("#7c3aed") // brand primary
	ColorVioletDark = lipgloss.Color("#6D28D9") // deep violet
	ColorPurple     = lipgloss.Color("#A855F7") // light violet

	// Accent pink/rose
	ColorPink = lipgloss.Color("#F43F5E")
	ColorRose = lipgloss.Color("#FB7185")

	// Semantic colors
	ColorSuccess = lipgloss.Color("#2ECC71")
	ColorError   = lipgloss.Color("#E74C3C")
	ColorWarning = lipgloss.Color("#F39C12")
	ColorInfo    = lipgloss.Color("#3498DB")

	// Neutral
	ColorMuted = lipgloss.Color("#95A5A6")
	ColorDark  = lipgloss.Color("#1F2937")
	ColorWhite = lipgloss.Color("#FFFFFF")
)

// Reusable text styles
var (
	// Brand styles
	BrandStyle = lipgloss.NewStyle().
			Foreground(ColorViolet).
			Bold(true)

	AccentStyle = lipgloss.NewStyle().
			Foreground(ColorPink)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	// Semantic styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)

	// Label styles (bold variant)
	SuccessLabelStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	ErrorLabelStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	WarningLabelStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	InfoLabelStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Bold(true)
)
