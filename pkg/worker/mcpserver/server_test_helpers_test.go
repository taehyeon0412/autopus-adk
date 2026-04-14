package mcpserver

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestServer(backendURL string) (*MCPServer, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	s := &MCPServer{
		backendURL:  backendURL,
		authToken:   "test-token",
		workspaceID: "ws-test",
		tools:       make(map[string]ToolHandler),
		writer:      buf,
	}
	s.resources = NewResourceRegistry(backendURL, "test-token")
	s.registerTools()
	return s, buf
}

func parseResponse(t *testing.T, buf *bytes.Buffer) jsonRPCResponse {
	t.Helper()
	var resp jsonRPCResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	return resp
}
