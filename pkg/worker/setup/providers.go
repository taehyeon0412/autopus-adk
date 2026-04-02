package setup

import (
	"fmt"
	"os/exec"
	"strings"
)

// ProviderStatus describes the installation state of a CLI provider.
type ProviderStatus struct {
	Name      string
	Binary    string
	Installed bool
	Version   string
}

// providerPackages maps provider binary names to their npm package names.
var providerPackages = map[string]string{
	"claude":   "@anthropic-ai/claude-code",
	"codex":    "@openai/codex",
	"gemini":   "@google/gemini-cli",
	"opencode": "opencode",
}

// providerBinaries is the ordered list of provider binaries to detect.
var providerBinaries = []string{"claude", "codex", "gemini", "opencode"}

// DetectProviders checks which CLI providers are installed on the system.
func DetectProviders() []ProviderStatus {
	results := make([]ProviderStatus, 0, len(providerBinaries))
	for _, bin := range providerBinaries {
		ps := ProviderStatus{
			Name:   bin,
			Binary: bin,
		}

		path, err := exec.LookPath(bin)
		if err != nil {
			results = append(results, ps)
			continue
		}

		ps.Installed = true
		ps.Version = detectVersion(path)
		results = append(results, ps)
	}
	return results
}

// detectVersion runs "{binary} --version" and returns the output.
func detectVersion(binaryPath string) string {
	out, err := exec.Command(binaryPath, "--version").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// InstallProvider installs a provider via npm.
func InstallProvider(name string) error {
	pkg, ok := providerPackages[name]
	if !ok {
		return fmt.Errorf("unknown provider: %s", name)
	}

	if !checkNPM() {
		return fmt.Errorf("npm is not installed; install Node.js first")
	}

	cmd := exec.Command("npm", "install", "-g", pkg)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install %s (%s): %w", name, pkg, err)
	}
	return nil
}

// CheckNodeJS returns true if node is available on PATH.
func CheckNodeJS() bool {
	_, err := exec.LookPath("node")
	return err == nil
}

// checkNPM returns true if npm is available on PATH.
func checkNPM() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

// InstallNodeJS attempts to install Node.js via Homebrew (macOS).
func InstallNodeJS() error {
	brewPath, err := exec.LookPath("brew")
	if err != nil {
		return fmt.Errorf("brew not found; install Node.js manually from https://nodejs.org")
	}

	cmd := exec.Command(brewPath, "install", "node")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew install node: %w", err)
	}
	return nil
}
