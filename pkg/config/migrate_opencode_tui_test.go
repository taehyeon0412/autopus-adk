package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SPEC-ORCH-014 R5: MigrateOrchestraConfig must migrate opencode from args mode
// to TUI mode: remove "run" from PaneArgs, clear InteractiveInput.

// TestMigrateOrchestraConfig_OpencodeArgsToTUI verifies that existing opencode
// config with interactive_input="args" and PaneArgs containing "run" is migrated
// to empty interactive_input and PaneArgs without "run".
// RED: MigrateOrchestraConfig does not perform this migration yet.
func TestMigrateOrchestraConfig_OpencodeArgsToTUI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		inputPaneArgs     []string
		inputInteractive  string
		expectPaneArgs    []string
		expectInteractive string
		expectChanged     bool
	}{
		{
			name:              "args mode with run migrated to TUI",
			inputPaneArgs:     []string{"run", "-m", "openai/gpt-5.4"},
			inputInteractive:  "args",
			expectPaneArgs:    []string{"-m", "openai/gpt-5.4"},
			expectInteractive: "",
			expectChanged:     true,
		},
		{
			name:              "already TUI mode unchanged",
			inputPaneArgs:     []string{"-m", "openai/gpt-5.4"},
			inputInteractive:  "",
			expectPaneArgs:    []string{"-m", "openai/gpt-5.4"},
			expectInteractive: "",
			expectChanged:     false,
		},
		{
			name:              "args mode without run in PaneArgs",
			inputPaneArgs:     []string{"-m", "openai/gpt-5.4"},
			inputInteractive:  "args",
			expectPaneArgs:    []string{"-m", "openai/gpt-5.4"},
			expectInteractive: "",
			expectChanged:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &HarnessConfig{
				Mode:        ModeFull,
				ProjectName: "test-project",
				Platforms:   []string{"claude-code", "opencode"},
				Orchestra: OrchestraConf{
					Enabled: true,
					Providers: map[string]ProviderEntry{
						"claude": {Binary: "claude", Args: []string{"--print"}, PaneArgs: []string{"--print"}},
						"opencode": {
							Binary:           "opencode",
							Args:             []string{"run", "-m", "openai/gpt-5.4"},
							PaneArgs:         tt.inputPaneArgs,
							PromptViaArgs:    true,
							InteractiveInput: tt.inputInteractive,
						},
					},
					Commands: map[string]CommandEntry{
						"review": {Strategy: "debate", Providers: []string{"claude", "opencode"}},
					},
				},
			}

			changed, err := MigrateOrchestraConfig(cfg)
			require.NoError(t, err)

			opencode := cfg.Orchestra.Providers["opencode"]
			assert.Equal(t, tt.expectPaneArgs, opencode.PaneArgs,
				"opencode PaneArgs must match expected after migration (R5)")
			assert.Equal(t, tt.expectInteractive, opencode.InteractiveInput,
				"opencode InteractiveInput must be empty after migration (R5)")

			if tt.expectChanged {
				assert.True(t, changed,
					"changed must be true when opencode config was migrated (R5)")
			}
		})
	}
}

// TestMigrateOrchestraConfig_OpencodeRunRemovedFromPaneArgs verifies that
// "run" subcmd is specifically removed from opencode PaneArgs during migration.
// RED: current migration does not touch PaneArgs "run".
func TestMigrateOrchestraConfig_OpencodeRunRemovedFromPaneArgs(t *testing.T) {
	t.Parallel()

	cfg := &HarnessConfig{
		Mode:        ModeFull,
		ProjectName: "test-project",
		Platforms:   []string{"opencode"},
		Orchestra: OrchestraConf{
			Enabled: true,
			Providers: map[string]ProviderEntry{
				"opencode": {
					Binary:           "opencode",
					Args:             []string{"run", "-m", "openai/gpt-5.4"},
					PaneArgs:         []string{"run", "-m", "openai/gpt-5.4"},
					PromptViaArgs:    true,
					InteractiveInput: "args",
				},
			},
			Commands: map[string]CommandEntry{},
		},
	}

	_, err := MigrateOrchestraConfig(cfg)
	require.NoError(t, err)

	opencode := cfg.Orchestra.Providers["opencode"]
	assert.NotContains(t, opencode.PaneArgs, "run",
		"PaneArgs must not contain 'run' after migration (R5)")
}

// TestDefaultProviderEntries_OpencodePostMigration verifies that
// defaultProviderEntries used by migration has correct TUI-mode defaults.
// RED: current defaultProviderEntries opencode has "run" in PaneArgs.
func TestDefaultProviderEntries_OpencodePostMigration(t *testing.T) {
	t.Parallel()

	entry, ok := defaultProviderEntries["opencode"]
	require.True(t, ok, "opencode must exist in defaultProviderEntries")

	// After SPEC-ORCH-014, the canonical default should not have "run" in PaneArgs.
	assert.NotContains(t, entry.PaneArgs, "run",
		"defaultProviderEntries opencode PaneArgs must not contain 'run' (R5)")
	assert.Empty(t, entry.InteractiveInput,
		"defaultProviderEntries opencode InteractiveInput must be empty (R5)")
}
