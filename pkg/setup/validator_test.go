package setup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_MissingDocs(t *testing.T) {
	t.Parallel()
	docsDir := t.TempDir() // empty dir
	projectDir := t.TempDir()

	report, err := Validate(docsDir, projectDir)
	require.NoError(t, err)

	assert.False(t, report.Valid)
	assert.NotEmpty(t, report.Warnings)
}

func TestValidate_ValidDocs(t *testing.T) {
	t.Parallel()
	projectDir := setupGoProject(t)
	docsDir := t.TempDir()

	// Create all doc files with valid content
	for _, fileName := range DocFiles {
		writeFile(t, docsDir, fileName, "# "+fileName+"\n\nSome content.\n")
	}
	writeFile(t, docsDir, ".meta.yaml", "generated_at: 2026-01-01T00:00:00Z\n")

	report, err := Validate(docsDir, projectDir)
	require.NoError(t, err)

	assert.True(t, report.Valid)
}

func TestValidate_StalePathReference(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	docsDir := t.TempDir()

	// Doc referencing a non-existent path
	writeFile(t, docsDir, "structure.md", "# Structure\n\nSee `internal/api/handler.go` for details.\n")
	// Create remaining docs
	for name, fileName := range DocFiles {
		if name != "structure" {
			writeFile(t, docsDir, fileName, "# "+fileName+"\n")
		}
	}
	writeFile(t, docsDir, ".meta.yaml", "generated_at: 2026-01-01T00:00:00Z\n")

	report, err := Validate(docsDir, projectDir)
	require.NoError(t, err)

	assert.False(t, report.Valid)
	var found bool
	for _, w := range report.Warnings {
		if w.Type == "stale_path" {
			found = true
			assert.Contains(t, w.Message, "internal/api/handler.go")
		}
	}
	assert.True(t, found, "should detect stale path reference")
}

func TestValidate_DriftScore(t *testing.T) {
	t.Parallel()
	docsDir := t.TempDir()
	projectDir := t.TempDir()

	// Only create some docs
	writeFile(t, docsDir, "index.md", "# Index\n")
	writeFile(t, docsDir, ".meta.yaml", "generated_at: 2026-01-01T00:00:00Z\n")

	report, err := Validate(docsDir, projectDir)
	require.NoError(t, err)

	assert.Greater(t, report.DriftScore, 0.0)
	assert.LessOrEqual(t, report.DriftScore, 1.0)
}

func TestValidateCommands_StaleBuildFile(t *testing.T) {
	t.Parallel()
	docsDir := t.TempDir()
	projectDir := t.TempDir()

	// backtickPathRe requires at least two path segments separated by /
	// so use a path-like reference that will match
	writeFile(t, docsDir, "commands.md", "# Commands\n\nSee `path/to/Makefile` for build targets.\n")

	warnings := ValidateCommands(docsDir, projectDir)
	// path/to/Makefile doesn't exist in projectDir
	assert.NotEmpty(t, warnings)
	assert.Equal(t, "stale_command", warnings[0].Type)
}
