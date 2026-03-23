// Package cli tests additional internal orchestra helper functions.
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// TestBuildReviewPrompt_MissingFile verifies that a missing file is handled
// gracefully without panicking and includes an error indication.
func TestBuildReviewPrompt_MissingFile(t *testing.T) {
	t.Parallel()

	prompt := buildReviewPrompt([]string{"/nonexistent/file.go"})
	assert.Contains(t, prompt, "읽기 실패", "prompt must indicate read failure for missing file")
}

// TestBuildFileContents_EmptySlice verifies buildFileContents returns empty
// string for a nil/empty file list.
func TestBuildFileContents_EmptySlice(t *testing.T) {
	t.Parallel()

	result := buildFileContents(nil)
	assert.Empty(t, result)

	result2 := buildFileContents([]string{})
	assert.Empty(t, result2)
}

// TestBuildFileContents_ExistingAndMissing verifies mixed existing/missing files
// produce both content and error entries.
func TestBuildFileContents_ExistingAndMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	existing := filepath.Join(dir, "real.go")
	require.NoError(t, os.WriteFile(existing, []byte("package real\n"), 0o644))

	result := buildFileContents([]string{existing, "/nonexistent/missing.go"})
	assert.Contains(t, result, "real.go")
	assert.Contains(t, result, "package real")
	assert.Contains(t, result, "읽기 실패")
}

// TestResolveStrategy_EmptyCommandStrategy verifies the global default is used
// when the command exists but has an empty strategy string.
func TestResolveStrategy_EmptyCommandStrategy(t *testing.T) {
	t.Parallel()

	conf := &config.OrchestraConf{
		DefaultStrategy: "pipeline",
		Commands: map[string]config.CommandEntry{
			"review": {Strategy: ""}, // empty strategy in command
		},
	}

	result := resolveStrategy(conf, "review", "")
	assert.Equal(t, "pipeline", result,
		"global default must be used when command strategy is empty")
}

// TestResolveProviderNames_EmptyCommandProviders verifies the global provider
// fallback is used when the command's provider list is empty.
func TestResolveProviderNames_EmptyCommandProviders(t *testing.T) {
	t.Parallel()

	conf := &config.OrchestraConf{
		Providers: map[string]config.ProviderEntry{
			"claude": {Binary: "claude"},
			"gemini": {Binary: "gemini"},
		},
		Commands: map[string]config.CommandEntry{
			"plan": {Providers: nil}, // empty providers list
		},
	}

	names := resolveProviderNames(conf, "plan", nil)
	// Should fall back to all global providers.
	assert.Len(t, names, 2)
}

// TestBuildProviderConfigs_MixedKnownUnknown verifies mixed known/unknown
// providers are handled correctly in a single call.
func TestBuildProviderConfigs_MixedKnownUnknown(t *testing.T) {
	t.Parallel()

	configs := buildProviderConfigs([]string{"claude", "my-tool"})
	require.Len(t, configs, 2)

	assert.Equal(t, "claude", configs[0].Binary, "claude must use hardcoded binary")
	assert.Equal(t, "my-tool", configs[1].Binary, "unknown provider must use name as binary")
}
