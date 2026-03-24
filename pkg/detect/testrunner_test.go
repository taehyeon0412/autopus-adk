package detect_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/detect"
)

// TestDetectTestRunner_Jest verifies that Jest is detected when package.json
// lists jest as a devDependency or dependency.
func TestDetectTestRunner_Jest(t *testing.T) {
	t.Parallel()

	// Given: a project directory with package.json referencing jest
	dir := t.TempDir()
	packageJSON := `{
  "name": "my-app",
  "devDependencies": {
    "jest": "^29.0.0"
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: jest is detected
	assert.Equal(t, "jest", runner)
}

// TestDetectTestRunner_Vitest verifies that vitest is detected when a
// vitest.config file exists in the project directory.
func TestDetectTestRunner_Vitest(t *testing.T) {
	t.Parallel()

	// Given: a project directory with vitest.config.ts
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "vitest.config.ts"), []byte(`export default {}`), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: vitest is detected
	assert.Equal(t, "vitest", runner)
}

// TestDetectTestRunner_GoTest verifies that go test is detected when a go.mod
// file exists in the project directory.
func TestDetectTestRunner_GoTest(t *testing.T) {
	t.Parallel()

	// Given: a project directory with go.mod
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: go test is detected
	assert.Equal(t, "go test", runner)
}

// TestDetectTestRunner_Unknown verifies that an empty string is returned when
// no known test runner configuration is found.
func TestDetectTestRunner_Unknown(t *testing.T) {
	t.Parallel()

	// Given: an empty project directory with no recognizable test config
	dir := t.TempDir()

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: an empty string (unknown) is returned
	assert.Equal(t, "", runner)
}

// TestDetectTestRunner_VitestJS verifies that vitest.config.js triggers vitest
// detection (not only .ts variant).
func TestDetectTestRunner_VitestJS(t *testing.T) {
	t.Parallel()

	// Given: a project directory with vitest.config.js
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "vitest.config.js"), []byte(`export default {}`), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: vitest is detected
	assert.Equal(t, "vitest", runner)
}

// TestDetectTestRunner_JestDependency verifies that jest in the dependencies
// (not devDependencies) section of package.json is also detected.
func TestDetectTestRunner_JestDependency(t *testing.T) {
	t.Parallel()

	// Given: package.json with jest in dependencies (not devDependencies)
	dir := t.TempDir()
	packageJSON := `{"dependencies": {"jest": "^29.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: jest is detected
	assert.Equal(t, "jest", runner)
}

// TestDetectTestRunner_PytestIni verifies that a pytest.ini file triggers
// pytest detection.
func TestDetectTestRunner_PytestIni(t *testing.T) {
	t.Parallel()

	// Given: a project directory with pytest.ini
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pytest.ini"), []byte("[pytest]\n"), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: pytest is detected
	assert.Equal(t, "pytest", runner)
}

// TestDetectTestRunner_PyprojectTomlWithPytest verifies that a pyproject.toml
// containing "pytest" triggers pytest detection.
func TestDetectTestRunner_PyprojectTomlWithPytest(t *testing.T) {
	t.Parallel()

	// Given: a pyproject.toml with pytest configuration
	dir := t.TempDir()
	content := "[tool.pytest.ini_options]\ntestpaths = [\"tests\"]\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(content), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: pytest is detected
	assert.Equal(t, "pytest", runner)
}

// TestDetectTestRunner_CargoToml verifies that a Cargo.toml triggers cargo test
// detection.
func TestDetectTestRunner_CargoToml(t *testing.T) {
	t.Parallel()

	// Given: a project directory with Cargo.toml
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"mylib\"\n"), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: cargo test is detected
	assert.Equal(t, "cargo test", runner)
}

// TestDetectTestRunner_PyprojectToml_NoPytest verifies that a pyproject.toml
// without pytest content does NOT trigger pytest detection.
func TestDetectTestRunner_PyprojectToml_NoPytest(t *testing.T) {
	t.Parallel()

	// Given: a pyproject.toml with no pytest reference
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[tool.poetry]\nname = \"myapp\"\n"), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: no runner is detected (pyproject.toml without pytest is not enough)
	assert.Equal(t, "", runner)
}

// TestDetectTestRunner_MalformedPackageJSON_NoError verifies that a malformed
// package.json does not cause an error, just skips jest detection.
func TestDetectTestRunner_MalformedPackageJSON_NoError(t *testing.T) {
	t.Parallel()

	// Given: a malformed package.json
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{invalid json`), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: no jest runner and no error (silently skipped)
	assert.Equal(t, "", runner)
}

// TestDetectTestRunner_PyprojectToml_EmptyFile verifies that an empty
// pyproject.toml (no pytest content) does not trigger pytest detection.
func TestDetectTestRunner_PyprojectToml_EmptyFile(t *testing.T) {
	t.Parallel()

	// Given: an empty pyproject.toml
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(""), 0o644))

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: pytest is NOT detected
	assert.Equal(t, "", runner)
}

// TestDetectTestRunner_PackageJSON_NoReadPermission verifies that when
// package.json exists but cannot be read, an error is propagated.
func TestDetectTestRunner_PackageJSON_NoReadPermission(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: permission tests are not meaningful")
	}
	t.Parallel()

	// Given: a package.json with no read permission (not a "not found" error)
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"devDependencies":{"jest":"^29"}}`), 0o644))
	require.NoError(t, os.Chmod(path, 0o000))
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	// When: DetectTestRunner is called
	_, err := detect.DetectTestRunner(dir)

	// Then: an error is returned (permission denied propagates)
	require.Error(t, err)
}

// TestDetectTestRunner_PyprojectToml_NoReadPermission verifies that when
// pyproject.toml exists but cannot be read, pytest is not detected and no
// error is returned.
func TestDetectTestRunner_PyprojectToml_NoReadPermission(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: permission tests are not meaningful")
	}
	t.Parallel()

	// Given: a pyproject.toml with no read permission
	dir := t.TempDir()
	path := filepath.Join(dir, "pyproject.toml")
	require.NoError(t, os.WriteFile(path, []byte("[tool.pytest]\n"), 0o644))
	require.NoError(t, os.Chmod(path, 0o000))
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	// When: DetectTestRunner is called
	runner, err := detect.DetectTestRunner(dir)
	require.NoError(t, err)

	// Then: pytest is not detected (read failed, silently ignored)
	assert.Equal(t, "", runner)
}
