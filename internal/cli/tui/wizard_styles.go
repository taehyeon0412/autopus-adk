// Package tui provides wizard-specific huh theme using Autopus brand colors.
package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// AutopusTheme returns a custom huh theme using Autopus brand colors.
// R4: Focused option highlighted with ColorViolet (#7c3aed).
// R8: Uses only colors defined in style.go.
func AutopusTheme() *huh.Theme {
	t := huh.ThemeCharm()

	// Focused styles — brand violet as primary accent.
	t.Focused.Base = t.Focused.Base.BorderForeground(ColorViolet)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(ColorViolet).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(ColorViolet).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(ColorMuted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(ColorError)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(ColorError)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(ColorViolet)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(ColorViolet)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(ColorViolet)
	t.Focused.Option = t.Focused.Option.Foreground(ColorWhite)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(ColorViolet)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(ColorViolet)
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(ColorSuccess)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(ColorMuted)
	t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(ColorMuted)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(ColorWhite).Background(ColorViolet)
	t.Focused.Next = t.Focused.FocusedButton
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(ColorMuted).Background(ColorDark)

	// Text input styles.
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(ColorPurple)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(ColorViolet)

	// Blurred styles — inherit focused, hide border.
	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.Title = t.Blurred.Title.Foreground(ColorMuted)
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	// Group styles.
	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description

	return t
}
