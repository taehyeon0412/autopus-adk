package issue

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/insajin/autopus-adk/pkg/version"
)

const (
	// maxTelemetryLines is the maximum number of lines read from telemetry files.
	maxTelemetryLines = 50
	// telemetryDir is the relative path to the telemetry directory.
	telemetryDir = ".autopus/telemetry"
)

// CollectContext gathers environment and error context for an issue report.
// It reads autopus.yaml and recent telemetry from the current working directory.
func CollectContext(errMsg, cmd string, exitCode int) IssueContext {
	ctx := IssueContext{
		ErrorMessage: errMsg,
		Command:      cmd,
		ExitCode:     exitCode,
		OS:           runtime.GOOS + "/" + runtime.GOARCH,
		GoVersion:    runtime.Version(),
		AutoVersion:  version.Version(),
		Platform:     detectPlatform(),
	}

	ctx.ConfigYAML = readConfig()
	ctx.Telemetry = readTelemetry()

	return ctx
}

// detectPlatform returns a best-effort platform identifier from env variables.
func detectPlatform() string {
	if os.Getenv("CLAUDE_CODE") != "" {
		return "claude-code"
	}
	if os.Getenv("CODEX") != "" {
		return "codex"
	}
	return "unknown"
}

// readConfig reads and sanitizes autopus.yaml from the current working directory.
func readConfig() string {
	data, err := os.ReadFile("autopus.yaml")
	if err != nil {
		return ""
	}
	return Sanitize(string(data))
}

// readTelemetry reads the most recent telemetry JSONL file and returns its last lines.
func readTelemetry() string {
	entries, err := os.ReadDir(telemetryDir)
	if err != nil {
		return ""
	}

	// Collect JSONL file names.
	var jsonlFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			jsonlFiles = append(jsonlFiles, e.Name())
		}
	}
	if len(jsonlFiles) == 0 {
		return ""
	}

	// Sort ascending then take the last one (most recent by name).
	sort.Strings(jsonlFiles)
	latest := jsonlFiles[len(jsonlFiles)-1]

	data, err := os.ReadFile(filepath.Join(telemetryDir, latest))
	if err != nil {
		return ""
	}

	return lastNLines(string(data), maxTelemetryLines)
}

// lastNLines returns the last n lines of s.
func lastNLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}
