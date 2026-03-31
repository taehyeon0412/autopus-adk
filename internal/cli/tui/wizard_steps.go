package tui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// Language code and label data for the language selection step.
var (
	langCodes  = []string{"en", "ko", "ja", "zh"}
	langLabels = []string{"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)"}
)

// InitWizardOpts holds options for the init wizard.
type InitWizardOpts struct {
	Quality      string   // pre-set from --quality flag
	NoReviewGate bool     // pre-set from --no-review-gate flag
	Platforms    []string // pre-set from --platforms flag
	Accessible   bool     // true for non-TTY / --yes mode
	Providers    []string // detected orchestra providers
}

// InitWizardResult holds all wizard selections.
type InitWizardResult struct {
	CommentsLang string
	CommitsLang  string
	AILang       string
	Quality      string
	ReviewGate   bool
	Methodology  string
	Cancelled    bool
}

// RunInitWizard runs the interactive init wizard using huh forms.
// When opts.Accessible is true, it returns defaults without launching a TUI.
func RunInitWizard(opts InitWizardOpts) (*InitWizardResult, error) {
	EnsureSafeEnv()
	InitStyles()

	result := defaultResult(opts)

	// Non-TTY / --yes mode: return defaults directly (R3, C2)
	if opts.Accessible {
		return result, nil
	}

	steps := buildStepList(opts)
	step := 0
	for step < len(steps) {
		form := steps[step](result)
		if err := form.Run(); err != nil {
			// User cancelled (ctrl+c or esc)
			return &InitWizardResult{Cancelled: true}, nil
		}
		step++
	}

	return result, nil
}

// stepBuilder is a function that creates a huh form for a given step.
type stepBuilder func(r *InitWizardResult) *huh.Form

// @AX:NOTE [AUTO]: buildStepList uses closure-rebinding to inject the final
// step count into each step's title. Steps are collected first, then totalRef
// is set so all closures share the correct total when invoked.
// buildStepList assembles the wizard steps, skipping pre-configured ones.
func buildStepList(opts InitWizardOpts) []stepBuilder {
	var steps []stepBuilder
	totalRef := new(int) // shared reference updated after building

	steps = append(steps, func(r *InitWizardResult) *huh.Form {
		return buildLangStep(len(steps), *totalRef, r)
	})

	if opts.Quality == "" {
		steps = append(steps, func(r *InitWizardResult) *huh.Form {
			return buildQualityStep(len(steps), *totalRef, r)
		})
	}

	if !opts.NoReviewGate {
		steps = append(steps, func(r *InitWizardResult) *huh.Form {
			return buildReviewGateStep(len(steps), *totalRef, r, opts)
		})
	}

	steps = append(steps, func(r *InitWizardResult) *huh.Form {
		return buildMethodologyStep(len(steps), *totalRef, r)
	})

	*totalRef = len(steps)

	// Rebind closures with correct indices now that total is known.
	total := len(steps)
	rebuilt := make([]stepBuilder, total)
	idx := 0

	rebuilt[idx] = func(r *InitWizardResult) *huh.Form {
		return buildLangStep(idx+1, total, r)
	}
	idx++

	if opts.Quality == "" {
		i := idx
		rebuilt[idx] = func(r *InitWizardResult) *huh.Form {
			return buildQualityStep(i+1, total, r)
		}
		idx++
	}

	if !opts.NoReviewGate {
		i := idx
		rebuilt[idx] = func(r *InitWizardResult) *huh.Form {
			return buildReviewGateStep(i+1, total, r, opts)
		}
		idx++
	}

	i := idx
	rebuilt[idx] = func(r *InitWizardResult) *huh.Form {
		return buildMethodologyStep(i+1, total, r)
	}

	return rebuilt
}

// buildLangStep creates the language selection step (3 sub-selects).
func buildLangStep(num, total int, r *InitWizardResult) *huh.Form {
	title := fmt.Sprintf("[%d/%d] Language Settings", num, total)
	opts := langOptions()

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Description("Code comments language").
				Options(opts...).
				Value(&r.CommentsLang),
			huh.NewSelect[string]().
				Title("Commit messages").
				Options(opts...).
				Value(&r.CommitsLang),
			huh.NewSelect[string]().
				Title("AI responses").
				Options(opts...).
				Value(&r.AILang),
		),
	).WithTheme(AutopusTheme()).WithWidth(bannerWidth + 10)
}

// buildQualityStep creates the quality mode selection step.
func buildQualityStep(num, total int, r *InitWizardResult) *huh.Form {
	title := fmt.Sprintf("[%d/%d] Quality Mode", num, total)

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(
					huh.NewOption("Ultra — strict review, high coverage", "ultra"),
					huh.NewOption("Balanced — pragmatic defaults", "balanced"),
				).
				Value(&r.Quality),
		),
	).WithTheme(AutopusTheme()).WithWidth(bannerWidth + 10)
}

// buildReviewGateStep creates the review gate confirmation step.
func buildReviewGateStep(num, total int, r *InitWizardResult, opts InitWizardOpts) *huh.Form {
	title := fmt.Sprintf("[%d/%d] Review Gate", num, total)
	desc := "Require human review before merging"
	if len(opts.Providers) > 0 {
		desc = fmt.Sprintf("Providers detected: %d — enable review gate?", len(opts.Providers))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(desc).
				Affirmative("Yes").
				Negative("No").
				Value(&r.ReviewGate),
		),
	).WithTheme(AutopusTheme()).WithWidth(bannerWidth + 10)
}

// buildMethodologyStep creates the methodology selection step.
func buildMethodologyStep(num, total int, r *InitWizardResult) *huh.Form {
	title := fmt.Sprintf("[%d/%d] Methodology", num, total)

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(
					huh.NewOption("TDD — test-driven development", "tdd"),
					huh.NewOption("None — no enforced methodology", "none"),
				).
				Value(&r.Methodology),
		),
	).WithTheme(AutopusTheme()).WithWidth(bannerWidth + 10)
}

// langOptions builds huh options from the language data.
func langOptions() []huh.Option[string] {
	opts := make([]huh.Option[string], len(langCodes))
	for i, code := range langCodes {
		opts[i] = huh.NewOption(langLabels[i], code)
	}
	return opts
}

// defaultResult returns a result populated with sensible defaults,
// respecting any pre-configured values from opts.
func defaultResult(opts InitWizardOpts) *InitWizardResult {
	r := &InitWizardResult{
		CommentsLang: "en",
		CommitsLang:  "en",
		AILang:       "en",
		Quality:      "balanced",
		ReviewGate:   len(opts.Providers) >= 2,
		Methodology:  "tdd",
	}

	// Apply pre-configured values (R9, R13)
	if opts.Quality != "" {
		r.Quality = opts.Quality
	}
	if opts.NoReviewGate {
		r.ReviewGate = false
	}

	return r
}
