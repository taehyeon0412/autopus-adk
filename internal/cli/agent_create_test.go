package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateAgentFrontmatter_Valid verifies that a well-formed agent
// frontmatter passes validation without error.
func TestValidateAgentFrontmatter_Valid(t *testing.T) {
	t.Parallel()

	content := []byte("---\nname: my-agent\ndescription: Does something useful\n---\n")

	err := validateAgentFrontmatter(content)
	assert.NoError(t, err)
}

// TestValidateAgentFrontmatter_MissingName verifies that a missing name field
// returns an appropriate error.
func TestValidateAgentFrontmatter_MissingName(t *testing.T) {
	t.Parallel()

	content := []byte("---\ndescription: Does something useful\n---\n")

	err := validateAgentFrontmatter(content)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

// TestValidateAgentFrontmatter_MissingDescription verifies that a missing
// description field returns an error.
func TestValidateAgentFrontmatter_MissingDescription(t *testing.T) {
	t.Parallel()

	content := []byte("---\nname: my-agent\n---\n")

	err := validateAgentFrontmatter(content)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "description")
}

// TestValidateAgentFrontmatter_NoDelimiter verifies that content without ---
// delimiters returns an error.
func TestValidateAgentFrontmatter_NoDelimiter(t *testing.T) {
	t.Parallel()

	content := []byte("name: my-agent\ndescription: Does something\n")

	err := validateAgentFrontmatter(content)
	require.Error(t, err)
}

// TestParseTools_CommaSeparated verifies that comma-separated tool names are
// split, trimmed, and returned correctly.
func TestParseTools_CommaSeparated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single tool",
			input:    "Read",
			expected: []string{"Read"},
		},
		{
			name:     "multiple tools",
			input:    "Read,Write,Bash",
			expected: []string{"Read", "Write", "Bash"},
		},
		{
			name:     "tools with spaces",
			input:    "Read, Write, Bash",
			expected: []string{"Read", "Write", "Bash"},
		},
		{
			name:     "empty string returns defaults",
			input:    "",
			expected: []string{"Read", "Write", "Bash"},
		},
		{
			name:     "whitespace-only entries dropped",
			input:    "Read,,Bash",
			expected: []string{"Read", "Bash"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseTools(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
