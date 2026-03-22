package constraint_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/insajin/autopus-adk/pkg/constraint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleYAML = `
deny:
  - pattern: "fmt.Println"
    reason: "use structured logging instead"
    suggest: "use slog.Info or logger.Info"
    category: convention
  - pattern: "os.Exit"
    reason: "prevents deferred cleanup"
    suggest: "return errors up the call stack"
    category: convention
  - pattern: "SELECT *"
    reason: "fetches unused columns"
    suggest: "list columns explicitly"
    category: performance
`

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "constraints.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestLoad_FileNotExist(t *testing.T) {
	reg, err := constraint.Load("/nonexistent/path/constraints.yaml")
	require.NoError(t, err)
	assert.True(t, reg.IsEmpty())
}

func TestLoad_ValidYAML(t *testing.T) {
	path := writeTempYAML(t, sampleYAML)
	reg, err := constraint.Load(path)
	require.NoError(t, err)
	assert.False(t, reg.IsEmpty())
	assert.Len(t, reg.Constraints(), 3)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte(":\t invalid"), 0o644))

	_, err := constraint.Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse constraints YAML")
}

func TestLoadFromDir(t *testing.T) {
	// Build the directory tree: <tmp>/.autopus/context/constraints.yaml
	base := t.TempDir()
	dir := filepath.Join(base, ".autopus", "context")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "constraints.yaml"), []byte(sampleYAML), 0o644))

	reg, err := constraint.LoadFromDir(base)
	require.NoError(t, err)
	assert.Len(t, reg.Constraints(), 3)
}

func TestLoadFromDir_MissingFile(t *testing.T) {
	reg, err := constraint.LoadFromDir(t.TempDir())
	require.NoError(t, err)
	assert.True(t, reg.IsEmpty())
}

func TestByCategory(t *testing.T) {
	path := writeTempYAML(t, sampleYAML)
	reg, err := constraint.Load(path)
	require.NoError(t, err)

	convention := reg.ByCategory(constraint.CategoryConvention)
	assert.Len(t, convention, 2)

	perf := reg.ByCategory(constraint.CategoryPerformance)
	assert.Len(t, perf, 1)
	assert.Equal(t, "SELECT *", perf[0].Pattern)

	security := reg.ByCategory(constraint.CategorySecurity)
	assert.Empty(t, security)
}

func TestGeneratePromptText_Empty(t *testing.T) {
	reg, err := constraint.Load("/nonexistent/constraints.yaml")
	require.NoError(t, err)
	assert.Equal(t, "", reg.GeneratePromptText())
}

func TestLoad_UnreadableFile(t *testing.T) {
	t.Parallel()
	// Create a file with no read permissions to trigger a non-NotExist read error.
	dir := t.TempDir()
	path := filepath.Join(dir, "constraints.yaml")
	require.NoError(t, os.WriteFile(path, []byte(sampleYAML), 0o644))
	require.NoError(t, os.Chmod(path, 0o000))
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	// Skip on systems where root bypasses permission checks.
	if os.Getuid() == 0 {
		t.Skip("running as root; permission check not enforced")
	}

	_, err := constraint.Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read constraints file")
}

func TestGeneratePromptText_Content(t *testing.T) {
	path := writeTempYAML(t, sampleYAML)
	reg, err := constraint.Load(path)
	require.NoError(t, err)

	text := reg.GeneratePromptText()
	assert.True(t, strings.HasPrefix(text, "# Anti-Pattern Constraints\n"))
	assert.Contains(t, text, "NEVER do the following")
	assert.Contains(t, text, "fmt.Println")
	assert.Contains(t, text, "use structured logging instead")
	assert.Contains(t, text, "use slog.Info or logger.Info")
	assert.Contains(t, text, "convention")
	// All three constraints numbered
	assert.Contains(t, text, "1. **NEVER**")
	assert.Contains(t, text, "2. **NEVER**")
	assert.Contains(t, text, "3. **NEVER**")
}
