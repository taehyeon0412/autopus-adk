package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/insajin/autopus-adk/pkg/version"
)

const bannerWidth = 40

// Banner prints the Autopus brand banner with version info.
func Banner(w io.Writer) {
	octopus := AccentStyle.Render("🐙")
	name := BrandStyle.Render("Autopus")
	line := MutedStyle.Render(strings.Repeat("─", bannerWidth-14))
	fmt.Fprintf(w, "%s %s %s\n", octopus, name, line)

	ver := MutedStyle.Render(fmt.Sprintf("v%s", version.Version()))
	fmt.Fprintf(w, "   %s\n", ver)
}

// BannerWithInfo prints the banner with project context.
func BannerWithInfo(w io.Writer, project, mode string) {
	octopus := AccentStyle.Render("🐙")
	name := BrandStyle.Render("Autopus")
	line := MutedStyle.Render(strings.Repeat("─", bannerWidth-14))
	fmt.Fprintf(w, "%s %s %s\n", octopus, name, line)

	ver := MutedStyle.Render(fmt.Sprintf("v%s", version.Version()))
	proj := lipgloss.NewStyle().Foreground(ColorPurple).Render(project)
	m := MutedStyle.Render(mode)
	fmt.Fprintf(w, "   %s │ %s │ %s\n", ver, proj, m)
}

// SectionHeader prints a styled section header.
func SectionHeader(w io.Writer, title string) {
	styled := lipgloss.NewStyle().
		Foreground(ColorViolet).
		Bold(true).
		Render(title)
	line := MutedStyle.Render(strings.Repeat("─", bannerWidth-len(title)-1))
	fmt.Fprintf(w, "\n%s %s\n", styled, line)
}

// Divider prints a thin muted divider line.
func Divider(w io.Writer) {
	fmt.Fprintln(w, MutedStyle.Render(strings.Repeat("─", bannerWidth)))
}

// Octopus prints the brand sign-off emoji.
func Octopus(w io.Writer) {
	fmt.Fprintln(w, AccentStyle.Render("🐙"))
}
