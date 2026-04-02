package setup

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePKCE_VerifierLength(t *testing.T) {
	t.Parallel()

	verifier, _, err := GeneratePKCE()
	require.NoError(t, err)

	// 32 bytes base64url-encoded without padding = 43 characters
	assert.Len(t, verifier, 43)
}

func TestGeneratePKCE_VerifierIsBase64URL(t *testing.T) {
	t.Parallel()

	verifier, _, err := GeneratePKCE()
	require.NoError(t, err)

	// Decoding should succeed for valid base64url
	decoded, err := base64.RawURLEncoding.DecodeString(verifier)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
}

func TestGeneratePKCE_ChallengeIsSHA256OfVerifier(t *testing.T) {
	t.Parallel()

	verifier, challenge, err := GeneratePKCE()
	require.NoError(t, err)

	// Recompute: challenge = base64url(SHA256(verifier))
	h := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(h[:])

	assert.Equal(t, expected, challenge)
}

func TestGeneratePKCE_UniquePairs(t *testing.T) {
	t.Parallel()

	v1, c1, err := GeneratePKCE()
	require.NoError(t, err)

	v2, c2, err := GeneratePKCE()
	require.NoError(t, err)

	assert.NotEqual(t, v1, v2, "verifiers should be unique")
	assert.NotEqual(t, c1, c2, "challenges should be unique")
}

func TestGeneratePKCE_ChallengeLength(t *testing.T) {
	t.Parallel()

	_, challenge, err := GeneratePKCE()
	require.NoError(t, err)

	// SHA256 = 32 bytes → base64url without padding = 43 characters
	assert.Len(t, challenge, 43)
}

func TestDeriveChallengeFromVerifier(t *testing.T) {
	t.Parallel()

	verifier, originalChallenge, err := GeneratePKCE()
	require.NoError(t, err)

	_, derivedChallenge, err := deriveChallengeFromVerifier(verifier)
	require.NoError(t, err)

	assert.Equal(t, originalChallenge, derivedChallenge)
}

func TestRefreshToken_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/oauth/token", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		})
	}))
	defer srv.Close()

	token, err := RefreshToken(context.Background(), srv.URL, "old-refresh-token")
	require.NoError(t, err)
	assert.Equal(t, "new-access-token", token.AccessToken)
	assert.Equal(t, "new-refresh-token", token.RefreshToken)
	assert.Equal(t, 3600, token.ExpiresIn)
}

func TestRefreshToken_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := RefreshToken(context.Background(), srv.URL, "bad-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestRefreshToken_InvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid-json"))
	}))
	defer srv.Close()

	_, err := RefreshToken(context.Background(), srv.URL, "some-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode")
}

func TestSaveCredentials_WritesFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	creds := map[string]any{
		"access_token": "test-token",
		"expires_in":   3600,
	}
	err := SaveCredentials(creds)
	require.NoError(t, err)

	// Verify the file was created.
	path := filepath.Join(tmp, ".config", "autopus", "credentials.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got map[string]any
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	assert.Equal(t, "test-token", got["access_token"])
}

func TestSaveCredentials_Permissions(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	err := SaveCredentials(map[string]any{"key": "val"})
	require.NoError(t, err)

	path := filepath.Join(tmp, ".config", "autopus", "credentials.json")
	info, err := os.Stat(path)
	require.NoError(t, err)
	// Verify file permissions are 0600 (owner read/write only).
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}
