package worker

import (
	"strings"
	"testing"
)

func TestContextBuilder_FullPayload(t *testing.T) {
	b := &ContextBuilder{}
	result := b.Build(TaskPayload{
		TaskID:        "task-001",
		Description:   "Implement feature X",
		PMNotes:       "Priority: high",
		PolicySummary: "No network access",
		KnowledgeCtx:  "Related to module Y",
		SpecID:        "SPEC-TEST-001",
	})

	for _, want := range []string{
		"# Task: task-001",
		"## Description",
		"Implement feature X",
		"## PM Notes",
		"Priority: high",
		"## Security Policy",
		"No network access",
		"## Knowledge Context",
		"Related to module Y",
		"## Reference",
		"Spec: SPEC-TEST-001",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestContextBuilder_MinimalPayload(t *testing.T) {
	b := &ContextBuilder{}
	result := b.Build(TaskPayload{
		TaskID:      "task-002",
		Description: "Fix bug",
	})

	if !strings.Contains(result, "# Task: task-002") {
		t.Error("missing task header")
	}
	if !strings.Contains(result, "Fix bug") {
		t.Error("missing description")
	}

	// Optional sections should be absent.
	for _, absent := range []string{
		"## PM Notes",
		"## Security Policy",
		"## Knowledge Context",
		"## Reference",
	} {
		if strings.Contains(result, absent) {
			t.Errorf("output should not contain %q for minimal payload", absent)
		}
	}
}

func TestContextBuilder_PartialPayload(t *testing.T) {
	b := &ContextBuilder{}
	result := b.Build(TaskPayload{
		TaskID:        "task-003",
		Description:   "Refactor module",
		PolicySummary: "Read-only FS",
		SpecID:        "SPEC-REF-001",
	})

	if !strings.Contains(result, "## Security Policy") {
		t.Error("missing security policy section")
	}
	if !strings.Contains(result, "Spec: SPEC-REF-001") {
		t.Error("missing spec reference")
	}
	if strings.Contains(result, "## PM Notes") {
		t.Error("PM Notes should be absent when empty")
	}
	if strings.Contains(result, "## Knowledge Context") {
		t.Error("Knowledge Context should be absent when empty")
	}
}
