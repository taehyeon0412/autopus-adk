package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Framework represents a detected project framework.
type Framework struct {
	Name   string
	Stack  string
	Signal string
}

// DetectFramework detects the framework used in the given project directory.
// Returns nil if no known framework is detected.
// @AX:ANCHOR [AUTO] @AX:REASON: public API — called by setup and planner for profile assignment; do not change signature
func DetectFramework(dir string) (*Framework, error) {
	// Check config-file-based frameworks first (highest confidence)
	// @AX:NOTE [AUTO] @AX:REASON: magic constants — signal table defines framework detection rules; extend here when adding new framework support
	configSignals := []struct {
		glob      string
		name      string
		stack     string
		exclusive bool // if true, check no competing framework
	}{
		{"next.config.*", "nextjs", "typescript", false},
		{"nuxt.config.*", "nuxtjs", "typescript", false},
	}

	for _, sig := range configSignals {
		matches, err := filepath.Glob(filepath.Join(dir, sig.glob))
		if err != nil {
			return nil, err
		}
		if len(matches) > 0 {
			return &Framework{Name: sig.name, Stack: sig.stack, Signal: filepath.Base(matches[0])}, nil
		}
	}

	// Check package.json dependencies
	if fw := detectFrameworkFromPackageJSON(dir); fw != nil {
		return fw, nil
	}

	// Check Python frameworks
	if fw := detectPythonFramework(dir); fw != nil {
		return fw, nil
	}

	// Check Go frameworks from go.mod
	if fw := detectGoFramework(dir); fw != nil {
		return fw, nil
	}

	// Check Rust frameworks from Cargo.toml
	if fw := detectRustFramework(dir); fw != nil {
		return fw, nil
	}

	return nil, nil
}

// DetectStack detects the primary project stack from directory signals.
// @AX:NOTE [AUTO] @AX:REASON: public API boundary — R2 stack detection; priority order matters (go > ts > python > rust)
func DetectStack(dir string) (string, error) {
	if fileExists(filepath.Join(dir, "go.mod")) {
		return "go", nil
	}
	if fileExists(filepath.Join(dir, "package.json")) {
		return "typescript", nil
	}
	if fileExists(filepath.Join(dir, "pyproject.toml")) || fileExists(filepath.Join(dir, "requirements.txt")) {
		return "python", nil
	}
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		return "rust", nil
	}
	return "", nil
}

func detectFrameworkFromPackageJSON(dir string) *Framework {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	deps := mergeDeps(pkg.Dependencies, pkg.DevDependencies)

	// NestJS
	if _, ok := deps["@nestjs/core"]; ok {
		return &Framework{Name: "nestjs", Stack: "typescript", Signal: "@nestjs/core in package.json"}
	}

	// Frontend frameworks (check after config-file frameworks)
	if _, ok := deps["svelte"]; ok {
		return &Framework{Name: "svelte", Stack: "typescript", Signal: "svelte in package.json"}
	}
	if _, ok := deps["vue"]; ok {
		return &Framework{Name: "vue", Stack: "typescript", Signal: "vue in package.json"}
	}
	if _, ok := deps["react"]; ok {
		return &Framework{Name: "react", Stack: "typescript", Signal: "react in package.json"}
	}

	return nil
}

func detectPythonFramework(dir string) *Framework {
	for _, filename := range []string{"pyproject.toml", "requirements.txt"} {
		path := filepath.Join(dir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)

		pythonFrameworks := []struct {
			keyword string
			name    string
		}{
			{"fastapi", "fastapi"},
			{"django", "django"},
			{"flask", "flask"},
		}

		for _, fw := range pythonFrameworks {
			if strings.Contains(content, fw.keyword) {
				return &Framework{
					Name:   fw.name,
					Stack:  "python",
					Signal: fw.keyword + " in " + filename,
				}
			}
		}
	}
	return nil
}

func detectGoFramework(dir string) *Framework {
	path := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)

	goFrameworks := []struct {
		module string
		name   string
	}{
		{"gin-gonic/gin", "gin"},
		{"labstack/echo", "echo"},
		{"go-chi/chi", "chi"},
	}

	for _, fw := range goFrameworks {
		if strings.Contains(content, fw.module) {
			return &Framework{Name: fw.name, Stack: "go", Signal: fw.module + " in go.mod"}
		}
	}

	return nil
}

func detectRustFramework(dir string) *Framework {
	path := filepath.Join(dir, "Cargo.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if strings.Contains(string(data), "tokio-rs/axum") || strings.Contains(string(data), "axum") {
		return &Framework{Name: "axum", Stack: "rust", Signal: "axum in Cargo.toml"}
	}
	return nil
}

func mergeDeps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
