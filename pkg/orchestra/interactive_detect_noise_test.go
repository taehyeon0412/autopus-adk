package orchestra

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- SPEC-ORCH-013 R4: OpenCode Output Refinement ---

// TestIsPromptLine_ShellLoginBanner verifies shell login banner filtering.
// S7: "Last login:" lines must be filtered as noise.
func TestIsPromptLine_ShellLoginBanner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"Last login standard", "Last login: Fri Mar 28 10:00:00 on ttys001", true},
		{"Last login lowercase", "last login: Fri Mar 28 10:00:00 on ttys001", true},
		{"Last login with extra spaces", "  Last login: Fri Mar 28 09:00:00 on ttys002  ", true},
		{"Not a login banner", "Last modified: yesterday", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isPromptLine(tt.line)
			assert.Equal(t, tt.expected, got,
				"shell login banner pattern must be recognized as noise")
		})
	}
}

// TestIsPromptLine_UserAtHostPrompt verifies user@host prompt filtering.
// S8: Lines matching "user@hostname $" pattern must be filtered.
func TestIsPromptLine_UserAtHostPrompt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"user@host dollar", "user@hostname $ ", true},
		{"root@server hash", "root@server.local # ", true},
		{"user@host percent", "dev@macbook % ", true},
		{"user@host no space", "user@host$", true},
		{"not a prompt", "email@example.com is my email", false},
		{"content with at sign", "send to admin@server for help", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isPromptLine(tt.line)
			assert.Equal(t, tt.expected, got,
				"user@host prompt pattern must be recognized as noise")
		})
	}
}

// TestIsPromptLine_OpencodeTUIChrome verifies opencode TUI chrome filtering.
// R4: opencode TUI chrome patterns must be filtered as noise.
func TestIsPromptLine_OpencodeTUIChrome(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"opencode session header", "Build · gpt-5.4 · OpenAI", true},
		{"opencode build line", "  Build GPT-5.4 OpenAI", true},
		{"opencode escape hint", "⬝⬝⬝ esc to cancel", true},
		{"opencode ctrl hint", "ctrl+c to quit", true},
		{"normal output", "The build process completed successfully", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isPromptLine(tt.line)
			assert.Equal(t, tt.expected, got,
				"opencode TUI chrome must be recognized as noise")
		})
	}
}
