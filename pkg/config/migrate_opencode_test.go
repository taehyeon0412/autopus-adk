package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrateOrchestraConfig_CodexToOpencode verifies R9: codex provider
// entries are migrated to opencode (binary, args, platform mapping).
func TestMigrateOrchestraConfig_CodexToOpencode(t *testing.T) {
	t.Parallel()

	cfg := &HarnessConfig{
		Mode:        ModeFull,
		ProjectName: "test-project",
		Platforms:   []string{"claude-code", "codex"},
		Orchestra: OrchestraConf{
			Enabled: true,
			Providers: map[string]ProviderEntry{
				"claude": {Binary: "claude", Args: []string{"--print"}},
				"codex":  {Binary: "codex", Args: []string{"--quiet"}, PromptViaArgs: true},
			},
			Commands: map[string]CommandEntry{
				"review": {Strategy: "debate", Providers: []string{"claude", "codex"}},
			},
		},
	}

	changed, err := MigrateCodexToOpencode(cfg)
	require.NoError(t, err)
	assert.True(t, changed, "changed must be true when codex is migrated to opencode")

	// codex entry must be removed.
	_, hasCodex := cfg.Orchestra.Providers["codex"]
	assert.False(t, hasCodex, "codex provider must be removed after migration")

	// opencode entry must exist with correct settings.
	opencode, hasOpencode := cfg.Orchestra.Providers["opencode"]
	require.True(t, hasOpencode, "opencode provider must exist after migration")
	assert.Equal(t, "opencode", opencode.Binary, "opencode binary must be 'opencode'")

	// Commands must reference opencode instead of codex.
	review := cfg.Orchestra.Commands["review"]
	assert.Contains(t, review.Providers, "opencode",
		"command providers must include opencode after migration")
	assert.NotContains(t, review.Providers, "codex",
		"command providers must not include codex after migration")
}

// TestMigrateOrchestraConfig_OpencodeAlreadyExists verifies that when
// opencode provider already exists, codex-to-opencode migration does
// not create a duplicate entry.
func TestMigrateOrchestraConfig_OpencodeAlreadyExists(t *testing.T) {
	t.Parallel()

	cfg := &HarnessConfig{
		Mode:        ModeFull,
		ProjectName: "test-project",
		Platforms:   []string{"claude-code", "codex"},
		Orchestra: OrchestraConf{
			Enabled: true,
			Providers: map[string]ProviderEntry{
				"claude":   {Binary: "claude", Args: []string{"--print"}},
				"codex":    {Binary: "codex", Args: []string{"--quiet"}, PromptViaArgs: true},
				"opencode": {Binary: "opencode", Args: []string{}},
			},
			Commands: map[string]CommandEntry{
				"plan": {Strategy: "consensus", Providers: []string{"claude", "codex", "opencode"}},
			},
		},
	}

	changed, err := MigrateCodexToOpencode(cfg)
	require.NoError(t, err)

	// codex must still be removed.
	_, hasCodex := cfg.Orchestra.Providers["codex"]
	assert.False(t, hasCodex, "codex must be removed even when opencode already exists")

	// opencode must not be duplicated — still exactly one entry.
	_, hasOpencode := cfg.Orchestra.Providers["opencode"]
	assert.True(t, hasOpencode, "opencode must still exist")

	// Commands must not list codex.
	plan := cfg.Orchestra.Commands["plan"]
	assert.NotContains(t, plan.Providers, "codex",
		"command providers must not include codex after migration")

	_ = changed
}

// TestPlatformToProvider_Opencode verifies that the "opencode" platform
// correctly maps to the "opencode" orchestra provider.
func TestPlatformToProvider_Opencode(t *testing.T) {
	t.Parallel()

	got := PlatformToProvider("opencode")
	assert.Equal(t, "opencode", got,
		"platform 'opencode' must map to provider 'opencode'")
}

// TestMigrateCodexToOpencode_OrchestraDisabled verifies no-op when
// orchestra is not enabled.
func TestMigrateCodexToOpencode_OrchestraDisabled(t *testing.T) {
	t.Parallel()

	cfg := &HarnessConfig{
		Orchestra: OrchestraConf{
			Enabled: false,
			Providers: map[string]ProviderEntry{
				"codex": {Binary: "codex"},
			},
		},
	}

	changed, err := MigrateCodexToOpencode(cfg)
	require.NoError(t, err)
	assert.False(t, changed, "must not change when orchestra disabled")
}

// TestMigrateCodexToOpencode_NilProviders verifies no-op when providers
// map is nil.
func TestMigrateCodexToOpencode_NilProviders(t *testing.T) {
	t.Parallel()

	cfg := &HarnessConfig{
		Orchestra: OrchestraConf{
			Enabled:   true,
			Providers: nil,
		},
	}

	changed, err := MigrateCodexToOpencode(cfg)
	require.NoError(t, err)
	assert.False(t, changed, "must not change when providers is nil")
}

// TestMigrateCodexToOpencode_NoCodex verifies no-op when codex is absent.
func TestMigrateCodexToOpencode_NoCodex(t *testing.T) {
	t.Parallel()

	cfg := &HarnessConfig{
		Orchestra: OrchestraConf{
			Enabled: true,
			Providers: map[string]ProviderEntry{
				"claude": {Binary: "claude"},
			},
		},
	}

	changed, err := MigrateCodexToOpencode(cfg)
	require.NoError(t, err)
	assert.False(t, changed, "must not change when codex not present")
}
