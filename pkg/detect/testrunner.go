package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// packageJSON represents the relevant fields in a package.json file.
type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// DetectTestRunner detects the test runner used in the given project directory.
// Returns the runner name ("vitest", "jest", "go test", "pytest", "cargo test")
// or an empty string if no known test runner is found.
func DetectTestRunner(dir string) (string, error) {
	// Priority 1: vitest config file takes precedence over jest
	for _, name := range []string{"vitest.config.ts", "vitest.config.js"} {
		if fileExists(filepath.Join(dir, name)) {
			return "vitest", nil
		}
	}

	// Priority 2: jest in package.json dependencies
	runner, err := detectFromPackageJSON(filepath.Join(dir, "package.json"))
	if err != nil {
		return "", err
	}
	if runner != "" {
		return runner, nil
	}

	// Priority 3: go.mod
	if fileExists(filepath.Join(dir, "go.mod")) {
		return "go test", nil
	}

	// Priority 4: pytest config
	if fileExists(filepath.Join(dir, "pytest.ini")) {
		return "pytest", nil
	}
	if fileExists(filepath.Join(dir, "pyproject.toml")) && hasPytest(filepath.Join(dir, "pyproject.toml")) {
		return "pytest", nil
	}

	// Priority 5: Cargo.toml
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		return "cargo test", nil
	}

	return "", nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func detectFromPackageJSON(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", nil // malformed package.json: skip silently
	}

	if _, ok := pkg.DevDependencies["jest"]; ok {
		return "jest", nil
	}
	if _, ok := pkg.Dependencies["jest"]; ok {
		return "jest", nil
	}
	return "", nil
}

func hasPytest(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "pytest")
}
