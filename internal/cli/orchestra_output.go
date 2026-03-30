package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/insajin/autopus-adk/pkg/orchestra"
)

// saveOrchestraResult writes orchestra results to a timestamped markdown file
// under .autopus/orchestra/. Returns the file path on success.
func saveOrchestraResult(command, strategy string, providers []string, result *orchestra.OrchestraResult) (string, error) {
	dir := ".autopus/orchestra"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s/%s-%s-%s.md", dir, command, strategy, ts)

	header := fmt.Sprintf("# Orchestra: %s (%s)\n\n**Date**: %s  \n**Strategy**: %s  \n**Providers**: %s  \n**Duration**: %s\n\n---\n\n",
		command, strategy,
		time.Now().Format("2006-01-02 15:04:05"),
		strategy,
		strings.Join(providers, ", "),
		result.Duration.Round(time.Second))

	content := header + result.Merged
	return filename, os.WriteFile(filename, []byte(content), 0o644)
}
