package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContextConf_ZeroValue verifies that a zero-value ContextConf has all fields false.
func TestContextConf_ZeroValue(t *testing.T) {
	t.Parallel()

	var c ContextConf
	assert.False(t, c.SignatureMap)
}

// TestDefaultFullConfig_HasContext verifies ContextConf is set correctly in full config.
func TestDefaultFullConfig_HasContext(t *testing.T) {
	t.Parallel()

	cfg := DefaultFullConfig("test")
	require.NotNil(t, cfg)

	assert.True(t, cfg.Context.SignatureMap, "Context.SignatureMap must be true in full config")
}

// TestDefaultLiteConfig_HasContext verifies ContextConf is set correctly in lite config.
func TestDefaultLiteConfig_HasContext(t *testing.T) {
	t.Parallel()

	cfg := DefaultLiteConfig("test")
	require.NotNil(t, cfg)

	assert.True(t, cfg.Context.SignatureMap, "Context.SignatureMap must be true in lite config")
}

// TestDefaultFullConfig_ContextValidatesOK verifies the full config with Context passes Validate.
func TestDefaultFullConfig_ContextValidatesOK(t *testing.T) {
	t.Parallel()

	cfg := DefaultFullConfig("my-project")
	require.NotNil(t, cfg)

	err := cfg.Validate()
	require.NoError(t, err, "DefaultFullConfig with Context must pass Validate()")
}

// TestHarnessConfig_ContextField verifies HarnessConfig embeds ContextConf with correct field.
func TestHarnessConfig_ContextField(t *testing.T) {
	t.Parallel()

	cfg := HarnessConfig{
		Mode:        ModeFull,
		ProjectName: "test",
		Platforms:   []string{"claude-code"},
		Context: ContextConf{
			SignatureMap: true,
		},
	}

	err := cfg.Validate()
	require.NoError(t, err)
	assert.True(t, cfg.Context.SignatureMap)
}
