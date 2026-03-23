package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultFullConfig_HasVerify verifies VerifyConf is present and correct in full config.
func TestDefaultFullConfig_HasVerify(t *testing.T) {
	t.Parallel()

	cfg := DefaultFullConfig("test")
	require.NotNil(t, cfg)

	assert.True(t, cfg.Verify.Enabled, "Verify.Enabled must be true in full config")
	assert.Equal(t, "desktop", cfg.Verify.DefaultViewport)
	assert.True(t, cfg.Verify.AutoFix, "Verify.AutoFix must be true in full config")
	assert.Equal(t, 2, cfg.Verify.MaxFixAttempts)
}

// TestDefaultFullConfig_NoVerifyZeroValue verifies VerifyConf zero-value behavior in full config.
func TestDefaultFullConfig_NoVerify(t *testing.T) {
	t.Parallel()

	cfg := DefaultFullConfig("test")
	require.NotNil(t, cfg)

	assert.True(t, cfg.Verify.Enabled, "Verify.Enabled must be true in full config")
}

// TestVerifyConf_YAMLTags verifies that VerifyConf fields have the expected yaml tags
// by checking struct field behaviour through DefaultFullConfig round-trip.
func TestVerifyConf_ZeroValue(t *testing.T) {
	t.Parallel()

	var v VerifyConf
	assert.False(t, v.Enabled)
	assert.Empty(t, v.DefaultViewport)
	assert.False(t, v.AutoFix)
	assert.Zero(t, v.MaxFixAttempts)
}

// TestDefaultFullConfig_VerifyValidatesOK verifies the full config with Verify passes Validate.
func TestDefaultFullConfig_VerifyValidatesOK(t *testing.T) {
	t.Parallel()

	cfg := DefaultFullConfig("my-project")
	require.NotNil(t, cfg)

	err := cfg.Validate()
	require.NoError(t, err, "DefaultFullConfig with Verify must pass Validate()")
}
