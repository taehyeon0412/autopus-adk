package setup

import (
	"os"
	"path/filepath"
)

// CheckProviderAuth verifies whether a provider has valid credentials.
// Returns (true, "") if authenticated, or (false, guide) with instructions.
func CheckProviderAuth(name string) (authenticated bool, guide string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, "Cannot determine home directory"
	}

	switch name {
	case "claude":
		return checkClaude(home)
	case "codex":
		return checkCodex(home)
	case "gemini":
		return checkGemini(home)
	case "opencode":
		return checkOpencode()
	default:
		return false, "Unknown provider: " + name
	}
}

func checkClaude(home string) (bool, string) {
	credPath := filepath.Join(home, ".claude", "credentials.json")
	if fileExists(credPath) {
		return true, ""
	}
	return false, "Run `claude login` to authenticate"
}

func checkCodex(home string) (bool, string) {
	if os.Getenv("OPENAI_API_KEY") != "" {
		return true, ""
	}
	codexDir := filepath.Join(home, ".codex")
	if dirExists(codexDir) {
		return true, ""
	}
	return false, "Set OPENAI_API_KEY or run `codex login` to authenticate"
}

func checkGemini(home string) (bool, string) {
	if os.Getenv("GOOGLE_API_KEY") != "" {
		return true, ""
	}
	geminiDir := filepath.Join(home, ".config", "gemini")
	if dirExists(geminiDir) {
		return true, ""
	}
	return false, "Set GOOGLE_API_KEY or run `gemini login` to authenticate"
}

func checkOpencode() (bool, string) {
	// opencode uses the same key as codex (OpenAI)
	if os.Getenv("OPENAI_API_KEY") != "" {
		return true, ""
	}
	return false, "Set OPENAI_API_KEY to authenticate opencode"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
