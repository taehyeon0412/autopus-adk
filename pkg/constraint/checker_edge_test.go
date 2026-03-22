package constraint_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/insajin/autopus-adk/pkg/constraint"
)

// TestCheck_CategoryFilterNoMatch verifies that category filter with no matching constraints returns nil.
func TestCheck_CategoryFilterNoMatch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTempFile(t, dir, "app.go", "package main\n\nfmt.Println(\"x\")\n")

	// CategoryTesting not present in sampleConstraints, so filterByCategory returns empty.
	opts := constraint.CheckOptions{
		Categories: []constraint.Category{constraint.CategoryTesting},
	}
	violations, err := constraint.Check(dir, sampleConstraints(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(violations))
	}
}

// TestCheck_SkipsNodeModulesDir verifies that node_modules directory is skipped.
func TestCheck_SkipsNodeModulesDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	nmDir := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(nmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTempFile(t, nmDir, "lib.go", "package lib\nfmt.Println(\"nm\")\n")

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (node_modules skipped), got %d", len(violations))
	}
}

// TestCheck_SkipsTestdataDir verifies that testdata directory is skipped.
func TestCheck_SkipsTestdataDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tdDir := filepath.Join(dir, "testdata")
	if err := os.MkdirAll(tdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTempFile(t, tdDir, "fixture.go", "package fixture\nfmt.Println(\"td\")\n")

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (testdata skipped), got %d", len(violations))
	}
}

// TestCheck_MultipleExtensions verifies extension filter allows multiple extensions.
func TestCheck_MultipleExtensions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTempFile(t, dir, "app.go", "package main\n\nfmt.Println(\"go\")\n")
	writeTempFile(t, dir, "app.ts", "// ts\nfmt.Println('ts')\n")
	writeTempFile(t, dir, "app.py", "# py\nfmt.Println('py')\n")

	opts := constraint.CheckOptions{
		Extensions: []string{".go", ".ts"},
	}
	violations, err := constraint.Check(dir, sampleConstraints(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// .go and .ts files both contain fmt.Println; .py is excluded.
	if len(violations) != 2 {
		t.Errorf("expected 2 violations, got %d", len(violations))
	}
}

// TestCheck_InaccessibleSubdir verifies that a sub-directory without execute permission is skipped.
func TestCheck_InaccessibleSubdir(t *testing.T) {
	t.Parallel()
	if os.Getuid() == 0 {
		t.Skip("running as root; permission check not enforced")
	}

	dir := t.TempDir()
	// Create a readable file in the top-level directory.
	writeTempFile(t, dir, "top.go", "package main\n\nfmt.Println(\"top\")\n")

	// Create a subdirectory with no execute permission so Walk triggers err != nil.
	subDir := filepath.Join(dir, "locked")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTempFile(t, subDir, "sub.go", "package sub\nfmt.Println(\"sub\")\n")
	if err := os.Chmod(subDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(subDir, 0o755) })

	// Walk should skip the inaccessible entry and still return violations from the top-level file.
	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// top.go contains fmt.Println — at least 1 violation expected.
	if len(violations) < 1 {
		t.Errorf("expected at least 1 violation from top.go, got %d", len(violations))
	}
}

// TestCheck_UnreadableFile verifies that a file that becomes unreadable during walk is skipped.
func TestCheck_UnreadableFile(t *testing.T) {
	t.Parallel()
	if os.Getuid() == 0 {
		t.Skip("running as root; permission check not enforced")
	}

	dir := t.TempDir()
	// Create two files; one will be made unreadable.
	writeTempFile(t, dir, "readable.go", "package main\n\nfmt.Println(\"hi\")\n")
	unreadable := writeTempFile(t, dir, "unreadable.go", "package main\n\nfmt.Println(\"locked\")\n")
	if err := os.Chmod(unreadable, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o644) })

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// readable.go contributes 1 violation; unreadable.go is silently skipped.
	if len(violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(violations))
	}
}

// TestFormatViolations_NoViolations checks the "no violations" message.
func TestFormatViolations_NoViolations(t *testing.T) {
	result := constraint.FormatViolations(nil)
	if result != "No constraint violations found." {
		t.Errorf("unexpected message: %q", result)
	}
}

// TestFormatViolations_WithViolations checks that the report contains key fields.
func TestFormatViolations_WithViolations(t *testing.T) {
	violations := []constraint.Violation{
		{
			Constraint: constraint.Constraint{
				Pattern:  "fmt.Println",
				Reason:   "use structured logging",
				Suggest:  "slog.Info",
				Category: constraint.CategoryConvention,
			},
			File:  "main.go",
			Line:  5,
			Match: "fmt.Println(\"hello\")",
		},
	}

	report := constraint.FormatViolations(violations)

	checks := []string{"1 constraint violation", "main.go:5", "fmt.Println", "use structured logging", "slog.Info"}
	for _, s := range checks {
		if !strings.Contains(report, s) {
			t.Errorf("report missing expected substring %q", s)
		}
	}
}

// TestFormatViolations_MultipleViolations checks formatting of multiple violations.
func TestFormatViolations_MultipleViolations(t *testing.T) {
	t.Parallel()
	violations := []constraint.Violation{
		{
			Constraint: constraint.Constraint{
				Pattern:  "fmt.Println",
				Reason:   "use logging",
				Suggest:  "slog.Info",
				Category: constraint.CategoryConvention,
			},
			File:  "a.go",
			Line:  1,
			Match: "fmt.Println(\"x\")",
		},
		{
			Constraint: constraint.Constraint{
				Pattern:  "os.Exit",
				Reason:   "no cleanup",
				Suggest:  "return error",
				Category: constraint.CategorySecurity,
			},
			File:  "b.go",
			Line:  10,
			Match: "os.Exit(1)",
		},
	}

	report := constraint.FormatViolations(violations)

	for _, want := range []string{"2 constraint violation", "a.go:1", "b.go:10", "1.", "2."} {
		if !strings.Contains(report, want) {
			t.Errorf("report missing %q", want)
		}
	}
}
