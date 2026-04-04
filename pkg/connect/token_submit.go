package connect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// SubmitTokenRequest holds the data needed to submit a provider token to the server.
type SubmitTokenRequest struct {
	ProviderToken string `json:"provider_token"`
	RefreshToken  string `json:"refresh_token,omitempty"`
	WorkspaceID   string `json:"workspace_id"`
	Provider      string `json:"provider"`
}

// SubmitToken sends the provider OAuth token to the Autopus backend callback endpoint.
func (c *Client) SubmitToken(ctx context.Context, req SubmitTokenRequest) error {
	url := fmt.Sprintf("%s/api/v1/workspaces/%s/ai-oauth/callback",
		c.serverURL, req.WorkspaceID)

	payload := map[string]string{
		"provider":      req.Provider,
		"access_token":  req.ProviderToken,
		"refresh_token": req.RefreshToken,
		"nonce":         uuid.New().String(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("submit token: marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("submit token: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.authToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("submit token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submit token failed (%d): %s", resp.StatusCode, respBody)
	}
	return nil
}
