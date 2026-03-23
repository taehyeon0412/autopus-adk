package issue

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/insajin/autopus-adk/templates"
)

const (
	// maxBodyBytes is the GitHub issue body size limit.
	maxBodyBytes    = 65536
	truncatedMarker = "\n... [truncated]"
)

// FormatMarkdown renders the issue report as a markdown string using the
// embedded issue-report.md.tmpl template.
func FormatMarkdown(report IssueReport) (string, error) {
	tmplData, err := templates.FS.ReadFile("shared/issue-report.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("formatter: read template: %w", err)
	}

	tmpl, err := template.New("issue-report").Parse(string(tmplData))
	if err != nil {
		return "", fmt.Errorf("formatter: parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, report); err != nil {
		return "", fmt.Errorf("formatter: render template: %w", err)
	}

	body := buf.String()
	if len(body) > maxBodyBytes {
		body = body[:maxBodyBytes] + truncatedMarker
	}

	return body, nil
}
