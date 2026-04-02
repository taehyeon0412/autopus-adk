package knowledge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExcluder_BuiltInDirectoryExclusions(t *testing.T) {
	t.Parallel()

	e, err := NewExcluder("/nonexistent/.gitignore")
	require.NoError(t, err, "missing .gitignore should not error")

	tests := []struct {
		name     string
		path     string
		excluded bool
	}{
		{"git directory", ".git/config", true},
		{"nested git", "sub/.git/HEAD", true},
		{"node_modules", "node_modules/pkg/index.js", true},
		{"nested node_modules", "app/node_modules/x.js", true},
		{"normal file", "src/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.excluded, e.IsExcluded(tt.path))
		})
	}
}

func TestExcluder_BuiltInExtensionExclusions(t *testing.T) {
	t.Parallel()

	e, err := NewExcluder("/nonexistent/.gitignore")
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		excluded bool
	}{
		{"exe file", "build/app.exe", true},
		{"bin file", "out/tool.bin", true},
		{"EXE uppercase", "build/app.EXE", true},
		{"go file", "main.go", false},
		{"js file", "app.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.excluded, e.IsExcluded(tt.path))
		})
	}
}

func TestExcluder_GlobPatterns(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gitignore := filepath.Join(dir, ".gitignore")
	content := "*.log\n*.tmp\nbuild/\n"
	require.NoError(t, os.WriteFile(gitignore, []byte(content), 0644))

	e, err := NewExcluder(gitignore)
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		excluded bool
	}{
		{"log file matched", "app.log", true},
		{"tmp file matched", "cache.tmp", true},
		{"go file not matched", "main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.excluded, e.IsExcluded(tt.path))
		})
	}
}

func TestExcluder_NegationPattern(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gitignore := filepath.Join(dir, ".gitignore")
	content := "*.log\n!important.log\n"
	require.NoError(t, os.WriteFile(gitignore, []byte(content), 0644))

	e, err := NewExcluder(gitignore)
	require.NoError(t, err)

	assert.True(t, e.IsExcluded("debug.log"), "debug.log should be excluded")
	assert.False(t, e.IsExcluded("important.log"), "important.log should NOT be excluded (negation)")
}

func TestExcluder_CommentsAndBlankLines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gitignore := filepath.Join(dir, ".gitignore")
	content := "# This is a comment\n\n*.log\n  # Another comment\n"
	require.NoError(t, os.WriteFile(gitignore, []byte(content), 0644))

	e, err := NewExcluder(gitignore)
	require.NoError(t, err)

	assert.Len(t, e.patterns, 1, "comments and blank lines should be skipped")
	assert.True(t, e.IsExcluded("app.log"))
}

func TestExcluder_DirectoryPattern(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gitignore := filepath.Join(dir, ".gitignore")
	content := "dist/\n"
	require.NoError(t, os.WriteFile(gitignore, []byte(content), 0644))

	e, err := NewExcluder(gitignore)
	require.NoError(t, err)

	assert.Len(t, e.patterns, 1)
	assert.True(t, e.patterns[0].isDir)
	assert.Equal(t, "dist", e.patterns[0].glob)
}

func TestExcluder_MissingGitignore(t *testing.T) {
	t.Parallel()

	e, err := NewExcluder("/does/not/exist/.gitignore")
	require.NoError(t, err, "missing file should return empty excluder, no error")
	assert.NotNil(t, e)
	// Only built-in exclusions should apply.
	assert.False(t, e.IsExcluded("readme.md"))
}
