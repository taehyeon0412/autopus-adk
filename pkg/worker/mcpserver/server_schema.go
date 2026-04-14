package mcpserver

import "sort"

// toolDescriptor describes a tool for the tools/list response.
type toolDescriptor struct {
	Name         string         `json:"name"`
	Title        string         `json:"title,omitempty"`
	Description  string         `json:"description"`
	InputSchema  map[string]any `json:"inputSchema"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
	Annotations  map[string]any `json:"annotations,omitempty"`
}

func sortedToolNames(handlers map[string]ToolHandler) []string {
	names := make([]string, 0, len(handlers))
	for name := range handlers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func buildToolDescriptor(name string) toolDescriptor {
	return toolDescriptor{
		Name:         name,
		Title:        toolTitle(name),
		Description:  toolDescription(name),
		InputSchema:  toolInputSchema(name),
		OutputSchema: toolOutputSchema(name),
		Annotations:  toolAnnotations(name),
	}
}

func toolTitle(name string) string {
	switch name {
	case "execute_task":
		return "Execute Task"
	case "search_knowledge":
		return "Search Knowledge"
	case "get_execution_status":
		return "Get Execution Status"
	case "list_agents":
		return "List Agents"
	case "approve_execution":
		return "Approve Execution"
	case "manage_workspace":
		return "Manage Workspace"
	default:
		return ""
	}
}

func toolDescription(name string) string {
	switch name {
	case "execute_task":
		return "Create a new Autopus task for execution."
	case "search_knowledge":
		return "Search workspace knowledge and return matching documents."
	case "get_execution_status":
		return "Fetch the current status for an execution."
	case "list_agents":
		return "List available Autopus agents."
	case "approve_execution":
		return "Approve a pending execution."
	case "manage_workspace":
		return "Get or update workspace metadata."
	default:
		return name
	}
}

func toolInputSchema(name string) map[string]any {
	switch name {
	case "execute_task":
		return map[string]any{
			"type": "object",
			"properties": map[string]any{
				"description": map[string]any{"type": "string"},
				"prompt":      map[string]any{"type": "string"},
			},
			"required": []string{"description"},
		}
	case "search_knowledge":
		return map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":      map[string]any{"type": "string"},
				"limit":      map[string]any{"type": "integer", "minimum": 1},
				"categories": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			},
			"required": []string{"query"},
		}
	case "get_execution_status", "approve_execution":
		return map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string"},
			},
			"required": []string{"id"},
		}
	case "list_agents":
		return map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	case "manage_workspace":
		return map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":     map[string]any{"type": "string"},
				"action": map[string]any{"type": "string", "enum": []string{"get", "update"}},
				"data":   map[string]any{"type": "object"},
			},
			"required": []string{"action"},
		}
	default:
		return map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	}
}

func toolOutputSchema(name string) map[string]any {
	if name == "list_agents" {
		return nil
	}
	return map[string]any{"type": "object"}
}

func toolAnnotations(name string) map[string]any {
	switch name {
	case "search_knowledge", "get_execution_status", "list_agents":
		return map[string]any{"readOnlyHint": true}
	case "execute_task", "approve_execution", "manage_workspace":
		return map[string]any{
			"destructiveHint": false,
			"openWorldHint":   true,
		}
	default:
		return nil
	}
}
