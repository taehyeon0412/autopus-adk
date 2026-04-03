package setup

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DeviceCode holds the response from the device authorization endpoint.
type DeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// TokenResponse holds the OAuth token returned after successful authorization.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// GeneratePKCE creates a PKCE code verifier and its S256 challenge.
// The verifier is 32 random bytes, base64url-encoded.
// The challenge is SHA256(verifier), base64url-encoded.
func GeneratePKCE() (verifier, challenge string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(buf)

	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return verifier, challenge, nil
}

// RequestDeviceCode initiates the device authorization flow.
func RequestDeviceCode(backendURL, codeVerifier string) (*DeviceCode, error) {
	endpoint := strings.TrimRight(backendURL, "/") + "/api/v1/auth/device/code"

	_, challenge, err := deriveChallengeFromVerifier(codeVerifier)
	if err != nil {
		return nil, err
	}

	payload := map[string]string{
		"code_challenge":        challenge,
		"code_challenge_method": "S256",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device code request failed (%d): %s", resp.StatusCode, respBody)
	}

	var dc DeviceCode
	if err := json.NewDecoder(resp.Body).Decode(&dc); err != nil {
		return nil, fmt.Errorf("decode device code response: %w", err)
	}
	return &dc, nil
}

// deriveChallengeFromVerifier recomputes the S256 challenge from an existing verifier.
func deriveChallengeFromVerifier(verifier string) (string, string, error) {
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])
	return verifier, challenge, nil
}

// PollForToken polls the token endpoint until the user authorizes or the context is cancelled.
// codeVerifier is required when PKCE was used in the device code request.
func PollForToken(ctx context.Context, backendURL, deviceCode, codeVerifier string, interval int) (*TokenResponse, error) {
	endpoint := strings.TrimRight(backendURL, "/") + "/api/v1/auth/device/token"
	if interval <= 0 {
		interval = 5
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			token, pending, err := tryTokenExchange(endpoint, deviceCode, codeVerifier)
			if err != nil {
				return nil, err
			}
			if pending {
				continue
			}
			return token, nil
		}
	}
}

// tryTokenExchange attempts a single token exchange request.
func tryTokenExchange(endpoint, deviceCode, codeVerifier string) (*TokenResponse, bool, error) {
	payload := map[string]string{
		"device_code":   deviceCode,
		"code_verifier": codeVerifier,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, false, fmt.Errorf("marshal token request: %w", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("poll token: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusBadRequest {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error == "authorization_pending" {
			return nil, true, nil
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("token request failed (%d): %s", resp.StatusCode, respBody)
	}

	var token TokenResponse
	if err := json.Unmarshal(respBody, &token); err != nil {
		return nil, false, fmt.Errorf("decode token response: %w", err)
	}
	return &token, false, nil
}

// RefreshToken exchanges a refresh token for a new access token via the backend.
func RefreshToken(ctx context.Context, backendURL, refreshToken string) (*TokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, backendURL+"/oauth/token", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed: HTTP %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}
	return &token, nil
}

// OpenBrowser opens the given URL in the default browser.
func OpenBrowser(u string) error {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return exec.Command(cmd, u).Start()
}

// SaveCredentials writes credentials to ~/.config/autopus/credentials.json.
func SaveCredentials(creds map[string]any) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	dir := filepath.Join(home, ".config", "autopus")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	path := filepath.Join(dir, "credentials.json")
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}
	return nil
}
