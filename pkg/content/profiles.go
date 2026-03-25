package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProfileDefinition represents an executor profile loaded from a markdown file.
// @AX:ANCHOR [AUTO] @AX:REASON: core data type — ProfileDefinition is consumed by planner, executor, and setup; do not rename or restructure without coordinating all consumers
type ProfileDefinition struct {
	Name          string   `yaml:"name"`
	Stack         string   `yaml:"stack"`
	Framework     string   `yaml:"framework,omitempty"`
	Extends       string   `yaml:"extends,omitempty"`
	Tools         []string `yaml:"tools,omitempty"`
	TestFramework string   `yaml:"test_framework,omitempty"`
	Linter        string   `yaml:"linter,omitempty"`
	Instructions  string   `yaml:"-"`
	Source        string   `yaml:"-"`
}

// profileFrontmatter is the internal struct for YAML frontmatter parsing.
type profileFrontmatter struct {
	Name          string   `yaml:"name"`
	Stack         string   `yaml:"stack"`
	Framework     string   `yaml:"framework,omitempty"`
	Extends       string   `yaml:"extends,omitempty"`
	Tools         []string `yaml:"tools,omitempty"`
	TestFramework string   `yaml:"test_framework,omitempty"`
	Linter        string   `yaml:"linter,omitempty"`
}

// LoadProfilesFromFS loads profile definitions from an embedded filesystem.
func LoadProfilesFromFS(fsys fs.FS, dir string) ([]ProfileDefinition, error) {
	profiles, _, err := LoadProfilesFromFSWithWarnings(fsys, dir)
	return profiles, err
}

// LoadProfilesFromFSWithWarnings loads profiles and returns warnings for skipped files.
// @AX:NOTE [AUTO] @AX:REASON: public API boundary — primary profile loading entry point; R11 validation (skip missing name/stack) happens inside parseProfileData
func LoadProfilesFromFSWithWarnings(fsys fs.FS, dir string) ([]ProfileDefinition, []string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, nil, fmt.Errorf("profile directory read failed %s: %w", dir, err)
	}

	var profiles []ProfileDefinition
	var warnings []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, err := fs.ReadFile(fsys, dir+"/"+entry.Name())
		if err != nil {
			return nil, warnings, fmt.Errorf("profile file read failed %s: %w", entry.Name(), err)
		}

		profile, warn, err := parseProfileData(data, entry.Name(), "builtin")
		if err != nil {
			return nil, warnings, fmt.Errorf("profile file parse failed %s: %w", entry.Name(), err)
		}
		if warn != "" {
			warnings = append(warnings, warn)
			continue
		}
		profiles = append(profiles, profile)
	}

	return profiles, warnings, nil
}

// LoadProfilesFromDir loads profile definitions from an OS filesystem directory.
func LoadProfilesFromDir(dir string) ([]ProfileDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("profile directory read failed %s: %w", dir, err)
	}

	var profiles []ProfileDefinition
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("profile file read failed %s: %w", entry.Name(), err)
		}

		profile, _, err := parseProfileData(data, entry.Name(), "custom")
		if err != nil {
			return nil, fmt.Errorf("profile file parse failed %s: %w", entry.Name(), err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// ResolveExtendsProfile resolves extends inheritance for a profile.
// Base Instructions are prepended; child frontmatter fields override base.
func ResolveExtendsProfile(profile ProfileDefinition, allProfiles map[string]ProfileDefinition) (ProfileDefinition, error) {
	if profile.Extends == "" {
		return profile, nil
	}

	base, ok := allProfiles[profile.Extends]
	if !ok {
		return profile, fmt.Errorf("base profile %q not found", profile.Extends)
	}

	resolved := profile
	// @AX:NOTE [AUTO] @AX:REASON: design choice — R4/S6 extends merge format; base instructions prepended with framework header separator
	resolved.Instructions = base.Instructions + "\n\n## Framework: " + profile.Name + "\n\n" + profile.Instructions
	return resolved, nil
}

// SelectProfile selects the best matching profile by stack and framework.
// Priority: framework profile > language profile.
// @AX:NOTE [AUTO] @AX:REASON: design choice — R3 priority order (framework > language > none); changing order breaks profile assignment semantics
func SelectProfile(profiles []ProfileDefinition, stack, framework string) (ProfileDefinition, bool) {
	if stack == "" && framework == "" {
		return ProfileDefinition{}, false
	}

	// Try framework match first
	if framework != "" {
		for _, p := range profiles {
			if p.Framework == framework {
				return p, true
			}
		}
	}

	// Fall back to language (stack) match without framework
	if stack != "" {
		for _, p := range profiles {
			if p.Stack == stack && p.Framework == "" {
				return p, true
			}
		}
	}

	return ProfileDefinition{}, false
}

// SelectProfileWithConf selects a profile with a default name fallback.
// defaultName is the profiles.executor.default value from autopus.yaml.
func SelectProfileWithConf(profiles []ProfileDefinition, stack, framework, defaultName string) (ProfileDefinition, bool) {
	// Try normal selection first
	if stack != "" || framework != "" {
		if p, ok := SelectProfile(profiles, stack, framework); ok {
			return p, true
		}
	}

	// Fall back to default name
	if defaultName != "" {
		for _, p := range profiles {
			if p.Name == defaultName {
				return p, true
			}
		}
	}

	return ProfileDefinition{}, false
}

// ApplyCustomOverride performs a shallow merge of override onto base.
// Only non-zero override fields replace base fields. Instructions are never overridden.
func ApplyCustomOverride(base, override ProfileDefinition) ProfileDefinition {
	result := base
	if override.Linter != "" {
		result.Linter = override.Linter
	}
	if override.TestFramework != "" {
		result.TestFramework = override.TestFramework
	}
	if len(override.Tools) > 0 {
		result.Tools = override.Tools
	}
	if override.Source != "" {
		result.Source = override.Source
	}
	return result
}

// parseProfileData parses a profile from raw markdown bytes.
func parseProfileData(data []byte, filename, source string) (ProfileDefinition, string, error) {
	fm, body, err := splitFrontmatter(string(data))
	if err != nil {
		return ProfileDefinition{}, "", fmt.Errorf("frontmatter parse failed: %w", err)
	}

	var front profileFrontmatter
	if err := yaml.Unmarshal([]byte(fm), &front); err != nil {
		return ProfileDefinition{}, "", fmt.Errorf("YAML parse failed: %w", err)
	}

	// R11: skip profiles missing required fields
	if front.Name == "" || front.Stack == "" {
		warn := fmt.Sprintf("skipping %s: missing required field (name=%q, stack=%q)", filename, front.Name, front.Stack)
		return ProfileDefinition{}, warn, nil
	}

	profile := ProfileDefinition{
		Name:          front.Name,
		Stack:         front.Stack,
		Framework:     front.Framework,
		Extends:       front.Extends,
		Tools:         front.Tools,
		TestFramework: front.TestFramework,
		Linter:        front.Linter,
		Instructions:  strings.TrimSpace(body),
		Source:        source,
	}

	return profile, "", nil
}
