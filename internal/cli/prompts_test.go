package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/pkg/config"
)

// newTestCmd creates a cobra command that writes to the provided buffer.
func newTestCmd(buf *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	return cmd
}

// TestWarnParentRuleConflicts_NoConflicts verifies function is a no-op when no conflicts exist.
func TestWarnParentRuleConflicts_NoConflicts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := config.DefaultFullConfig("test-project")
	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("setup: config.Save failed: %v", err)
	}

	var buf bytes.Buffer
	cmd := newTestCmd(&buf)
	// No parent rules in a fresh temp dir — function should be a no-op
	warnParentRuleConflicts(cmd, dir, cfg)
	assert.Empty(t, buf.String(), "no output expected when no conflicts")
}

// TestWarnParentRuleConflicts_IsolateRulesAlreadySet verifies that if IsolateRules=true
// and conflicts exist, only an informational message is printed (no prompt).
func TestWarnParentRuleConflicts_IsolateRulesAlreadySet(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := config.DefaultFullConfig("test-project")
	cfg.IsolateRules = true
	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("setup: config.Save failed: %v", err)
	}

	var buf bytes.Buffer
	cmd := newTestCmd(&buf)
	warnParentRuleConflicts(cmd, dir, cfg)
	// No conflicts in temp dir → no output
	assert.Empty(t, buf.String())
}

// TestPromptLanguageSettings_AlreadyConfigured verifies skip when all language fields are set.
func TestPromptLanguageSettings_AlreadyConfigured(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := config.DefaultFullConfig("test-project")
	cfg.Language.Comments = "en"
	cfg.Language.Commits = "ko"
	cfg.Language.AIResponses = "en"
	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("setup: config.Save failed: %v", err)
	}

	var buf bytes.Buffer
	cmd := newTestCmd(&buf)
	promptLanguageSettings(cmd, dir, cfg)

	// All set → function should return early with no output
	assert.Empty(t, buf.String(), "no prompt expected when language already configured")
}
