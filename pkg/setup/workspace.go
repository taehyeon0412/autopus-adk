package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// DetectWorkspaces scans for monorepo workspace configurations.
func DetectWorkspaces(dir string) []Workspace {
	var workspaces []Workspace

	// Go workspaces (go.work)
	workspaces = append(workspaces, detectGoWorkspaces(dir)...)

	// npm/yarn/pnpm workspaces (package.json)
	workspaces = append(workspaces, detectNPMWorkspaces(dir)...)

	// Cargo workspaces (Cargo.toml)
	workspaces = append(workspaces, detectCargoWorkspaces(dir)...)

	return workspaces
}

func detectGoWorkspaces(dir string) []Workspace {
	goWorkPath := filepath.Join(dir, "go.work")
	data, err := os.ReadFile(goWorkPath)
	if err != nil {
		return nil
	}

	var workspaces []Workspace
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		// Parse "use" directives: use ./path or use ( ./path )
		if strings.HasPrefix(line, "use ") && !strings.Contains(line, "(") {
			wsPath := strings.TrimSpace(strings.TrimPrefix(line, "use "))
			wsPath = strings.TrimPrefix(wsPath, "./")
			if wsPath != "" {
				workspaces = append(workspaces, Workspace{
					Name: filepath.Base(wsPath),
					Path: wsPath,
					Type: "go.work",
				})
			}
		} else if !strings.HasPrefix(line, "use") && !strings.HasPrefix(line, ")") &&
			!strings.HasPrefix(line, "go ") && !strings.HasPrefix(line, "//") &&
			line != "" && line != "(" {
			// Lines inside use ( ... ) block
			wsPath := strings.TrimSpace(line)
			wsPath = strings.TrimPrefix(wsPath, "./")
			if wsPath != "" && !strings.HasPrefix(wsPath, "go ") {
				workspaces = append(workspaces, Workspace{
					Name: filepath.Base(wsPath),
					Path: wsPath,
					Type: "go.work",
				})
			}
		}
	}
	return workspaces
}

func detectNPMWorkspaces(dir string) []Workspace {
	pkgPath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}

	// Detect manager type
	wsType := "npm"
	if fileExists(filepath.Join(dir, "pnpm-workspace.yaml")) {
		wsType = "pnpm"
	} else if fileExists(filepath.Join(dir, "yarn.lock")) {
		wsType = "yarn"
	}

	var patterns []string

	// pnpm-workspace.yaml defines workspaces independently of package.json
	if wsType == "pnpm" {
		patterns = parsePnpmWorkspaces(filepath.Join(dir, "pnpm-workspace.yaml"))
	}

	// Also check package.json workspaces field
	if len(patterns) == 0 {
		var pkg struct {
			Workspaces interface{} `json:"workspaces"`
		}
		if err := json.Unmarshal(data, &pkg); err != nil || pkg.Workspaces == nil {
			return nil
		}

		switch v := pkg.Workspaces.(type) {
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok {
					patterns = append(patterns, s)
				}
			}
		case map[string]interface{}:
			// yarn workspace format: { "packages": ["packages/*"] }
			if pkgs, ok := v["packages"]; ok {
				if arr, ok := pkgs.([]interface{}); ok {
					for _, item := range arr {
						if s, ok := item.(string); ok {
							patterns = append(patterns, s)
						}
					}
				}
			}
		}
	}

	if len(patterns) == 0 {
		return nil
	}

	return resolveWorkspacePatterns(dir, patterns, wsType)
}

func parsePnpmWorkspaces(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			pattern := strings.TrimPrefix(line, "- ")
			pattern = strings.Trim(pattern, "'\"")
			if pattern != "" {
				patterns = append(patterns, pattern)
			}
		}
	}
	return patterns
}

func detectCargoWorkspaces(dir string) []Workspace {
	cargoPath := filepath.Join(dir, "Cargo.toml")
	data, err := os.ReadFile(cargoPath)
	if err != nil {
		return nil
	}

	content := string(data)
	if !strings.Contains(content, "[workspace]") {
		return nil
	}

	var workspaces []Workspace
	inMembers := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "members = [" || strings.HasPrefix(line, "members = [") {
			inMembers = true
			// Handle single-line: members = ["a", "b"]
			if strings.Contains(line, "]") {
				inner := line[strings.Index(line, "[")+1 : strings.Index(line, "]")]
				for _, item := range strings.Split(inner, ",") {
					member := strings.Trim(strings.TrimSpace(item), `"'`)
					if member != "" {
						workspaces = append(workspaces, resolveCargoMember(dir, member)...)
					}
				}
				inMembers = false
			}
			continue
		}
		if inMembers {
			if strings.Contains(line, "]") {
				inMembers = false
				continue
			}
			member := strings.Trim(strings.TrimRight(line, ","), `"' `)
			if member != "" {
				workspaces = append(workspaces, resolveCargoMember(dir, member)...)
			}
		}
	}
	return workspaces
}

func resolveCargoMember(dir, pattern string) []Workspace {
	if strings.Contains(pattern, "*") {
		return resolveWorkspacePatterns(dir, []string{pattern}, "cargo")
	}
	return []Workspace{{
		Name: filepath.Base(pattern),
		Path: pattern,
		Type: "cargo",
	}}
}

// resolveWorkspacePatterns expands glob patterns to actual workspace directories.
func resolveWorkspacePatterns(dir string, patterns []string, wsType string) []Workspace {
	var workspaces []Workspace
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Expand glob
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || !info.IsDir() {
				continue
			}
			rel, _ := filepath.Rel(dir, match)
			if seen[rel] {
				continue
			}
			seen[rel] = true
			workspaces = append(workspaces, Workspace{
				Name: filepath.Base(rel),
				Path: rel,
				Type: wsType,
			})
		}
	}
	return workspaces
}
