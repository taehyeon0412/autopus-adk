package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteReactReport_CreatesFile verifies that writeReactReport creates a
// markdown file at the given path containing run metadata.
func TestWriteReactReport_CreatesFile(t *testing.T) {
	t.Parallel()

	// Given: a temp directory and a ciRun with known fields
	dir := t.TempDir()
	path := filepath.Join(dir, "12345.md")
	run := ciRun{
		DatabaseID: 12345,
		Name:       "CI Build",
		HeadBranch: "main",
		Conclusion: "failure",
		UpdatedAt:  "2026-03-24T12:00:00Z",
	}
	logs := "Error: test failed at line 42"

	// When: writeReactReport is called
	err := writeReactReport(path, run, logs)
	require.NoError(t, err)

	// Then: the file exists and contains the run metadata
	data, readErr := os.ReadFile(path)
	require.NoError(t, readErr)

	content := string(data)
	assert.Contains(t, content, "12345")
	assert.Contains(t, content, "CI Build")
	assert.Contains(t, content, "main")
	assert.Contains(t, content, "failure")
	assert.Contains(t, content, "Error: test failed at line 42")
}

// TestWriteReactReport_NonExistentDir_ReturnsError verifies that
// writeReactReport returns an error when the parent directory does not exist.
func TestWriteReactReport_NonExistentDir_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: a path inside a non-existent directory
	path := filepath.Join(t.TempDir(), "no-such-dir", "report.md")
	run := ciRun{DatabaseID: 1}

	// When: writeReactReport is called
	err := writeReactReport(path, run, "logs")

	// Then: an error is returned
	require.Error(t, err)
}

// TestExtractReportSummary_ExtractsHeaderAndBullets verifies that
// extractReportSummary returns lines starting with # and - **.
func TestExtractReportSummary_ExtractsHeaderAndBullets(t *testing.T) {
	t.Parallel()

	// Given: a report with header, bullet metadata, and body
	report := `# CI Failure Report

- **Run ID**: 12345
- **Name**: CI Build
- **Branch**: main
- **Conclusion**: failure
- **Updated**: 2026-03-24T12:00:00Z
- **Generated**: 2026-03-24T12:01:00Z

## Failure Logs

` + "```" + `
Error: test failed
` + "```" + `
`

	// When: extractReportSummary is called
	summary := extractReportSummary(report)

	// Then: header and metadata lines are present, body is omitted
	assert.Contains(t, summary, "# CI Failure Report")
	assert.Contains(t, summary, "- **Run ID**: 12345")
	assert.Contains(t, summary, "- **Name**: CI Build")
	assert.NotContains(t, summary, "Error: test failed")
}

// TestExtractReportSummary_EmptyContent verifies that extractReportSummary
// returns an empty string for empty input.
func TestExtractReportSummary_EmptyContent(t *testing.T) {
	t.Parallel()

	summary := extractReportSummary("")
	assert.Equal(t, "", summary)
}

// TestExtractReportSummary_LimitsTo6Lines verifies that extractReportSummary
// returns at most 6 matching lines.
func TestExtractReportSummary_LimitsTo6Lines(t *testing.T) {
	t.Parallel()

	// Given: a report with more than 6 metadata lines
	report := "# Title\n- **A**: 1\n- **B**: 2\n- **C**: 3\n- **D**: 4\n- **E**: 5\n- **F**: 6\n- **G**: 7\n"

	// When: extractReportSummary is called
	summary := extractReportSummary(report)
	lines := splitNonEmpty(summary)

	// Then: at most 6 lines are returned
	assert.LessOrEqual(t, len(lines), 6)
}

// splitNonEmpty is a test helper that splits a string by newlines, ignoring empty lines.
func splitNonEmpty(s string) []string {
	var result []string
	for _, line := range splitLines(s) {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	result := make([]string, 0)
	start := 0
	for i, c := range s {
		if c == '\n' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
