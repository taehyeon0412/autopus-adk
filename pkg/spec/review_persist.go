package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
