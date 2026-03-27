package orchestra

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// RoundSignalName generates a round-scoped signal filename.
// Format: "{provider}-round{N}-{suffix}" (e.g., "claude-round2-done").
func RoundSignalName(provider string, round int, suffix string) string {
	return fmt.Sprintf("%s-round%d-%s", sanitizeProviderName(provider), round, suffix)
}

// CleanRoundSignals removes done signal files for the given round,
// preserving result files. Scans the session directory for files
// matching the "*-round{N}-done" pattern.
func CleanRoundSignals(session *HookSession, round int) {
	pattern := fmt.Sprintf("*-round%d-done", round)
	matches, err := filepath.Glob(filepath.Join(session.Dir(), pattern))
	if err != nil {
		return
	}
	for _, m := range matches {
		_ = os.Remove(m)
	}
}

// SetRoundEnv sets the AUTOPUS_ROUND environment variable to the current round number.
func SetRoundEnv(round int) {
	_ = os.Setenv("AUTOPUS_ROUND", fmt.Sprintf("%d", round))
}

// SendRoundEnvToPane sends "export AUTOPUS_ROUND=N" to the specified terminal pane.
func SendRoundEnvToPane(ctx context.Context, term terminal.Terminal, paneID terminal.PaneID, round int) error {
	cmd := fmt.Sprintf("export AUTOPUS_ROUND=%d", round)
	return term.SendCommand(ctx, paneID, cmd)
}
