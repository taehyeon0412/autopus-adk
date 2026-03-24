// Package e2e provides user-facing scenario-based E2E test infrastructure.
package e2e

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

// ErrSkipScenario indicates a scenario should be skipped (e.g., missing API key in CI).
var ErrSkipScenario = errors.New("scenario skipped: required env not available")

// EnvResolveOptions configures the environment resolver.
type EnvResolveOptions struct {
	ProjectDir     string            // project root directory
	ScenarioEnv    map[string]string // per-scenario env overrides
	NonInteractive bool              // true if TTY is NOT available (CI mode)
	TestEnvFile    string            // path to .autopus/test.env (optional)
	RequiredVars   []string          // env vars that must be resolved
}

// @AX:NOTE [AUTO] @AX:REASON: magic constants — domain-specific safe defaults for test env; update when adding new standard service ports or DB types
var safeDefaults = map[string]string{
	"DATABASE_URL": "sqlite://test.db",
	"PORT":         "8080",
	"HOST":         "localhost",
	"LOG_LEVEL":    "debug",
	"ENV":          "test",
	"APP_ENV":      "test",
	"NODE_ENV":     "test",
}

// @AX:NOTE [AUTO] @AX:REASON: magic constants — security-sensitive keyword list for secret detection; extend with care to avoid false negatives
var secretKeywords = []string{
	"API_KEY", "SECRET", "TOKEN", "PASSWORD", "PASSWD", "PRIVATE_KEY", "CREDENTIALS",
}

// isSecret reports whether the env var key is likely a secret.
func isSecret(key string) bool {
	upper := strings.ToUpper(key)
	for _, kw := range secretKeywords {
		if strings.Contains(upper, kw) {
			return true
		}
	}
	return false
}

// goEnvVars resolves Go toolchain environment variables (GOPATH, GOROOT, etc.).
// These may not be set as OS env vars but are computed by `go env`.
func goEnvVars() map[string]string {
	keys := []string{"GOPATH", "GOROOT", "GOMODCACHE", "GOCACHE"}
	result := make(map[string]string, len(keys))
	out, err := exec.Command("go", append([]string{"env"}, keys...)...).Output()
	if err != nil {
		for _, k := range keys {
			if v := os.Getenv(k); v != "" {
				result[k] = v
			}
		}
		return result
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if i < len(keys) {
			if v := strings.TrimSpace(line); v != "" {
				result[keys[i]] = v
			}
		}
	}
	return result
}

// detectEnvFromProject scans .env.example and docker-compose.yml to auto-detect
// env var names and their example values.
func detectEnvFromProject(dir string) map[string]string {
	env := make(map[string]string)

	// Pull well-known Go env vars via the toolchain (handles computed defaults).
	for k, v := range goEnvVars() {
		if v != "" {
			env[k] = v
		}
	}

	// Parse .env.example
	envExamplePath := filepath.Join(dir, ".env.example")
	if f, err := os.Open(envExamplePath); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if idx := strings.IndexByte(line, '='); idx > 0 {
				key := strings.TrimSpace(line[:idx])
				val := strings.TrimSpace(line[idx+1:])
				if key != "" {
					env[key] = val
				}
			}
		}
	}

	// Parse docker-compose.yml for environment keys.
	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		composePath := filepath.Join(dir, name)
		if f, err := os.Open(composePath); err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			inEnvBlock := false
			for scanner.Scan() {
				line := scanner.Text()
				trimmed := strings.TrimSpace(line)

				if trimmed == "environment:" {
					inEnvBlock = true
					continue
				}
				// Exit env block on unindented non-empty line that is not a list item.
				if inEnvBlock && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != "" {
					inEnvBlock = false
				}
				if !inEnvBlock {
					continue
				}
				// Handle "- KEY=VALUE" or "KEY=VALUE" or "- KEY" forms.
				item := strings.TrimPrefix(trimmed, "- ")
				if idx := strings.IndexByte(item, '='); idx > 0 {
					key := strings.TrimSpace(item[:idx])
					val := strings.TrimSpace(item[idx+1:])
					if key != "" {
						env[key] = val
					}
				} else if item != "" && !strings.Contains(item, " ") {
					env[item] = ""
				}
			}
			break
		}
	}

	return env
}

// applySafeDefaults applies well-known safe default values for recognized env var names.
// It does NOT overwrite keys that already have a non-empty value.
func applySafeDefaults(env map[string]string) map[string]string {
	result := make(map[string]string, len(env))
	for k, v := range env {
		result[k] = v
	}
	for key, def := range safeDefaults {
		if _, exists := result[key]; !exists {
			result[key] = def
		}
	}
	return result
}

// loadTestEnvFile reads a .autopus/test.env file (KEY=VALUE format) into a map.
func loadTestEnvFile(path string) map[string]string {
	env := make(map[string]string)
	if path == "" {
		return env
	}
	f, err := os.Open(path)
	if err != nil {
		return env
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			if key != "" {
				env[key] = val
			}
		}
	}
	return env
}

// isTTY reports whether os.Stdin is connected to a terminal.
func isTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// promptForSecret interactively prompts the user to enter a secret value.
// Uses x/term.ReadPassword for echo-suppressed input.
func promptForSecret(key string) (string, error) {
	fmt.Fprintf(os.Stderr, "Enter value for %s (leave blank to skip): ", key)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after hidden input
	if err != nil {
		// Fallback to plain read if terminal is not available.
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return strings.TrimSpace(scanner.Text()), nil
		}
		return "", scanner.Err()
	}
	return strings.TrimSpace(string(password)), nil
}

// @AX:NOTE [AUTO] @AX:REASON: public API boundary — 4-layer env merge contract (project detection → safe defaults → per-scenario override → test.env); layer order must not be changed; fan_in=0 (currently unused outside package tests)
// ResolveEnv resolves environment variables using the 4-layer hierarchy.
// Returns merged env map and nil error on success.
// Returns ErrSkipScenario if a required secret is missing in non-interactive mode.
func ResolveEnv(opts EnvResolveOptions) (map[string]string, error) {
	// Layer 1: auto-detect from project files.
	env := detectEnvFromProject(opts.ProjectDir)

	// Layer 2: apply safe defaults (overrides auto-detected empty values for known keys).
	env = applySafeDefaults(env)

	// Layer 3: per-scenario overrides.
	for k, v := range opts.ScenarioEnv {
		env[k] = v
	}

	// Layer 4: test.env file (highest priority).
	testEnvPath := opts.TestEnvFile
	if testEnvPath == "" && opts.ProjectDir != "" {
		testEnvPath = filepath.Join(opts.ProjectDir, ".autopus", "test.env")
	}
	for k, v := range loadTestEnvFile(testEnvPath) {
		env[k] = v
	}

	// Validate required vars and handle secrets.
	interactive := !opts.NonInteractive && isTTY()
	for _, key := range opts.RequiredVars {
		val, exists := env[key]
		if exists && val != "" {
			continue
		}
		if !isSecret(key) {
			// Non-secret missing var: use empty string, not a skip.
			continue
		}
		// Secret key is missing.
		if interactive {
			prompted, err := promptForSecret(key)
			if err != nil {
				return nil, err
			}
			if prompted != "" {
				env[key] = prompted
				continue
			}
		}
		// Non-interactive or user skipped: return skip error.
		return nil, fmt.Errorf("%w: %s not set in non-interactive mode", ErrSkipScenario, key)
	}

	return env, nil
}
