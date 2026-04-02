package worker

import (
	"fmt"
	"strings"
)

// TaskPayload contains backend-provided task data for prompt assembly.
type TaskPayload struct {
	TaskID        string
	Description   string
	PMNotes       string // PM instructions (optional)
	PolicySummary string // security policy summary
	KnowledgeCtx  string // Knowledge Hub context (optional)
	SpecID        string // SPEC reference (optional)
}

// ContextBuilder assembles the Layer 4 prompt for subprocess execution.
type ContextBuilder struct{}

// Build assembles the complete prompt string for stdin injection.
// Only non-empty sections are included in the output.
func (b *ContextBuilder) Build(payload TaskPayload) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Task: %s\n\n", payload.TaskID)
	sb.WriteString("## Description\n\n")
	sb.WriteString(payload.Description)
	sb.WriteString("\n")

	if payload.PMNotes != "" {
		sb.WriteString("\n## PM Notes\n\n")
		sb.WriteString(payload.PMNotes)
		sb.WriteString("\n")
	}

	if payload.PolicySummary != "" {
		sb.WriteString("\n## Security Policy\n\n")
		sb.WriteString(payload.PolicySummary)
		sb.WriteString("\n")
	}

	if payload.KnowledgeCtx != "" {
		sb.WriteString("\n## Knowledge Context\n\n")
		sb.WriteString(payload.KnowledgeCtx)
		sb.WriteString("\n")
	}

	if payload.SpecID != "" {
		fmt.Fprintf(&sb, "\n## Reference\n\nSpec: %s\n", payload.SpecID)
	}

	return sb.String()
}
