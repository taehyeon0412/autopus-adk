// Package constraint provides anti-pattern registry for project-specific deny patterns.
package constraint

// Category classifies constraint patterns.
type Category string

const (
	CategorySecurity    Category = "security"
	CategoryPerformance Category = "performance"
	CategoryConvention  Category = "convention"
	CategoryTesting     Category = "testing"
)

// Constraint represents a single deny pattern.
type Constraint struct {
	Pattern  string   `yaml:"pattern"`
	Reason   string   `yaml:"reason"`
	Suggest  string   `yaml:"suggest"`
	Category Category `yaml:"category"`
}

// ConstraintsFile is the top-level structure of constraints.yaml.
type ConstraintsFile struct {
	Deny []Constraint `yaml:"deny"`
}

// Violation represents a detected pattern violation.
type Violation struct {
	Constraint Constraint
	File       string
	Line       int
	Match      string
}
