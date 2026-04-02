package tui

import (
	"fmt"
	"strings"
)

// SubagentNode represents one node in the subagent execution tree.
type SubagentNode struct {
	Name     string
	Status   string // "running", "completed", "failed"
	Children []*SubagentNode
}

// RenderTree renders the subagent tree using box-drawing characters.
func RenderTree(root *SubagentNode) string {
	if root == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(renderNode(root))
	b.WriteString("\n")
	renderChildren(&b, root.Children, "")
	return b.String()
}

func renderNode(node *SubagentNode) string {
	icon := statusIcon(node.Status)
	return fmt.Sprintf("%s %s", icon, node.Name)
}

func statusIcon(status string) string {
	switch status {
	case "running":
		return warningStyle.Render("⟳")
	case "completed":
		return successStyle.Render("✓")
	case "failed":
		return errorStyle.Render("✗")
	default:
		return mutedStyle.Render("○")
	}
}

func renderChildren(b *strings.Builder, children []*SubagentNode, prefix string) {
	for i, child := range children {
		isLast := i == len(children)-1

		var connector, childPrefix string
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		} else {
			connector = "├── "
			childPrefix = prefix + "│   "
		}

		b.WriteString(prefix)
		b.WriteString(connector)
		b.WriteString(renderNode(child))
		b.WriteString("\n")

		if len(child.Children) > 0 {
			renderChildren(b, child.Children, childPrefix)
		}
	}
}

// ParseStreamEvent parses a stream event into a SubagentNode.
// eventType: "agent_start", "agent_complete", "agent_fail"
// data: the agent name or identifier
func ParseStreamEvent(eventType, data string) *SubagentNode {
	node := &SubagentNode{Name: data}

	switch eventType {
	case "agent_start":
		node.Status = "running"
	case "agent_complete":
		node.Status = "completed"
	case "agent_fail":
		node.Status = "failed"
	default:
		node.Status = "running"
	}

	return node
}
