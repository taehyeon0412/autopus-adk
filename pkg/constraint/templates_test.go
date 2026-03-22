package constraint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/insajin/autopus-adk/pkg/constraint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectLanguage_Go verifies go.mod marker maps to LangGo.
func TestDetectLanguage_Go(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example"), 0o644))

	lang := constraint.DetectLanguage(dir)
	assert.Equal(t, constraint.LangGo, lang)
}

// TestDetectLanguage_TypeScript verifies package.json marker maps to LangTypeScript.
func TestDetectLanguage_TypeScript(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644))

	lang := constraint.DetectLanguage(dir)
	assert.Equal(t, constraint.LangTypeScript, lang)
}

// TestDetectLanguage_Python_PyProject verifies pyproject.toml maps to LangPython.
func TestDetectLanguage_Python_PyProject(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[tool.poetry]"), 0o644))

	lang := constraint.DetectLanguage(dir)
	assert.Equal(t, constraint.LangPython, lang)
}

// TestDetectLanguage_Python_Requirements verifies requirements.txt maps to LangPython.
func TestDetectLanguage_Python_Requirements(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("requests==2.28.0"), 0o644))

	lang := constraint.DetectLanguage(dir)
	assert.Equal(t, constraint.LangPython, lang)
}

// TestDetectLanguage_Default verifies unknown projects fall back to LangGo.
func TestDetectLanguage_Default(t *testing.T) {
	dir := t.TempDir()

	lang := constraint.DetectLanguage(dir)
	assert.Equal(t, constraint.LangGo, lang)
}

// TestDetectLanguage_GoPrecedence verifies go.mod takes priority over package.json.
func TestDetectLanguage_GoPrecedence(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644))

	lang := constraint.DetectLanguage(dir)
	assert.Equal(t, constraint.LangGo, lang)
}

// TestDefaultConstraints_Go verifies Go constraints are non-empty and include expected patterns.
func TestDefaultConstraints_Go(t *testing.T) {
	cs := constraint.DefaultConstraints(constraint.LangGo)
	assert.NotEmpty(t, cs)

	patterns := make(map[string]bool)
	for _, c := range cs {
		patterns[c.Pattern] = true
	}
	assert.True(t, patterns["context.Background()"])
	assert.True(t, patterns["os.Exit("])
}

// TestDefaultConstraints_TypeScript verifies TypeScript constraints include security patterns.
func TestDefaultConstraints_TypeScript(t *testing.T) {
	cs := constraint.DefaultConstraints(constraint.LangTypeScript)
	assert.NotEmpty(t, cs)

	var hasEval, hasXSS bool
	for _, c := range cs {
		if c.Pattern == "eval(" {
			hasEval = true
			assert.Equal(t, constraint.CategorySecurity, c.Category)
		}
		if c.Pattern == "innerHTML" {
			hasXSS = true
			assert.Equal(t, constraint.CategorySecurity, c.Category)
		}
	}
	assert.True(t, hasEval, "should contain eval( pattern")
	assert.True(t, hasXSS, "should contain innerHTML pattern")
}

// TestDefaultConstraints_Python verifies Python constraints include security patterns.
func TestDefaultConstraints_Python(t *testing.T) {
	cs := constraint.DefaultConstraints(constraint.LangPython)
	assert.NotEmpty(t, cs)

	var hasPickle bool
	for _, c := range cs {
		if c.Pattern == "pickle.load" {
			hasPickle = true
			assert.Equal(t, constraint.CategorySecurity, c.Category)
		}
	}
	assert.True(t, hasPickle, "should contain pickle.load pattern")
}

// TestDefaultConstraints_UnknownFallsBackToGo verifies unknown language returns Go defaults.
func TestDefaultConstraints_UnknownFallsBackToGo(t *testing.T) {
	unknown := constraint.DefaultConstraints("unknown")
	goDefaults := constraint.DefaultConstraints(constraint.LangGo)
	assert.Equal(t, goDefaults, unknown)
}

// TestGenerateDefaultFile_CreatesFile verifies that the file is created with content.
func TestGenerateDefaultFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example"), 0o644))

	path, err := constraint.GenerateDefaultFile(dir)
	require.NoError(t, err)
	assert.FileExists(t, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "deny:")
}

// TestGenerateDefaultFile_DoesNotOverwrite verifies existing files are not modified.
func TestGenerateDefaultFile_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	constraintsDir := filepath.Join(dir, ".autopus", "context")
	require.NoError(t, os.MkdirAll(constraintsDir, 0o755))

	existingContent := "deny: []\n"
	existingPath := filepath.Join(constraintsDir, "constraints.yaml")
	require.NoError(t, os.WriteFile(existingPath, []byte(existingContent), 0o644))

	path, err := constraint.GenerateDefaultFile(dir)
	require.NoError(t, err)
	assert.Equal(t, existingPath, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, existingContent, string(data))
}

// TestGenerateDefaultFile_CorrectPath verifies the returned path matches DefaultPath.
func TestGenerateDefaultFile_CorrectPath(t *testing.T) {
	dir := t.TempDir()

	path, err := constraint.GenerateDefaultFile(dir)
	require.NoError(t, err)

	expected := filepath.Join(dir, ".autopus", "context", "constraints.yaml")
	assert.Equal(t, expected, path)
}

// TestGenerateDefaultFile_TypeScriptProject verifies TypeScript patterns in generated file.
func TestGenerateDefaultFile_TypeScriptProject(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644))

	path, err := constraint.GenerateDefaultFile(dir)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "innerHTML")
}

// TestGenerateDefaultFile_PythonProject verifies Python patterns in generated file.
func TestGenerateDefaultFile_PythonProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("requests"), 0o644))

	path, err := constraint.GenerateDefaultFile(dir)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "pickle.load")
}

// TestGenerateDefaultFile_MkdirAllError verifies error when parent path is a file (not directory).
func TestGenerateDefaultFile_MkdirAllError(t *testing.T) {
	t.Parallel()
	if os.Getuid() == 0 {
		t.Skip("running as root; permission check not enforced")
	}

	// Make .autopus a file so MkdirAll fails trying to create .autopus/context.
	dir := t.TempDir()
	blockPath := filepath.Join(dir, ".autopus")
	require.NoError(t, os.WriteFile(blockPath, []byte("not-a-dir"), 0o644))

	_, err := constraint.GenerateDefaultFile(dir)
	assert.Error(t, err)
}

// TestGenerateDefaultFile_WriteError verifies error when output directory is read-only.
func TestGenerateDefaultFile_WriteError(t *testing.T) {
	t.Parallel()
	if os.Getuid() == 0 {
		t.Skip("running as root; permission check not enforced")
	}

	dir := t.TempDir()
	constraintsDir := filepath.Join(dir, ".autopus", "context")
	require.NoError(t, os.MkdirAll(constraintsDir, 0o755))
	// Remove write permission on the parent directory so WriteFile fails.
	require.NoError(t, os.Chmod(constraintsDir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(constraintsDir, 0o755) })

	_, err := constraint.GenerateDefaultFile(dir)
	assert.Error(t, err)
}
