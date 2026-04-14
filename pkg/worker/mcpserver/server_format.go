package mcpserver

import (
	"encoding/json"
	"fmt"
)

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content           []textContent `json:"content"`
	StructuredContent any           `json:"structuredContent,omitempty"`
	IsError           bool          `json:"isError,omitempty"`
}

type resourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

func formatToolResult(result any) toolResult {
	text, structured := renderValue(result)
	resp := toolResult{
		Content: []textContent{{
			Type: "text",
			Text: text,
		}},
	}
	if structuredMap, ok := structured.(map[string]any); ok {
		resp.StructuredContent = structuredMap
	}
	return resp
}

func formatToolError(err error) toolResult {
	return toolResult{
		Content: []textContent{{
			Type: "text",
			Text: err.Error(),
		}},
		IsError: true,
	}
}

func formatResourceContent(uri string, value any) resourceContent {
	text, _ := renderValue(value)
	mimeType := "application/json"
	if _, ok := value.(string); ok {
		mimeType = "text/plain"
	}
	return resourceContent{
		URI:      uri,
		MimeType: mimeType,
		Text:     text,
	}
}

func renderValue(value any) (string, any) {
	switch v := value.(type) {
	case nil:
		return "null", nil
	case string:
		return v, nil
	default:
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Sprint(v), nil
		}
		return string(data), v
	}
}
