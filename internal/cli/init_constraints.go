package cli

import (
	"io"

	"github.com/insajin/autopus-adk/internal/cli/tui"
	"github.com/insajin/autopus-adk/pkg/constraint"
)

// generateDefaultConstraints creates a default constraints.yaml for the project.
// It detects the project language and writes language-specific deny patterns.
// The file is written to .autopus/context/constraints.yaml relative to dir.
// If the file already exists, this is a no-op.
func generateDefaultConstraints(dir string, out io.Writer) error {
	path, err := constraint.GenerateDefaultFile(dir)
	if err != nil {
		return err
	}
	tui.Bullet(out, "constraints: "+path)
	return nil
}
