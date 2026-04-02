package security

import (
	"fmt"
	"os"
)

// SetupAutonomousMode configures the subprocess environment for autonomous execution.
// It injects PreToolUse hooks and prepares the --dangerously-skip-permissions flag.
func SetupAutonomousMode(workDir string, policyPath string) error {
	// Verify the policy file exists and is readable.
	f, err := os.Open(policyPath)
	if err != nil {
		return fmt.Errorf("policy file not accessible: %w", err)
	}
	f.Close()

	// Inject PreToolUse hooks into the working directory's settings.
	if err := WriteHookConfig(workDir, policyPath); err != nil {
		return fmt.Errorf("setup autonomous hooks: %w", err)
	}
	return nil
}

// CleanupAutonomousMode removes injected hook configuration.
func CleanupAutonomousMode(workDir string) error {
	if err := RemoveHookConfig(workDir); err != nil {
		return fmt.Errorf("cleanup autonomous hooks: %w", err)
	}
	return nil
}

// AutonomousFlags returns additional CLI flags for autonomous subprocess execution.
func AutonomousFlags() []string {
	return []string{"--dangerously-skip-permissions"}
}
