package connect_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insajin/autopus-adk/pkg/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmitToken_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		token       string
		workspaceID string
		provider    string
	}{
		{
			name:        "submit OpenAI token",
			token:       "sk-test-abc123",
			workspaceID: "ws-001",
			provider:    "openai",
		},
		{
			name:        "submit Anthropic token",
			token:       "sk-ant-test",
			workspaceID: "ws-002",
			provider:    "anthropic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Given: a mock server that accepts the token submission
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				expectedPath := "/api/v1/workspaces/" + tt.workspaceID + "/ai-oauth/callback"
				assert.Equal(t, expectedPath, r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer auth-token", r.Header.Get("Authorization"))

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				var payload map[string]string
				require.NoError(t, json.Unmarshal(body, &payload))
				assert.Equal(t, tt.provider, payload["provider"])
				assert.Equal(t, tt.token, payload["provider_token"])

				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			client := connect.NewClient("auth-token").WithServerURL(srv.URL)
			req := connect.SubmitTokenRequest{
				ProviderToken: tt.token,
				WorkspaceID:   tt.workspaceID,
				Provider:      tt.provider,
			}

			// When: submitting the token
			err := client.SubmitToken(context.Background(), req)

			// Then: no error
			require.NoError(t, err)
		})
	}
}

func TestSubmitToken_ServerError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    string
	}{
		{
			name:       "500 internal server error",
			statusCode: http.StatusInternalServerError,
			body:       "internal error",
			wantErr:    "submit token failed (500)",
		},
		{
			name:       "403 forbidden",
			statusCode: http.StatusForbidden,
			body:       "forbidden",
			wantErr:    "submit token failed (403)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Given: a server that returns an error status
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			client := connect.NewClient("auth-token").WithServerURL(srv.URL)
			req := connect.SubmitTokenRequest{
				ProviderToken: "sk-test",
				WorkspaceID:   "ws-001",
				Provider:      "openai",
			}

			// When: submitting the token
			err := client.SubmitToken(context.Background(), req)

			// Then: error should contain status code
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestSubmitToken_CancelledContext(t *testing.T) {
	t.Parallel()

	// Given: a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not reach server")
	}))
	defer srv.Close()

	client := connect.NewClient("auth-token").WithServerURL(srv.URL)
	req := connect.SubmitTokenRequest{
		ProviderToken: "sk-test",
		WorkspaceID:   "ws-001",
		Provider:      "openai",
	}

	// When: submitting with cancelled context
	err := client.SubmitToken(ctx, req)

	// Then: context error should be returned
	require.Error(t, err)
}
