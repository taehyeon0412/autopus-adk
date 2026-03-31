package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/insajin/autopus-adk/internal/cli/tui"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/detect"
)

// isStdinTTY reports whether stdin is an interactive terminal.
func isStdinTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptChoice presents a numbered selection to the user and returns the chosen index.
// Returns defaultIdx in non-interactive environments.
func promptChoice(out io.Writer, question string, options []string, defaultIdx int) int {
	if !isStdinTTY() {
		return defaultIdx
	}
	fmt.Fprintf(out, "\n  %s\n", question)
	for i, opt := range options {
		marker := "  "
		if i == defaultIdx {
			marker = "* "
		}
		fmt.Fprintf(out, "    %s%d) %s\n", marker, i+1, opt)
	}
	fmt.Fprintf(out, "  Choose [%d]: ", defaultIdx+1)

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)

	if answer == "" {
		return defaultIdx
	}

	for i := range options {
		if answer == fmt.Sprintf("%d", i+1) {
			return i
		}
	}
	return defaultIdx
}

// Language code and label data — shared with wizard_steps.go via tui package,
// kept here for update.go backward compatibility.
var langCodes = []string{"en", "ko", "ja", "zh"}
var langLabels = []string{
	"English",
	"Korean (한국어)",
	"Japanese (日本語)",
	"Chinese (中文)",
}

// promptLanguageSettings asks the user to configure project language settings.
// Skips already-configured fields. Used by update.go.
func promptLanguageSettings(cmd *cobra.Command, dir string, cfg *config.HarnessConfig) {
	if cfg.Language.Comments != "" && cfg.Language.Commits != "" && cfg.Language.AIResponses != "" {
		return
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "\n  Language Settings:")

	if cfg.Language.Comments == "" {
		idx := promptChoice(out, "Code comments language?", langLabels, 0)
		cfg.Language.Comments = langCodes[idx]
	}
	if cfg.Language.Commits == "" {
		idx := promptChoice(out, "Commit message language?", langLabels, 0)
		cfg.Language.Commits = langCodes[idx]
	}
	if cfg.Language.AIResponses == "" {
		idx := promptChoice(out, "AI response language?", langLabels, 0)
		cfg.Language.AIResponses = langCodes[idx]
	}

	if err := config.Save(dir, cfg); err != nil {
		fmt.Fprintf(out, "  [ERROR] autopus.yaml save failed: %v\n", err)
		return
	}
	fmt.Fprintf(out, "\n  Language configured: comments=%s, commits=%s, ai=%s\n",
		cfg.Language.Comments, cfg.Language.Commits, cfg.Language.AIResponses)
}

// warnParentRuleConflicts detects parent directory rule conflicts and
// offers to isolate them using a huh Confirm dialog.
func warnParentRuleConflicts(cmd *cobra.Command, dir string, cfg *config.HarnessConfig) {
	conflicts := detect.CheckParentRuleConflicts(dir)
	if len(conflicts) == 0 {
		return
	}

	out := cmd.OutOrStdout()

	// Already isolated — just inform.
	if cfg.IsolateRules {
		fmt.Fprintln(out, "\n  Parent rules detected (isolated via isolate_rules: true):")
		for _, c := range conflicts {
			fmt.Fprintf(out, "    - %s/.claude/rules/%s/ (ignored)\n", c.ParentDir, c.Namespace)
		}
		return
	}

	// Show conflicts.
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "  Parent rule conflicts detected:")
	for _, c := range conflicts {
		fmt.Fprintf(out, "    - %s/.claude/rules/%s/\n", c.ParentDir, c.Namespace)
	}
	fmt.Fprintln(out, "  Claude Code inherits rules from parent directories.")
	fmt.Fprintln(out, "  These rules will apply alongside autopus rules in this project.")
	fmt.Fprintln(out, "")

	// Non-TTY: don't prompt.
	if !isStdinTTY() {
		return
	}

	var isolate bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Ignore parent rules?").
				Description("Sets isolate_rules: true in autopus.yaml").
				Affirmative("Yes").
				Negative("No").
				Value(&isolate),
		),
	).WithTheme(tui.AutopusTheme())

	if err := form.Run(); err != nil {
		return
	}

	if isolate {
		cfg.IsolateRules = true
		if err := config.Save(dir, cfg); err != nil {
			fmt.Fprintf(out, "  [ERROR] autopus.yaml save failed: %v\n", err)
			return
		}
		fmt.Fprintln(out, "  isolate_rules: true set in autopus.yaml")
	}
}
