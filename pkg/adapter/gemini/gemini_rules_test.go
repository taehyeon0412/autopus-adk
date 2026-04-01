package gemini_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter/gemini"
	"github.com/insajin/autopus-adk/pkg/config"
)

func TestGeminiGenerateRules(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := gemini.NewWithRoot(dir)
	cfg := config.DefaultFullConfig("test-project")

	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	// Verify all 4 rule files are created
	rules := []string{
		"lore-commit.md",
		"file-size-limit.md",
		"subagent-delegation.md",
		"language-policy.md",
	}
	rulesDir := filepath.Join(dir, ".gemini", "rules", "autopus")
	for _, rule := range rules {
		rulePath := filepath.Join(rulesDir, rule)
		_, statErr := os.Stat(rulePath)
		assert.NoError(t, statErr, "rule file should exist: %s", rule)
	}
}

func TestGeminiRulesImport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := gemini.NewWithRoot(dir)
	cfg := config.DefaultFullConfig("test-project")

	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	// Check GEMINI.md contains @import references for rules
	geminiMDPath := filepath.Join(dir, "GEMINI.md")
	data, err := os.ReadFile(geminiMDPath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "@.gemini/rules/autopus/lore-commit.md",
		"GEMINI.md should have @import for lore-commit")
	assert.Contains(t, content, "@.gemini/rules/autopus/file-size-limit.md",
		"GEMINI.md should have @import for file-size-limit")
	assert.Contains(t, content, "@.gemini/rules/autopus/subagent-delegation.md",
		"GEMINI.md should have @import for subagent-delegation")
	assert.Contains(t, content, "@.gemini/rules/autopus/language-policy.md",
		"GEMINI.md should have @import for language-policy")
}

func TestGeminiRulesContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := gemini.NewWithRoot(dir)
	cfg := config.DefaultFullConfig("test-project")

	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)

	// Verify lore-commit rule has key content
	lorePath := filepath.Join(dir, ".gemini", "rules", "autopus", "lore-commit.md")
	data, err := os.ReadFile(lorePath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "Lore Commit", "should contain rule title")
	assert.Contains(t, content, "platform: gemini-cli",
		"should have gemini-cli platform in frontmatter")

	// Verify file-size-limit rule exists and has key content
	fsPath := filepath.Join(dir, ".gemini", "rules", "autopus", "file-size-limit.md")
	fsData, err := os.ReadFile(fsPath)
	require.NoError(t, err)
	assert.Contains(t, string(fsData), "300 lines",
		"file-size-limit should reference 300 lines")
}
