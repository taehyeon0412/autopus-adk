package spec

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// REQ-008: Static analysis integration — golangci-lint JSON output parsing

const sampleGolangciOutput = `{"Issues":[
	{"Text":"exported function Foo should have comment","FromLinter":"revive","Pos":{"Filename":"pkg/spec/types.go","Line":12}},
	{"Text":"var x is unused","FromLinter":"deadcode","Pos":{"Filename":"pkg/spec/resolver.go","Line":55}},
	{"Text":"cyclomatic complexity 15 of function Bar","FromLinter":"gocyclo","Pos":{"Filename":"internal/cli/run.go","Line":100}}
]}`

func TestParseGolangciOutput_ValidJSON(t *testing.T) {
	t.Parallel()

	findings, err := ParseGolangciOutput([]byte(sampleGolangciOutput))
	require.NoError(t, err)
	require.Len(t, findings, 3)

	// All must be style category
	for _, f := range findings {
		assert.Equal(t, FindingCategoryStyle, f.Category, "golangci-lint findings must be style category")
	}
}

func TestParseGolangciOutput_ScopeRefContainsFileAndLine(t *testing.T) {
	t.Parallel()

	findings, err := ParseGolangciOutput([]byte(sampleGolangciOutput))
	require.NoError(t, err)
	require.NotEmpty(t, findings)

	// First finding: types.go:12
	assert.Contains(t, findings[0].ScopeRef, "pkg/spec/types.go")
	assert.Contains(t, findings[0].ScopeRef, "12", "ScopeRef must include line number")
}

func TestParseGolangciOutput_DescriptionContainsLinterAndText(t *testing.T) {
	t.Parallel()

	findings, err := ParseGolangciOutput([]byte(sampleGolangciOutput))
	require.NoError(t, err)
	require.NotEmpty(t, findings)

	// Description must contain both linter name and issue text
	assert.Contains(t, findings[0].Description, "revive", "description must include linter name")
	assert.Contains(t, findings[0].Description, "exported function Foo", "description must include issue text")
}

func TestParseGolangciOutput_EmptyIssues(t *testing.T) {
	t.Parallel()

	input := `{"Issues":[]}`
	findings, err := ParseGolangciOutput([]byte(input))
	require.NoError(t, err)
	assert.Empty(t, findings, "empty issues array must yield no findings")
}

func TestParseGolangciOutput_MalformedJSON(t *testing.T) {
	t.Parallel()

	findings, err := ParseGolangciOutput([]byte("not json at all"))
	assert.Error(t, err, "malformed JSON must return error")
	assert.Nil(t, findings)
}

// Binary not installed graceful skip

func TestRunStaticAnalysis_BinaryMissingGracefulSkip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// When: golangci-lint is not in PATH (simulate by passing a non-existent binary name)
	findings, err := RunStaticAnalysis(dir, "nonexistent-linter-binary-xyz")
	// Must not propagate binary-not-found as a hard error
	require.NoError(t, err, "missing binary must not cause error — graceful skip expected")
	assert.Empty(t, findings, "missing binary must yield empty findings, not panic")
}

func TestRunStaticAnalysis_ReturnsStyleFindings(t *testing.T) {
	t.Parallel()

	// Skip if golangci-lint is not available in this environment
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		t.Skip("golangci-lint not installed, skipping integration test")
	}

	dir := t.TempDir()
	// Write a simple Go file with a known lint issue
	src := `package example

func unusedFunc() string { return "unused" }
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "example.go"), []byte(src), 0o644))

	findings, err := RunStaticAnalysis(dir, "golangci-lint")
	require.NoError(t, err)

	// All returned findings must be style category
	for _, f := range findings {
		assert.Equal(t, FindingCategoryStyle, f.Category)
	}
}

// Dedup with LLM findings

func TestDeduplicateWithLLMFindings_RemovesOverlap(t *testing.T) {
	t.Parallel()

	staticFindings := []ReviewFinding{
		{
			ID:          "F-001",
			Category:    FindingCategoryStyle,
			Description: "revive: exported function Bar should have comment",
			ScopeRef:    "pkg/spec/types.go:20",
			Provider:    "golangci-lint",
		},
	}
	llmFindings := []ReviewFinding{
		{
			ID:          "F-002",
			Category:    FindingCategoryStyle,
			Description: "Function Bar is missing documentation comment",
			ScopeRef:    "pkg/spec/types.go:20",
			Provider:    "claude",
		},
	}

	// When: static and LLM findings reference the same location with similar meaning
	merged := MergeStaticWithLLMFindings(staticFindings, llmFindings)

	// Then: deduplicated — same file:line findings should not double-count
	var atLine20 []ReviewFinding
	for _, f := range merged {
		if f.ScopeRef == "pkg/spec/types.go:20" {
			atLine20 = append(atLine20, f)
		}
	}
	assert.LessOrEqual(t, len(atLine20), 1, "same file:line finding from both static and LLM must be deduped to one")
}

func TestDeduplicateWithLLMFindings_KeepsBothWhenDifferentScope(t *testing.T) {
	t.Parallel()

	staticFindings := []ReviewFinding{
		{ID: "F-001", Category: FindingCategoryStyle, Description: "naming issue", ScopeRef: "pkg/a.go:10", Provider: "golangci-lint"},
	}
	llmFindings := []ReviewFinding{
		{ID: "F-002", Category: FindingCategoryCorrectness, Description: "nil dereference possible", ScopeRef: "pkg/b.go:55", Provider: "claude"},
	}

	merged := MergeStaticWithLLMFindings(staticFindings, llmFindings)

	assert.Len(t, merged, 2, "findings at different scopes must both be preserved")
}

func TestDeduplicateWithLLMFindings_EmptyStatic(t *testing.T) {
	t.Parallel()

	llmFindings := []ReviewFinding{
		{ID: "F-001", Category: FindingCategoryCompleteness, Description: "Missing acceptance criteria", ScopeRef: "REQ-003", Provider: "gemini"},
	}

	merged := MergeStaticWithLLMFindings(nil, llmFindings)
	assert.Len(t, merged, 1, "empty static findings must not drop LLM findings")
}

func TestDeduplicateWithLLMFindings_EmptyLLM(t *testing.T) {
	t.Parallel()

	staticFindings := []ReviewFinding{
		{ID: "F-001", Category: FindingCategoryStyle, Description: "revive: missing comment", ScopeRef: "cmd/run.go:5", Provider: "golangci-lint"},
	}

	merged := MergeStaticWithLLMFindings(staticFindings, nil)
	assert.Len(t, merged, 1, "empty LLM findings must not drop static findings")
}
