// Package e2e provides user-facing scenario-based E2E test infrastructure.
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsSecret_KeywordMatching verifies that isSecret correctly identifies
// secret-like env var names based on known keywords.
func TestIsSecret_KeywordMatching(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key      string
		expected bool
	}{
		{"API_KEY", true},
		{"MY_API_KEY", true},
		{"SECRET", true},
		{"DB_SECRET", true},
		{"TOKEN", true},
		{"ACCESS_TOKEN", true},
		{"PASSWORD", true},
		{"PASSWD", true},
		{"PRIVATE_KEY", true},
		{"CREDENTIALS", true},
		{"DATABASE_URL", false},
		{"PORT", false},
		{"HOST", false},
		{"LOG_LEVEL", false},
		{"MY_TOKEN_COUNT", true}, // contains "TOKEN"
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, isSecret(tt.key))
		})
	}
}

// TestLoadTestEnvFile_ValidFile verifies that loadTestEnvFile reads KEY=VALUE
// pairs from a properly formatted file.
func TestLoadTestEnvFile_ValidFile(t *testing.T) {
	t.Parallel()

	// Given: a .autopus/test.env file with valid entries
	dir := t.TempDir()
	envPath := filepath.Join(dir, "test.env")
	content := `# Comment line
API_KEY=my-secret-key
DATABASE_URL=postgres://localhost/testdb

PORT=9090
`
	require.NoError(t, os.WriteFile(envPath, []byte(content), 0o644))

	// When: loadTestEnvFile is called
	env := loadTestEnvFile(envPath)

	// Then: three key-value pairs are loaded, comment and blank skipped
	assert.Equal(t, "my-secret-key", env["API_KEY"])
	assert.Equal(t, "postgres://localhost/testdb", env["DATABASE_URL"])
	assert.Equal(t, "9090", env["PORT"])
	assert.NotContains(t, env, "# Comment line")
}

// TestLoadTestEnvFile_EmptyPath verifies that an empty path returns empty map.
func TestLoadTestEnvFile_EmptyPath(t *testing.T) {
	t.Parallel()

	env := loadTestEnvFile("")
	assert.Empty(t, env)
}

// TestLoadTestEnvFile_NonExistentFile verifies that a missing file returns empty map without error.
func TestLoadTestEnvFile_NonExistentFile(t *testing.T) {
	t.Parallel()

	env := loadTestEnvFile("/tmp/does-not-exist-e2e-test.env")
	assert.Empty(t, env)
}

// TestDetectEnvFromProject_EnvExampleFile verifies that .env.example values
// are loaded into the auto-detected environment map.
func TestDetectEnvFromProject_EnvExampleFile(t *testing.T) {
	t.Parallel()

	// Given: a project directory with .env.example
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ".env.example"),
		[]byte("MY_CUSTOM_VAR=example_value\nANOTHER_VAR=another\n"),
		0o644,
	))

	// When: detectEnvFromProject is called
	env := detectEnvFromProject(dir)

	// Then: values from .env.example are present
	assert.Equal(t, "example_value", env["MY_CUSTOM_VAR"])
	assert.Equal(t, "another", env["ANOTHER_VAR"])
}

// TestDetectEnvFromProject_DockerComposeFile verifies that environment keys
// from a docker-compose.yml are picked up.
func TestDetectEnvFromProject_DockerComposeFile(t *testing.T) {
	t.Parallel()

	// Given: a project directory with docker-compose.yml containing env block
	dir := t.TempDir()
	compose := `version: "3"
services:
  app:
    image: myapp
    environment:
      - COMPOSE_DB=postgres://db/test
      - COMPOSE_PORT=5432
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(compose), 0o644))

	// When: detectEnvFromProject is called
	env := detectEnvFromProject(dir)

	// Then: compose environment entries are included
	assert.Equal(t, "postgres://db/test", env["COMPOSE_DB"])
	assert.Equal(t, "5432", env["COMPOSE_PORT"])
}

// TestDetectEnvFromProject_EmptyDir verifies that an empty directory returns
// at least the Go toolchain env vars without error.
func TestDetectEnvFromProject_EmptyDir(t *testing.T) {
	t.Parallel()

	env := detectEnvFromProject(t.TempDir())
	// Should not panic; result may contain Go env vars
	assert.NotNil(t, env)
}

// TestResolveEnv_TestEnvFile_TakesPrecedence verifies that values in test.env
// override safe defaults and auto-detected values.
func TestResolveEnv_TestEnvFile_TakesPrecedence(t *testing.T) {
	t.Parallel()

	// Given: a test.env file with a DATABASE_URL override
	dir := t.TempDir()
	envPath := filepath.Join(dir, "test.env")
	require.NoError(t, os.WriteFile(envPath, []byte("DATABASE_URL=testenv://override\n"), 0o644))

	opts := EnvResolveOptions{
		ProjectDir:     dir,
		ScenarioEnv:    map[string]string{},
		NonInteractive: true,
		TestEnvFile:    envPath,
	}

	// When: ResolveEnv is called
	env, err := ResolveEnv(opts)

	// Then: test.env value takes precedence over safe default
	require.NoError(t, err)
	assert.Equal(t, "testenv://override", env["DATABASE_URL"])
}

// TestResolveEnv_NonSecretMissingVar_DoesNotSkip verifies that a missing
// non-secret required var does NOT trigger ErrSkipScenario.
func TestResolveEnv_NonSecretMissingVar_DoesNotSkip(t *testing.T) {
	t.Parallel()

	// Given: a non-secret required var that is not set (no secret keywords in name)
	opts := EnvResolveOptions{
		ProjectDir:     t.TempDir(),
		ScenarioEnv:    map[string]string{},
		NonInteractive: true,
		RequiredVars:   []string{"MY_CUSTOM_CONFIG_VAR"}, // no secret keywords
	}

	// When: ResolveEnv is called
	env, err := ResolveEnv(opts)

	// Then: no skip error — non-secret vars are not required to have values
	require.NoError(t, err)
	assert.NotNil(t, env)
}

// TestResolveEnv_AutoDiscovery_DefaultTestEnvPath verifies that ResolveEnv
// auto-discovers .autopus/test.env when TestEnvFile is empty.
func TestResolveEnv_AutoDiscovery_DefaultTestEnvPath(t *testing.T) {
	t.Parallel()

	// Given: a .autopus/test.env file at default location
	dir := t.TempDir()
	autopusDir := filepath.Join(dir, ".autopus")
	require.NoError(t, os.MkdirAll(autopusDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(autopusDir, "test.env"),
		[]byte("AUTO_DISCOVERED=from-default-path\n"),
		0o644,
	))

	opts := EnvResolveOptions{
		ProjectDir:     dir,
		ScenarioEnv:    map[string]string{},
		NonInteractive: true,
		// TestEnvFile intentionally empty — should auto-discover
	}

	// When: ResolveEnv is called
	env, err := ResolveEnv(opts)

	// Then: value from auto-discovered test.env is present
	require.NoError(t, err)
	assert.Equal(t, "from-default-path", env["AUTO_DISCOVERED"])
}

// TestGoEnvVars_ReturnsMap verifies that goEnvVars returns a non-nil map.
// It cannot assert specific values since they depend on the environment.
func TestGoEnvVars_ReturnsMap(t *testing.T) {
	t.Parallel()

	// When: goEnvVars is called
	env := goEnvVars()

	// Then: returns a non-nil map (may be empty in constrained environments)
	assert.NotNil(t, env)
}
