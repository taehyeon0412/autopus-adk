package pipeline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateBranchName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{"valid feature branch", "feature/my-branch", false},
		{"valid worktree branch", "worktree/SPEC-001/T1", false},
		{"valid mixed chars", "a/b.c-d_e", false},
		{"empty string detach mode", "", false},
		{"semicolon injection", "main; rm -rf /", true},
		{"subshell injection", "branch$(whoami)", true},
		{"double dot traversal", "..exploit", true},
		{"dash prefix flag injection", "-flag-injection", true},
		{"exceeds 255 chars", strings.Repeat("a", 256), true},
		{"space in name", "my branch", true},
		{"backtick in name", "branch`id`", true},
		{"pipe in name", "branch|evil", true},
		{"ampersand in name", "branch&&evil", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateBranchName(tt.branch)

			if tt.wantErr {
				require.Error(t, err, "expected error for branch: %q", tt.branch)
				assert.Contains(t, err.Error(), "invalid",
					"error message should mention 'invalid' for: %q", tt.branch)
			} else {
				assert.NoError(t, err, "expected no error for branch: %q", tt.branch)
			}
		})
	}
}

func TestSanitizeBranchName_ReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"double dot prefix", "..exploit", "", true},
		{"dash prefix", "-flag", "", true},
		{"exceeds 255 chars", strings.Repeat("x", 256), "", true},
		{"valid name passthrough", "valid-name", "valid-name", false},
		{"valid with slash", "feature/branch", "feature/branch", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sanitizeBranchName(tt.input)

			if tt.wantErr {
				require.Error(t, err, "expected error for input: %q", tt.input)
			} else {
				require.NoError(t, err, "expected no error for input: %q", tt.input)
				assert.Equal(t, tt.want, got,
					"sanitized result mismatch for: %q", tt.input)
			}
		})
	}
}

func TestValidateBranchName_ExactBoundary(t *testing.T) {
	t.Parallel()

	// 255 chars should be valid (boundary)
	name255 := strings.Repeat("a", 255)
	err := ValidateBranchName(name255)
	assert.NoError(t, err, "255-char branch name should be valid")

	// 256 chars should be invalid
	name256 := strings.Repeat("a", 256)
	err = ValidateBranchName(name256)
	require.Error(t, err, "256-char branch name should be invalid")
}

func TestValidateBranchName_DoubleDotMiddle(t *testing.T) {
	t.Parallel()

	// Double dot anywhere in name should be rejected.
	err := ValidateBranchName("feature/..exploit/main")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "..")
}

func TestValidateBranchName_ErrorMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		branch    string
		wantInErr string
	}{
		{"length error includes count", strings.Repeat("x", 300), "300"},
		{"dash prefix error", "-bad", "'-'"},
		{"double-dot error", "a..b", "'..'"},
		{"regex error includes name", "bad name!", "bad name!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateBranchName(tt.branch)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantInErr)
		})
	}
}

func TestSanitizeBranchName_RejectsInvalidCharsViaValidation(t *testing.T) {
	t.Parallel()

	// sanitizeBranchName calls ValidateBranchName first, so chars outside
	// the allowed regex are rejected before replacement happens.
	invalidInputs := []struct {
		name  string
		input string
	}{
		{"space", "my branch"},
		{"tilde", "branch~1"},
		{"caret", "branch^2"},
		{"colon", "branch:ref"},
		{"question mark", "branch?"},
		{"asterisk", "branch*"},
		{"backslash", "branch\\path"},
		{"bracket", "branch[0]"},
	}

	for _, tt := range invalidInputs {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := sanitizeBranchName(tt.input)
			require.Error(t, err, "should reject %q via ValidateBranchName", tt.input)
		})
	}
}

func TestSanitizeBranchName_PassthroughValidNames(t *testing.T) {
	t.Parallel()

	// Valid names should pass through unchanged.
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "main", "main"},
		{"with slash", "feature/foo", "feature/foo"},
		{"with dot", "v1.0.0", "v1.0.0"},
		{"with underscore", "my_branch", "my_branch"},
		{"with dash", "my-branch", "my-branch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sanitizeBranchName(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
