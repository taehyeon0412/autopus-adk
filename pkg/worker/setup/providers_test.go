package setup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectProviders_ReturnsKnownNames(t *testing.T) {
	t.Parallel()

	results := DetectProviders()

	// Should return all known provider binaries
	knownNames := map[string]bool{
		"claude":   true,
		"codex":    true,
		"gemini":   true,
		"opencode": true,
	}

	assert.Len(t, results, len(knownNames))
	for _, ps := range results {
		assert.True(t, knownNames[ps.Name], "unexpected provider: %s", ps.Name)
		assert.Equal(t, ps.Name, ps.Binary)
	}
}

func TestDetectProviders_OrderPreserved(t *testing.T) {
	t.Parallel()

	results := DetectProviders()
	expected := []string{"claude", "codex", "gemini", "opencode"}

	names := make([]string, len(results))
	for i, ps := range results {
		names[i] = ps.Name
	}
	assert.Equal(t, expected, names)
}

func TestDetectProviders_InstalledFieldsSet(t *testing.T) {
	t.Parallel()

	results := DetectProviders()
	for _, ps := range results {
		if ps.Installed {
			// If installed, version should be non-empty
			assert.NotEmpty(t, ps.Version, "installed provider %s should have version", ps.Name)
		} else {
			// Not installed should have empty version
			assert.Empty(t, ps.Version, "uninstalled provider %s should have empty version", ps.Name)
		}
	}
}

func TestCheckNodeJS(t *testing.T) {
	t.Parallel()

	// Just ensure it doesn't panic; actual result depends on environment
	_ = CheckNodeJS()
}

func TestProviderPackages_AllBinariesMapped(t *testing.T) {
	t.Parallel()

	for _, bin := range providerBinaries {
		_, ok := providerPackages[bin]
		assert.True(t, ok, "provider %s should have an npm package mapping", bin)
	}
}
