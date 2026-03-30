package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultProviderEntries_OpencodePromptViaArgsFalse verifies the canonical
// default entry for opencode has PromptViaArgs=false.
func TestDefaultProviderEntries_OpencodePromptViaArgsFalse(t *testing.T) {
	t.Parallel()

	entry, ok := defaultProviderEntries["opencode"]
	require.True(t, ok, "opencode must exist in defaultProviderEntries")
	assert.False(t, entry.PromptViaArgs,
		"opencode default must have PromptViaArgs=false")
}
