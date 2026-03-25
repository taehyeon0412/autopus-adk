// Package pipeline provides pipeline state management types and persistence.
package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
)

// @AX:NOTE [AUTO] @AX:REASON: magic constants — ANSI escape sequences for terminal coloring; extend when adding new agent roles
const (
	colorCyan    = "\033[36m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorRed     = "\033[31m"
	colorMagenta = "\033[35m"
)

// @AX:NOTE [AUTO] @AX:REASON: magic constants — role-to-color mapping must stay in sync with ANSI constants above; add entries here when new roles are introduced
var roleColors = map[string]string{
	"lead":     colorCyan,
	"planner":  colorCyan,
	"builder":  colorGreen,
	"executor": colorGreen,
	"tester":   colorYellow,
	"guardian": colorRed,
	"reviewer": colorRed,
	"auditor":  colorMagenta,
}

// RoleColor returns the ANSI color code for the given agent role.
// Returns an empty string for unknown roles.
func RoleColor(role string) string {
	return roleColors[role]
}

// PipelineLogger writes pipeline events to JSONL and text log files.
type PipelineLogger struct {
	jsonlFile *os.File
	textFile  *os.File
	logDir    string
}

// NewPipelineLogger creates a PipelineLogger that writes to the given directory.
// It creates pipeline.jsonl and pipeline.log files in logDir.
func NewPipelineLogger(logDir string) (*PipelineLogger, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	jsonlFile, err := os.OpenFile(
		filepath.Join(logDir, "pipeline.jsonl"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644,
	)
	if err != nil {
		return nil, fmt.Errorf("open jsonl log: %w", err)
	}

	textFile, err := os.OpenFile(
		filepath.Join(logDir, "pipeline.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644,
	)
	if err != nil {
		jsonlFile.Close()
		return nil, fmt.Errorf("open text log: %w", err)
	}

	return &PipelineLogger{
		jsonlFile: jsonlFile,
		textFile:  textFile,
		logDir:    logDir,
	}, nil
}

// @AX:NOTE [AUTO] @AX:REASON: design choice — write failures silently ignored per R9; callers must not rely on LogEvent for error propagation
// LogEvent writes the event to both JSONL and text log files.
// Write failures are silently ignored (R9).
func (l *PipelineLogger) LogEvent(event Event) error {
	if l.jsonlFile != nil {
		if data, err := event.MarshalJSONL(); err == nil {
			_, _ = l.jsonlFile.Write(append(data, '\n'))
		}
	}

	if l.textFile != nil {
		line := formatTextLine(event)
		_, _ = l.textFile.WriteString(line)
	}

	return nil
}

// formatTextLine produces a human-readable log line for the event.
func formatTextLine(e Event) string {
	agent := e.Agent
	if agent == "" {
		agent = "-"
	}
	return fmt.Sprintf("[%s] [%s] [%s] %s\n",
		e.Timestamp.Format("15:04:05"), e.Phase, agent, e.Message)
}

// PromptInjection returns a formatted prompt section showing the log file path.
func (l *PipelineLogger) PromptInjection() string {
	return fmt.Sprintf("### Pipeline Monitor\nLog file: %s",
		filepath.Join(l.logDir, "pipeline.jsonl"))
}

// Close closes the underlying file handles.
func (l *PipelineLogger) Close() error {
	var firstErr error
	if l.jsonlFile != nil {
		if err := l.jsonlFile.Close(); err != nil {
			firstErr = err
		}
		l.jsonlFile = nil
	}
	if l.textFile != nil {
		if err := l.textFile.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		l.textFile = nil
	}
	return firstErr
}
