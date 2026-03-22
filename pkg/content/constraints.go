package content

import (
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/constraint"
)

// GenerateConstraintInstruction produces the anti-pattern constraint section
// for agent prompt injection. Returns an empty string when constraints are
// disabled, the file is missing, or no patterns are defined.
func GenerateConstraintInstruction(projectDir string, cfg config.ConstraintConf) string {
	if !cfg.Enabled {
		return ""
	}

	path := cfg.Path
	if path == "" {
		path = constraint.DefaultPath
	}

	// Use project-relative path if not absolute.
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, path)
	}

	reg, err := constraint.Load(path)
	if err != nil || reg.IsEmpty() {
		return ""
	}

	return reg.GeneratePromptText()
}
