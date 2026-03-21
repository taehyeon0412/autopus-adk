package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BuildReviewPrompt constructs a review prompt from a SPEC document and code context.
func BuildReviewPrompt(doc *SpecDocument, codeContext string) string {
	var sb strings.Builder

	sb.WriteString("You are reviewing a SPEC document for correctness, completeness, and feasibility.\n\n")
	fmt.Fprintf(&sb, "## SPEC: %s — %s\n\n", doc.ID, doc.Title)

	if len(doc.Requirements) > 0 {
		sb.WriteString("### Requirements\n\n")
		for _, req := range doc.Requirements {
			fmt.Fprintf(&sb, "- **%s** [%s]: %s\n", req.ID, req.Type, req.Description)
		}
		sb.WriteString("\n")
	}

	if len(doc.AcceptanceCriteria) > 0 {
		sb.WriteString("### Acceptance Criteria\n\n")
		for _, ac := range doc.AcceptanceCriteria {
			fmt.Fprintf(&sb, "- %s: %s\n", ac.ID, ac.Description)
		}
		sb.WriteString("\n")
	}

	if codeContext != "" {
		sb.WriteString("### Existing Code Context\n\n")
		sb.WriteString("```\n")
		sb.WriteString(codeContext)
		sb.WriteString("\n```\n\n")
	}

	sb.WriteString("### Instructions\n\n")
	sb.WriteString("Review the SPEC and respond with:\n")
	sb.WriteString("1. VERDICT: PASS, REVISE, or REJECT\n")
	sb.WriteString("2. For each issue found, write: FINDING: [severity] description\n")
	sb.WriteString("   Severity levels: critical, major, minor, suggestion\n")
	sb.WriteString("3. Provide reasoning for your verdict.\n")

	return sb.String()
}

// CollectContext recursively reads source files from a directory up to maxLines total.
func CollectContext(dir string, maxLines int) (string, error) {
	var sb strings.Builder
	lineCount := 0

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			return nil
		}
		if lineCount >= maxLines {
			return filepath.SkipAll
		}
		if !isSourceFile(d.Name()) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		if relPath == "" {
			relPath = d.Name()
		}

		lines := strings.Split(string(content), "\n")
		remaining := maxLines - lineCount
		if remaining <= 0 {
			return filepath.SkipAll
		}

		fmt.Fprintf(&sb, "--- %s ---\n", relPath)
		lineCount++

		end := min(len(lines), remaining)
		for _, line := range lines[:end] {
			sb.WriteString(line)
			sb.WriteString("\n")
			lineCount++
		}
		sb.WriteString("\n")
		lineCount++
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	return sb.String(), nil
}

// isSourceFile returns true if the filename is a recognized source file.
func isSourceFile(name string) bool {
	exts := []string{".go", ".py", ".ts", ".js", ".rs", ".java", ".rb"}
	for _, ext := range exts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
