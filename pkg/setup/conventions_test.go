package setup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeConventions_GoProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "go.mod", "module example\n\ngo 1.23\n")
	writeFile(t, dir, "cmd/app/main.go", `package main

import (
	"fmt"
	"os"

	"example/pkg/util"
)

func main() {
	if err := util.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
`)
	writeFile(t, dir, "pkg/util/run_helper.go", `package util

import "fmt"

func Run() error {
	if err := doSomething(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

func doSomething() error { return nil }
`)
	writeFile(t, dir, "pkg/util/string_utils.go", "package util\n")
	writeFile(t, dir, ".golangci.yml", "linters:\n  enable:\n    - govet\n")

	conventions := AnalyzeConventions(dir, []Language{{Name: "Go"}})

	require.Contains(t, conventions, "Go")
	sample := conventions["Go"]

	assert.Equal(t, "snake_case", sample.FileNaming)
	assert.True(t, sample.HasLinter)
	assert.Equal(t, "golangci-lint", sample.LinterName)
	assert.True(t, sample.HasFormatter)
	assert.Equal(t, "gofmt", sample.FormatterName)
	assert.NotEmpty(t, sample.ExampleFiles)
}

func TestAnalyzeConventions_GoErrorPatterns(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "handler.go", `package handler

import "fmt"

func Handle() error {
	if err := process(); err != nil {
		return fmt.Errorf("handle: %w", err)
	}
	return nil
}

func process() error { return nil }
`)

	conventions := AnalyzeConventions(dir, []Language{{Name: "Go"}})
	sample := conventions["Go"]

	assert.NotEmpty(t, sample.ErrorPatterns)
	hasWrap := false
	for _, p := range sample.ErrorPatterns {
		if p == "fmt.Errorf with %w wrapping" {
			hasWrap = true
		}
	}
	assert.True(t, hasWrap, "should detect fmt.Errorf with %%w wrapping")
}

func TestAnalyzeConventions_TSProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "user-service.ts", "export class UserService {}\n")
	writeFile(t, dir, "api-handler.ts", "export function handle() {}\n")
	writeFile(t, dir, ".eslintrc.json", "{}\n")
	writeFile(t, dir, ".prettierrc", "{}\n")

	conventions := AnalyzeConventions(dir, []Language{{Name: "TypeScript"}})

	require.Contains(t, conventions, "TypeScript")
	sample := conventions["TypeScript"]

	assert.Equal(t, "kebab-case", sample.FileNaming)
	assert.True(t, sample.HasLinter)
	assert.Equal(t, "ESLint", sample.LinterName)
	assert.True(t, sample.HasFormatter)
	assert.Equal(t, "Prettier", sample.FormatterName)
}

func TestAnalyzeConventions_PythonProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "main_app.py", "def main(): pass\n")
	writeFile(t, dir, "data_loader.py", "class DataLoader: pass\n")
	writeFile(t, dir, "pyproject.toml", "[tool.ruff]\nline-length = 88\n\n[tool.black]\nline-length = 88\n")

	conventions := AnalyzeConventions(dir, []Language{{Name: "Python"}})

	require.Contains(t, conventions, "Python")
	sample := conventions["Python"]

	assert.Equal(t, "snake_case", sample.FileNaming)
	assert.True(t, sample.HasLinter)
	assert.Equal(t, "Ruff", sample.LinterName)
	assert.True(t, sample.HasFormatter)
	assert.Equal(t, "Black", sample.FormatterName)
}

func TestAnalyzeConventions_RustProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	writeFile(t, dir, "main_module.rs", "fn main() {}\n")
	writeFile(t, dir, "rustfmt.toml", "max_width = 100\n")

	conventions := AnalyzeConventions(dir, []Language{{Name: "Rust"}})

	require.Contains(t, conventions, "Rust")
	sample := conventions["Rust"]

	assert.True(t, sample.HasLinter)
	assert.Equal(t, "clippy", sample.LinterName)
	assert.True(t, sample.HasFormatter)
	assert.Contains(t, sample.FormatterName, "rustfmt")
}

func TestDetectFileNaming_SnakeCase(t *testing.T) {
	t.Parallel()
	files := []string{"my_file.go", "another_file.go", "third_file.go"}
	assert.Equal(t, "snake_case", detectFileNaming(files))
}

func TestDetectFileNaming_KebabCase(t *testing.T) {
	t.Parallel()
	files := []string{"my-component.ts", "api-handler.ts", "data-service.ts"}
	assert.Equal(t, "kebab-case", detectFileNaming(files))
}

func TestDetectFileNaming_PascalCase(t *testing.T) {
	t.Parallel()
	files := []string{"MyComponent.tsx", "ApiHandler.tsx", "DataService.tsx"}
	assert.Equal(t, "PascalCase", detectFileNaming(files))
}

func TestDetectFileNaming_CamelCase(t *testing.T) {
	t.Parallel()
	files := []string{"myComponent.js", "apiHandler.js", "dataService.js"}
	assert.Equal(t, "camelCase", detectFileNaming(files))
}

func TestRender_ConventionsWithDetectedPatterns(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name:      "test",
		Languages: []Language{{Name: "Go"}},
		Conventions: map[string]ConventionSample{
			"Go": {
				FileNaming:    "snake_case",
				ErrorPatterns: []string{"fmt.Errorf with %w wrapping", "if err != nil guard"},
				ImportStyle:   "grouped (stdlib / internal / external)",
				HasLinter:     true,
				LinterName:    "golangci-lint",
				HasFormatter:  true,
				FormatterName: "gofmt",
				ExampleFiles:  []string{"pkg/setup/scanner.go"},
			},
		},
	}

	ds := Render(info, nil)

	// Should include detected patterns, not just generic ones
	assert.Contains(t, ds.Conventions, "snake_case")
	assert.Contains(t, ds.Conventions, "fmt.Errorf with %w wrapping")
	assert.Contains(t, ds.Conventions, "grouped")
	assert.Contains(t, ds.Conventions, "golangci-lint")
	assert.Contains(t, ds.Conventions, "gofmt")
}

func TestRender_ConventionsWithoutDetectedPatterns(t *testing.T) {
	t.Parallel()
	info := &ProjectInfo{
		Name:      "test",
		Languages: []Language{{Name: "Go"}},
	}

	ds := Render(info, nil)

	// Should fall back to generic patterns
	assert.Contains(t, ds.Conventions, "Go")
	assert.Contains(t, ds.Conventions, "PascalCase")
	assert.Contains(t, ds.Conventions, "camelCase")
}
