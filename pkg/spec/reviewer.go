package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	verdictRe = regexp.MustCompile(`(?i)VERDICT:\s*(PASS|REVISE|REJECT)`)
	findingRe = regexp.MustCompile(`(?i)FINDING:\s*\[(\w+)]\s*(.+)`)
)

// ParseVerdict extracts a ReviewResult from raw provider output.
func ParseVerdict(specID, output, provider string, revision int) ReviewResult {
	result := ReviewResult{
		SpecID:    specID,
		Verdict:   VerdictPass,
		Responses: []string{output},
		Revision:  revision,
	}

	// Extract verdict
	if m := verdictRe.FindStringSubmatch(output); len(m) >= 2 {
		switch strings.ToUpper(m[1]) {
		case "PASS":
			result.Verdict = VerdictPass
		case "REVISE":
			result.Verdict = VerdictRevise
		case "REJECT":
			result.Verdict = VerdictReject
		}
	}

	// Extract findings
	for _, m := range findingRe.FindAllStringSubmatch(output, -1) {
		if len(m) >= 3 {
			result.Findings = append(result.Findings, ReviewFinding{
				Provider:    provider,
				Severity:    strings.ToLower(m[1]),
				Description: strings.TrimSpace(m[2]),
			})
		}
	}

	return result
}

// MergeVerdicts combines multiple review results into a single verdict.
// REJECT wins over REVISE, REVISE wins over PASS.
func MergeVerdicts(results []ReviewResult) ReviewVerdict {
	verdict := VerdictPass
	for _, r := range results {
		switch r.Verdict {
		case VerdictReject:
			return VerdictReject
		case VerdictRevise:
			verdict = VerdictRevise
		}
	}
	return verdict
}

// PersistReview writes a ReviewResult to review.md in the given directory.
func PersistReview(dir string, result *ReviewResult) error {
	content := formatReviewMd(result)
	path := filepath.Join(dir, "review.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write review.md: %w", err)
	}
	return nil
}

// formatReviewMd formats a ReviewResult as Markdown.
func formatReviewMd(r *ReviewResult) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Review: %s\n\n", r.SpecID)
	fmt.Fprintf(&sb, "**Verdict**: %s\n", r.Verdict)
	fmt.Fprintf(&sb, "**Revision**: %d\n", r.Revision)
	fmt.Fprintf(&sb, "**Date**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	if len(r.Findings) > 0 {
		sb.WriteString("## Findings\n\n")
		sb.WriteString("| Provider | Severity | Description |\n")
		sb.WriteString("|----------|----------|-------------|\n")
		for _, f := range r.Findings {
			fmt.Fprintf(&sb, "| %s | %s | %s |\n", f.Provider, f.Severity, f.Description)
		}
		sb.WriteString("\n")
	}

	if len(r.Responses) > 0 {
		sb.WriteString("## Provider Responses\n\n")
		for i, resp := range r.Responses {
			fmt.Fprintf(&sb, "### Response %d\n\n", i+1)
			sb.WriteString(resp)
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}
