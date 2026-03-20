package setup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan_PythonProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "pyproject.toml", `[project.scripts]
dev = "uvicorn main:app"
test = "pytest"
`)
	writeFile(t, dir, "main.py", "print('hello')\n")
	writeFile(t, dir, "tests/__init__.py", "")
	writeFile(t, dir, "tests/test_main.py", "def test_main(): pass\n")

	info, err := Scan(dir)
	require.NoError(t, err)

	require.Len(t, info.Languages, 1)
	assert.Equal(t, "Python", info.Languages[0].Name)

	// Build files
	var hasPyproject bool
	for _, bf := range info.BuildFiles {
		if bf.Type == "pyproject.toml" {
			hasPyproject = true
			assert.Contains(t, bf.Commands, "dev")
			assert.Contains(t, bf.Commands, "test")
		}
	}
	assert.True(t, hasPyproject)

	// Test config
	assert.Equal(t, "pytest", info.TestConfig.Framework)
	assert.NotEmpty(t, info.TestConfig.Dirs)

	// Entry points
	assert.NotEmpty(t, info.EntryPoints)
}

func TestScan_RustProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "Cargo.toml", "[package]\nname = \"test\"\nversion = \"0.1.0\"\n")
	writeFile(t, dir, "src/main.rs", "fn main() {}\n")

	info, err := Scan(dir)
	require.NoError(t, err)

	require.Len(t, info.Languages, 1)
	assert.Equal(t, "Rust", info.Languages[0].Name)

	var hasCargo bool
	for _, bf := range info.BuildFiles {
		if bf.Type == "cargo.toml" {
			hasCargo = true
			assert.Contains(t, bf.Commands, "build")
			assert.Contains(t, bf.Commands, "test")
		}
	}
	assert.True(t, hasCargo)
}

func TestScan_WithDockerCompose(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "docker-compose.yml", "version: '3'\nservices:\n  app:\n    build: .\n")

	info, err := Scan(dir)
	require.NoError(t, err)

	var hasDocker bool
	for _, bf := range info.BuildFiles {
		if bf.Type == "docker-compose" {
			hasDocker = true
			assert.Contains(t, bf.Commands, "up")
			assert.Contains(t, bf.Commands, "down")
		}
	}
	assert.True(t, hasDocker)
}

func TestScan_TSProjectWithJest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "package.json", `{
		"devDependencies": {"typescript": "5.0.0"},
		"scripts": {"test": "jest --coverage", "build": "tsc"}
	}`)
	writeFile(t, dir, "src/index.ts", "export default {}\n")
	writeFile(t, dir, "tests/app.test.ts", "test('works', () => {})\n")

	info, err := Scan(dir)
	require.NoError(t, err)

	require.Len(t, info.Languages, 1)
	assert.Equal(t, "TypeScript", info.Languages[0].Name)
	assert.Equal(t, "5.0.0", info.Languages[0].Version)

	assert.Equal(t, "Jest", info.TestConfig.Framework)
}

func TestRender_ConventionsMultiLanguage(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name: "multi",
		Languages: []Language{
			{Name: "Go", Version: "1.23"},
			{Name: "TypeScript", Version: "5.0"},
			{Name: "Python"},
			{Name: "Rust"},
		},
	}

	ds := Render(info, nil)
	assert.Contains(t, ds.Conventions, "Go")
	assert.Contains(t, ds.Conventions, "TypeScript")
	assert.Contains(t, ds.Conventions, "Python")
	assert.Contains(t, ds.Conventions, "Rust")
}

func TestScan_IgnoredDirs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "src/main.go", "package main\n")
	writeFile(t, dir, "node_modules/pkg/index.js", "module.exports = {}\n")
	writeFile(t, dir, ".git/config", "")
	writeFile(t, dir, "vendor/dep/dep.go", "package dep\n")

	info, err := Scan(dir)
	require.NoError(t, err)

	assert.False(t, hasDirNamed(info.Structure, "node_modules"))
	assert.False(t, hasDirNamed(info.Structure, ".git"))
	assert.False(t, hasDirNamed(info.Structure, "vendor"))
}

func TestResolveDocsDir(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/project/.autopus/docs", resolveDocsDir("/project", ""))
	assert.Equal(t, "/custom/dir", resolveDocsDir("/project", "/custom/dir"))
	assert.Equal(t, "/project/my-docs", resolveDocsDir("/project", "my-docs"))
}
