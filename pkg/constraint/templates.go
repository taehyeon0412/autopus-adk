package constraint

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Language represents a detected project language.
type Language string

const (
	LangGo         Language = "go"
	LangTypeScript Language = "typescript"
	LangPython     Language = "python"
)

// DetectLanguage determines the primary language of a project directory.
// It checks for language-specific marker files in the given directory.
func DetectLanguage(dir string) Language {
	if fileExists(filepath.Join(dir, "go.mod")) {
		return LangGo
	}
	if fileExists(filepath.Join(dir, "package.json")) {
		return LangTypeScript
	}
	if fileExists(filepath.Join(dir, "pyproject.toml")) || fileExists(filepath.Join(dir, "requirements.txt")) {
		return LangPython
	}
	return LangGo // default fallback
}

// DefaultConstraints returns language-specific default deny patterns.
func DefaultConstraints(lang Language) []Constraint {
	switch lang {
	case LangTypeScript:
		return defaultTypeScriptConstraints()
	case LangPython:
		return defaultPythonConstraints()
	default:
		return defaultGoConstraints()
	}
}

// GenerateDefaultFile creates a constraints.yaml with default patterns for the detected language.
// Returns the path of the created file. Does nothing if the file already exists.
func GenerateDefaultFile(projectDir string) (string, error) {
	path := filepath.Join(projectDir, DefaultPath)

	if fileExists(path) {
		return path, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	lang := DetectLanguage(projectDir)
	cf := ConstraintsFile{Deny: DefaultConstraints(lang)}

	data, err := yaml.Marshal(&cf)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func defaultGoConstraints() []Constraint {
	return []Constraint{
		{
			Pattern:  "context.Background()",
			Reason:   "always propagate parent context",
			Suggest:  "add ctx context.Context parameter",
			Category: CategoryConvention,
		},
		{
			Pattern:  "context.TODO()",
			Reason:   "TODO context is temporary",
			Suggest:  "pass proper context",
			Category: CategoryConvention,
		},
		{
			Pattern:  `fmt.Sprintf("SELECT`,
			Reason:   "SQL injection risk",
			Suggest:  "use parameterized queries: db.Query(query, args...)",
			Category: CategorySecurity,
		},
		{
			Pattern:  "time.Sleep(",
			Reason:   "slows tests, non-deterministic",
			Suggest:  "use ticker, timer, or channel",
			Category: CategoryTesting,
		},
		{
			Pattern:  "os.Exit(",
			Reason:   "prevents defer execution and testing",
			Suggest:  "return error to caller",
			Category: CategoryConvention,
		},
	}
}

func defaultTypeScriptConstraints() []Constraint {
	return []Constraint{
		{
			Pattern:  "any",
			Reason:   "defeats type safety",
			Suggest:  "use specific types or unknown",
			Category: CategoryConvention,
		},
		{
			Pattern:  "eval(",
			Reason:   "code injection risk",
			Suggest:  "use safe alternatives",
			Category: CategorySecurity,
		},
		{
			Pattern:  "innerHTML",
			Reason:   "XSS vulnerability",
			Suggest:  "use textContent or sanitized rendering",
			Category: CategorySecurity,
		},
		{
			Pattern:  "console.log(",
			Reason:   "debug output in production",
			Suggest:  "use a proper logger",
			Category: CategoryConvention,
		},
	}
}

func defaultPythonConstraints() []Constraint {
	return []Constraint{
		{
			Pattern:  "eval(",
			Reason:   "code injection risk",
			Suggest:  "use ast.literal_eval or safe parsing",
			Category: CategorySecurity,
		},
		{
			Pattern:  "exec(",
			Reason:   "arbitrary code execution",
			Suggest:  "use subprocess with strict args",
			Category: CategorySecurity,
		},
		{
			Pattern:  "import *",
			Reason:   "pollutes namespace",
			Suggest:  "import specific names",
			Category: CategoryConvention,
		},
		{
			Pattern:  "pickle.load",
			Reason:   "deserialization attack risk",
			Suggest:  "use JSON or safe format",
			Category: CategorySecurity,
		},
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
