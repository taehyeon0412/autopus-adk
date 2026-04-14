package mcpserver

import (
	"encoding/json"
	"fmt"
)

const (
	currentProtocolVersion = "2025-06-18"
	legacyProtocolVersion  = "2024-11-05"
)

func (s *MCPServer) handleInitialize(req *jsonRPCRequest) {
	var params struct {
		ProtocolVersion string         `json:"protocolVersion"`
		Capabilities    map[string]any `json:"capabilities,omitempty"`
		ClientInfo      map[string]any `json:"clientInfo,omitempty"`
	}
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, -32602, "invalid params")
			return
		}
	}

	version, ok := negotiateProtocolVersion(params.ProtocolVersion)
	if !ok {
		s.sendError(req.ID, -32602, fmt.Sprintf("unsupported protocol version: %s", params.ProtocolVersion))
		return
	}

	s.sendResult(req.ID, map[string]any{
		"protocolVersion": version,
		"serverInfo":      map[string]string{"name": "autopus-adk", "version": "0.1.0"},
		"capabilities": map[string]any{
			"tools":     map[string]any{},
			"resources": map[string]any{},
		},
		"instructions": "Autopus worker MCP server for task execution, workspace state, and knowledge lookup.",
	})
}

func negotiateProtocolVersion(clientVersion string) (string, bool) {
	switch clientVersion {
	case "", currentProtocolVersion:
		return currentProtocolVersion, true
	case legacyProtocolVersion:
		return legacyProtocolVersion, true
	default:
		return "", false
	}
}
