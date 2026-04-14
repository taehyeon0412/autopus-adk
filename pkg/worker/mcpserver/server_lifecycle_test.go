package mcpserver

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServer_Initialize(t *testing.T) {
	t.Parallel()

	s, buf := newTestServer("http://localhost")
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}`),
	}
	s.dispatch(t.Context(), req)

	resp := parseResponse(t, buf)
	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(1), resp.ID)
	assert.Nil(t, resp.Error)

	result, ok := resp.Result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "2025-06-18", result["protocolVersion"])

	serverInfo := result["serverInfo"].(map[string]any)
	assert.Equal(t, "autopus-adk", serverInfo["name"])
	assert.NotEmpty(t, result["instructions"])
}

func TestMCPServer_InitializeLegacyProtocol(t *testing.T) {
	t.Parallel()

	s, buf := newTestServer("http://localhost")
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}`),
	}
	s.dispatch(t.Context(), req)

	resp := parseResponse(t, buf)
	require.Nil(t, resp.Error)

	result, ok := resp.Result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "2024-11-05", result["protocolVersion"])
}

func TestMCPServer_InitializeUnsupportedProtocol(t *testing.T) {
	t.Parallel()

	s, buf := newTestServer("http://localhost")
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"1999-01-01","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}`),
	}
	s.dispatch(t.Context(), req)

	resp := parseResponse(t, buf)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32602, resp.Error.Code)
}

func TestMCPServer_UnknownMethod(t *testing.T) {
	t.Parallel()

	s, buf := newTestServer("http://localhost")
	req := &jsonRPCRequest{JSONRPC: "2.0", ID: float64(3), Method: "bogus/method"}
	s.dispatch(t.Context(), req)

	resp := parseResponse(t, buf)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32601, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "method not found")
}

func TestMCPServer_InitializedNotification_IsIgnored(t *testing.T) {
	t.Parallel()

	s, buf := newTestServer("http://localhost")
	req := &jsonRPCRequest{JSONRPC: "2.0", Method: "notifications/initialized"}
	s.dispatch(t.Context(), req)

	assert.Empty(t, buf.String(), "notifications must not produce a JSON-RPC response")
}

func TestMCPServer_UnknownNotification_IsIgnored(t *testing.T) {
	t.Parallel()

	s, buf := newTestServer("http://localhost")
	req := &jsonRPCRequest{JSONRPC: "2.0", Method: "notifications/progress"}
	s.dispatch(t.Context(), req)

	assert.Empty(t, buf.String(), "unknown notifications must be ignored")
}

func TestMCPServer_Ping(t *testing.T) {
	t.Parallel()

	s, buf := newTestServer("http://localhost")
	req := &jsonRPCRequest{JSONRPC: "2.0", ID: float64(8), Method: "ping"}
	s.dispatch(t.Context(), req)

	resp := parseResponse(t, buf)
	assert.Equal(t, float64(8), resp.ID)
	assert.Nil(t, resp.Error)

	result, ok := resp.Result.(map[string]any)
	require.True(t, ok)
	assert.Empty(t, result)
}

func TestMCPServer_SendError(t *testing.T) {
	t.Parallel()

	s, buf := newTestServer("http://localhost")
	s.sendError("req-1", -32600, "invalid request")

	resp := parseResponse(t, buf)
	assert.Equal(t, "req-1", resp.ID)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32600, resp.Error.Code)
	assert.Equal(t, "invalid request", resp.Error.Message)
}
