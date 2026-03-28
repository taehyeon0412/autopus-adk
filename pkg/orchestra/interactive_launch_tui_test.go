package orchestra

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// SPEC-ORCH-014 R2: buildInteractiveLaunchCmd for opencode with empty
// InteractiveInput must produce "opencode -m openai/gpt-5.4" (no "run").

// TestBuildInteractiveLaunchCmd_OpencodeTUI_NoRun verifies opencode TUI mode
// (InteractiveInput="") produces a command without "run" subcmd.
// RED: this tests the NEW expected behavior after R1 config changes.
// With PaneArgs=["-m", "openai/gpt-5.4"] and empty InteractiveInput,
// the command should be "opencode -m openai/gpt-5.4".
func TestBuildInteractiveLaunchCmd_OpencodeTUI_NoRun(t *testing.T) {
	t.Parallel()

	p := ProviderConfig{
		Name:             "opencode",
		Binary:           "opencode",
		PaneArgs:         []string{"-m", "openai/gpt-5.4"},
		InteractiveInput: "", // TUI mode: sendkeys
	}

	cmd := buildInteractiveLaunchCmd(p, "fix the bug")
	assert.Equal(t, "opencode -m openai/gpt-5.4", cmd,
		"opencode TUI mode must not contain 'run' and must not append prompt (R2)")
}

// TestBuildInteractiveLaunchCmd_OpencodeTUI_PromptNotAppended verifies that
// in TUI mode (InteractiveInput=""), the prompt is NOT appended as CLI arg.
// Prompt delivery happens via SendLongText instead.
func TestBuildInteractiveLaunchCmd_OpencodeTUI_PromptNotAppended(t *testing.T) {
	t.Parallel()

	p := ProviderConfig{
		Name:             "opencode",
		Binary:           "opencode",
		PaneArgs:         []string{"-m", "openai/gpt-5.4"},
		InteractiveInput: "", // TUI mode
	}

	cmd := buildInteractiveLaunchCmd(p, "explain this function")
	assert.NotContains(t, cmd, "explain this function",
		"TUI mode must not append prompt as CLI arg (R2)")
	assert.NotContains(t, cmd, "run",
		"TUI mode must not contain 'run' subcmd (R2)")
}

// TestBuildInteractiveLaunchCmd_OpencodeTUI_SessionPersist verifies opencode
// in TUI mode produces a persistent session command (like claude/gemini).
// The command should just launch the TUI without "run" (which exits after one task).
func TestBuildInteractiveLaunchCmd_OpencodeTUI_SessionPersist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		paneArgs  []string
		prompt    string
		expectCmd string
	}{
		{
			name:      "standard TUI launch",
			paneArgs:  []string{"-m", "openai/gpt-5.4"},
			prompt:    "review code",
			expectCmd: "opencode -m openai/gpt-5.4",
		},
		{
			name:      "empty prompt",
			paneArgs:  []string{"-m", "openai/gpt-5.4"},
			prompt:    "",
			expectCmd: "opencode -m openai/gpt-5.4",
		},
		{
			name:      "different model",
			paneArgs:  []string{"-m", "openai/o3"},
			prompt:    "test",
			expectCmd: "opencode -m openai/o3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := ProviderConfig{
				Name:             "opencode",
				Binary:           "opencode",
				PaneArgs:         tt.paneArgs,
				InteractiveInput: "", // TUI mode
			}

			cmd := buildInteractiveLaunchCmd(p, tt.prompt)
			assert.Equal(t, tt.expectCmd, cmd,
				"opencode TUI mode must produce persistent session command (R2)")
		})
	}
}
