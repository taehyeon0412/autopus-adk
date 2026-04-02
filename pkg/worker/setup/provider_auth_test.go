package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckProviderAuth_UnknownProvider(t *testing.T) {
	t.Parallel()

	authenticated, guide := CheckProviderAuth("unknown-provider")
	assert.False(t, authenticated)
	assert.Contains(t, guide, "Unknown provider")
}

func TestCheckProviderAuth_Claude_NoCredentials(t *testing.T) {
	// Use a temp dir as HOME so credentials.json won't exist.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	authenticated, guide := CheckProviderAuth("claude")
	assert.False(t, authenticated)
	assert.Contains(t, guide, "claude login")
}

func TestCheckProviderAuth_Claude_WithCredentials(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Create the credentials file.
	credDir := filepath.Join(tmp, ".claude")
	err := os.MkdirAll(credDir, 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(credDir, "credentials.json"), []byte("{}"), 0600)
	assert.NoError(t, err)

	authenticated, guide := CheckProviderAuth("claude")
	assert.True(t, authenticated)
	assert.Empty(t, guide)
}

func TestCheckProviderAuth_Codex_WithEnvVar(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test-key")

	authenticated, guide := CheckProviderAuth("codex")
	assert.True(t, authenticated)
	assert.Empty(t, guide)
}

func TestCheckProviderAuth_Gemini_WithEnvVar(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "test-google-key")

	authenticated, guide := CheckProviderAuth("gemini")
	assert.True(t, authenticated)
	assert.Empty(t, guide)
}
