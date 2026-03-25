package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectFramework_NuxtJS tests detection via nuxt.config.* config signal.
func TestDetectFramework_NuxtJS(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nuxt.config.ts"), []byte("// nuxt config"), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "nuxtjs", fw.Name)
	assert.Equal(t, "typescript", fw.Stack)
}

// TestDetectFramework_NestJS tests detection via @nestjs/core in package.json dependencies.
func TestDetectFramework_NestJS(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pkg := `{"dependencies":{"@nestjs/core":"^10.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "nestjs", fw.Name)
	assert.Equal(t, "typescript", fw.Stack)
}

// TestDetectFramework_Svelte tests detection via svelte in package.json dependencies.
func TestDetectFramework_Svelte(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pkg := `{"dependencies":{"svelte":"^4.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "svelte", fw.Name)
}

// TestDetectFramework_Vue tests detection via vue in package.json devDependencies.
func TestDetectFramework_Vue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pkg := `{"devDependencies":{"vue":"^3.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "vue", fw.Name)
}

// TestDetectFramework_React tests detection via react in package.json dependencies.
func TestDetectFramework_React(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pkg := `{"dependencies":{"react":"^18.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "react", fw.Name)
	assert.Equal(t, "typescript", fw.Stack)
}

// TestDetectFramework_PackageJSONInvalidJSON tests graceful handling of malformed package.json.
func TestDetectFramework_PackageJSONInvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte("not json {{{"), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	assert.Nil(t, fw, "invalid JSON must not panic and return nil")
}

// TestDetectFramework_PackageJSONNoDeps tests package.json with no known framework deps.
func TestDetectFramework_PackageJSONNoDeps(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pkg := `{"name":"my-app","dependencies":{"lodash":"^4.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	assert.Nil(t, fw)
}

// TestDetectFramework_RustAxum tests detection of axum in Cargo.toml.
func TestDetectFramework_RustAxum(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cargo := "[package]\nname = \"myapp\"\n[dependencies]\naxum = \"0.7\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(cargo), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "axum", fw.Name)
	assert.Equal(t, "rust", fw.Stack)
}

// TestDetectFramework_RustNoFramework tests Cargo.toml with no known framework.
func TestDetectFramework_RustNoFramework(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cargo := "[package]\nname = \"myapp\"\n[dependencies]\nserde = \"1.0\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(cargo), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	assert.Nil(t, fw)
}

// TestDetectFramework_GoGin tests detection of gin framework in go.mod.
func TestDetectFramework_GoGin(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	gomod := "module example.com/test\n\ngo 1.21\n\nrequire github.com/gin-gonic/gin v1.9.1\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644))

	fw, err := DetectFramework(dir)
	require.NoError(t, err)
	require.NotNil(t, fw)
	assert.Equal(t, "gin", fw.Name)
	assert.Equal(t, "go", fw.Stack)
}

// TestDetectStack_RustStack tests Cargo.toml stack detection.
func TestDetectStack_RustStack(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\n"), 0644))

	stack, err := DetectStack(dir)
	require.NoError(t, err)
	assert.Equal(t, "rust", stack)
}

// TestDetectStack_RequirementsTxt tests requirements.txt → python stack.
func TestDetectStack_RequirementsTxt(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("fastapi\n"), 0644))

	stack, err := DetectStack(dir)
	require.NoError(t, err)
	assert.Equal(t, "python", stack)
}
