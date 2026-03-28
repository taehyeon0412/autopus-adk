package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SPEC-ORCH-014 R1: opencode default PaneArgs must NOT contain "run" subcmd.
// Only TUI flags like "-m openai/gpt-5.4" should be present.
// RED: current defaultProviderEntries has PaneArgs: ["run", "-m", "openai/gpt-5.4"]

func TestDefaultProviderEntries_OpencodePaneArgsNoRun(t *testing.T) {
	t.Parallel()

	entry, ok := defaultProviderEntries["opencode"]
	require.True(t, ok, "opencode must exist in defaultProviderEntries")

	for _, arg := range entry.PaneArgs {
		assert.NotEqual(t, "run", arg,
			"opencode PaneArgs must NOT contain 'run' subcmd (R1: TUI mode only)")
	}
}

// SPEC-ORCH-014 R1: opencode default PaneArgs should be exactly ["-m", "openai/gpt-5.4"].
func TestDefaultProviderEntries_OpencodePaneArgsExact(t *testing.T) {
	t.Parallel()

	entry, ok := defaultProviderEntries["opencode"]
	require.True(t, ok, "opencode must exist in defaultProviderEntries")

	assert.Equal(t, []string{"-m", "openai/gpt-5.4"}, entry.PaneArgs,
		"opencode PaneArgs must be TUI flags only, no 'run' subcmd (R1)")
}

// SPEC-ORCH-014 R1: opencode default InteractiveInput must be empty string.
// Empty string means prompt is delivered via sendkeys/SendLongText, not CLI args.
func TestDefaultProviderEntries_OpencodeInteractiveInputEmpty(t *testing.T) {
	t.Parallel()

	entry, ok := defaultProviderEntries["opencode"]
	require.True(t, ok, "opencode must exist in defaultProviderEntries")

	assert.Empty(t, entry.InteractiveInput,
		"opencode InteractiveInput must be empty string (R1: sendkeys mode, not args)")
}

// SPEC-ORCH-014 R1: DefaultFullConfig opencode PaneArgs must NOT contain "run".
func TestDefaultFullConfig_OpencodePaneArgsNoRun(t *testing.T) {
	t.Parallel()

	cfg := DefaultFullConfig("test-project")
	require.NotNil(t, cfg)

	opencode, ok := cfg.Orchestra.Providers["opencode"]
	require.True(t, ok, "opencode provider must exist in DefaultFullConfig")

	for _, arg := range opencode.PaneArgs {
		assert.NotEqual(t, "run", arg,
			"DefaultFullConfig opencode PaneArgs must NOT contain 'run' (R1)")
	}
}

// SPEC-ORCH-014 R1: DefaultFullConfig opencode Args should still contain "run"
// for non-interactive (batch) mode, but PaneArgs should not.
func TestDefaultFullConfig_OpencodeArgsSeparateFromPaneArgs(t *testing.T) {
	t.Parallel()

	cfg := DefaultFullConfig("test-project")
	require.NotNil(t, cfg)

	opencode, ok := cfg.Orchestra.Providers["opencode"]
	require.True(t, ok, "opencode provider must exist")

	// Args (batch mode) may still contain "run" — that's fine.
	// PaneArgs (interactive TUI mode) must NOT contain "run".
	assert.NotEqual(t, opencode.Args, opencode.PaneArgs,
		"opencode Args and PaneArgs should differ: Args has 'run', PaneArgs does not (R1)")
}
