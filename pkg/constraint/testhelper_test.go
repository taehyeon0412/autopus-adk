package constraint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/insajin/autopus-adk/pkg/constraint"
)

// sampleConstraints returns a minimal set of constraints for tests.
func sampleConstraints() []constraint.Constraint {
	return []constraint.Constraint{
		{
			Pattern:  "fmt.Println",
			Reason:   "use structured logging instead",
			Suggest:  "slog.Info or zap.Info",
			Category: constraint.CategoryConvention,
		},
		{
			Pattern:  "os.Exit",
			Reason:   "hard exits prevent cleanup",
			Suggest:  "return error and let main handle",
			Category: constraint.CategorySecurity,
		},
		{
			Pattern:  "time.Sleep",
			Reason:   "busy-wait degrades throughput",
			Suggest:  "use channels or ticker",
			Category: constraint.CategoryPerformance,
		},
	}
}

// writeTempFile creates a temporary file with the given content and returns its path.
func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return path
}
