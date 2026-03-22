package content_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/content"
)

func TestGenerateConstraintInstruction_Disabled(t *testing.T) {
	t.Parallel()

	cfg := config.ConstraintConf{Enabled: false}
	result := content.GenerateConstraintInstruction(t.TempDir(), cfg)
	assert.Empty(t, result)
}

func TestGenerateConstraintInstruction_MissingFile(t *testing.T) {
	t.Parallel()

	// File does not exist — should return empty without error.
	cfg := config.ConstraintConf{Enabled: true}
	result := content.GenerateConstraintInstruction(t.TempDir(), cfg)
	assert.Empty(t, result)
}

func TestGenerateConstraintInstruction_EmptyPatterns(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	yamlContent := "deny: []\n"
	writeConstraintsYAML(t, dir, yamlContent)

	cfg := config.ConstraintConf{Enabled: true}
	result := content.GenerateConstraintInstruction(dir, cfg)
	assert.Empty(t, result)
}

func TestGenerateConstraintInstruction_WithPatterns(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	yamlContent := `deny:
  - pattern: "panic("
    reason: "unrecovered panic crashes the process"
    suggest: "return error instead"
    category: convention
`
	writeConstraintsYAML(t, dir, yamlContent)

	cfg := config.ConstraintConf{Enabled: true}
	result := content.GenerateConstraintInstruction(dir, cfg)

	assert.Contains(t, result, "Anti-Pattern Constraints")
	assert.Contains(t, result, "panic(")
	assert.Contains(t, result, "unrecovered panic")
}

func TestGenerateConstraintInstruction_CustomPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	custom := filepath.Join(dir, "custom-constraints.yaml")
	yamlContent := `deny:
  - pattern: "TODO"
    reason: "unresolved work item"
    suggest: "open an issue instead"
    category: convention
`
	require.NoError(t, os.WriteFile(custom, []byte(yamlContent), 0o644))

	cfg := config.ConstraintConf{Enabled: true, Path: custom}
	result := content.GenerateConstraintInstruction(dir, cfg)

	assert.Contains(t, result, "TODO")
}

// writeConstraintsYAML creates the default constraints.yaml at the expected path.
func writeConstraintsYAML(t *testing.T, projectDir, yamlContent string) {
	t.Helper()

	path := filepath.Join(projectDir, ".autopus", "context")
	require.NoError(t, os.MkdirAll(path, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(path, "constraints.yaml"), []byte(yamlContent), 0o644))
}
