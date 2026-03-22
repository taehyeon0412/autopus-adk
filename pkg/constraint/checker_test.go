package constraint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/insajin/autopus-adk/pkg/constraint"
)

// TestCheck_NoConstraints verifies early return when constraint list is empty.
func TestCheck_NoConstraints(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "main.go", "package main\nfunc main() {}\n")

	violations, err := constraint.Check(dir, nil, constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(violations))
	}
}

// TestCheck_DetectsViolations verifies that matching patterns are reported.
func TestCheck_DetectsViolations(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "app.go", "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n")

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.Constraint.Pattern != "fmt.Println" {
		t.Errorf("expected pattern fmt.Println, got %s", v.Constraint.Pattern)
	}
	if v.Line != 4 {
		t.Errorf("expected line 4, got %d", v.Line)
	}
	if v.Match == "" {
		t.Error("match text should not be empty")
	}
}

// TestCheck_MultipleViolationsInSingleFile checks all matching lines are reported.
func TestCheck_MultipleViolationsInSingleFile(t *testing.T) {
	dir := t.TempDir()
	src := "package main\n\nfunc a() {\n\tfmt.Println(\"a\")\n\tos.Exit(1)\n}\n"
	writeTempFile(t, dir, "multi.go", src)

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}
}

// TestCheck_CategoryFilter verifies that category filtering limits results.
func TestCheck_CategoryFilter(t *testing.T) {
	dir := t.TempDir()
	src := "package main\n\nfunc main() {\n\tfmt.Println(\"x\")\n\tos.Exit(1)\n}\n"
	writeTempFile(t, dir, "filtered.go", src)

	opts := constraint.CheckOptions{
		Categories: []constraint.Category{constraint.CategorySecurity},
	}
	violations, err := constraint.Check(dir, sampleConstraints(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only os.Exit (security) should match, not fmt.Println (convention)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Constraint.Category != constraint.CategorySecurity {
		t.Errorf("expected security category, got %s", violations[0].Constraint.Category)
	}
}

// TestCheck_ExtensionFilter verifies that non-matching extensions are skipped.
func TestCheck_ExtensionFilter(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "script.sh", "#!/bin/sh\necho fmt.Println\n")
	writeTempFile(t, dir, "app.go", "package main\n\nfunc main() {\n\tfmt.Println(\"y\")\n}\n")

	opts := constraint.CheckOptions{
		Extensions: []string{".sh"},
	}
	violations, err := constraint.Check(dir, sampleConstraints(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only .sh file scanned, which contains the literal string "fmt.Println"
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation from .sh file, got %d", len(violations))
	}
}

// TestCheck_SkipsHiddenDirs verifies that .git and similar directories are ignored.
func TestCheck_SkipsHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	hiddenDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(hiddenDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTempFile(t, hiddenDir, "config.go", "package git\nfmt.Println(\"internal\")\n")

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (hidden dir skipped), got %d", len(violations))
	}
}

// TestCheck_SkipsVendorDir verifies that vendor directory is skipped.
func TestCheck_SkipsVendorDir(t *testing.T) {
	dir := t.TempDir()
	vendorDir := filepath.Join(dir, "vendor")
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTempFile(t, vendorDir, "dep.go", "package dep\nfmt.Println(\"vendor\")\n")

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (vendor dir skipped), got %d", len(violations))
	}
}

// TestCheck_ViolationFieldsPopulated ensures all Violation fields are set correctly.
func TestCheck_ViolationFieldsPopulated(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "sample.go", "package main\n\nfunc f() {\n\tos.Exit(2)\n}\n")

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.File == "" {
		t.Error("File field must not be empty")
	}
	if v.Line == 0 {
		t.Error("Line field must not be zero")
	}
	if v.Match == "" {
		t.Error("Match field must not be empty")
	}
	if v.Constraint.Suggest == "" {
		t.Error("Constraint.Suggest must not be empty")
	}
}

// TestCheck_EmptyDirectory verifies no error and no violations on an empty dir.
func TestCheck_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(violations))
	}
}

// TestCheck_NoViolationsWhenPatternAbsent verifies clean code returns no violations.
func TestCheck_NoViolationsWhenPatternAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTempFile(t, dir, "clean.go", "package main\n\nfunc main() {\n\tlog.Info(\"hello\")\n}\n")

	violations, err := constraint.Check(dir, sampleConstraints(), constraint.CheckOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(violations))
	}
}
