package mcp

import "encoding/json"

// ProgressReport represents a progress update sent via the MCP report_progress tool.
type ProgressReport struct {
	TaskID  string         `json:"task_id"`
	Phase   string         `json:"phase"`
	Status  string         `json:"status"`
	Details map[string]any `json:"details,omitempty"`
}

// ProgressReporter sends progress updates to the backend.
type ProgressReporter interface {
	ReportProgress(report ProgressReport) error
}

// ToolCall represents an MCP tool invocation request.
type ToolCall struct {
	Name   string          `json:"name"`
	Params json.RawMessage `json:"params"`
}

// ToolResult is the response from an MCP tool execution.
type ToolResult struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}
