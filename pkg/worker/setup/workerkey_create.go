// Package setup - workerkey_create.go: creates a Worker API Key via the backend REST API.
//
// After Device Auth OAuth, the setup wizard calls CreateWorkerAPIKey to exchange
// the short-lived JWT for a long-lived acos_worker_ key. The key is then stored
// in credentials.json so the worker daemon uses it for A2A WebSocket auth.
package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// workerKeyResponse is the payload returned by POST /api/v1/workspaces/:id/worker-keys.
type workerKeyResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Key       string  `json:"key"`
	ExpiresAt *string `json:"expires_at"`
}

// CreateWorkerAPIKey calls POST /api/v1/workspaces/:id/worker-keys with the JWT
// and returns the raw acos_worker_ key. The key is shown by the backend only once.
func CreateWorkerAPIKey(backendURL, jwtToken, workspaceID string) (string, error) {
	endpoint := strings.TrimRight(backendURL, "/") +
		"/api/v1/workspaces/" + workspaceID + "/worker-keys"

	body, err := json.Marshal(map[string]any{
		"name":         "adk-worker",
		"expires_days": 0, // never expires
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request worker key: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create worker key failed (%d): %s", resp.StatusCode, respBody)
	}

	keyResp, err := unwrap[workerKeyResponse](respBody)
	if err != nil {
		return "", fmt.Errorf("decode worker key response: %w", err)
	}
	if keyResp.Key == "" {
		return "", fmt.Errorf("backend returned empty key")
	}
	return keyResp.Key, nil
}
