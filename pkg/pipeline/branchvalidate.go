package pipeline

import (
	"fmt"
	"regexp"
	"strings"
)

// @AX:NOTE [AUTO]: regex defines the branch name character allow-list — changes here affect git injection protection
var validBranchRegex = regexp.MustCompile(`^[a-zA-Z0-9/_.-]+$`)

// ValidateBranchName validates that a branch name contains only safe characters.
// Empty string is allowed (detach mode).
func ValidateBranchName(name string) error {
	if name == "" {
		return nil // empty = detach mode
	}
	if len(name) > 255 {
		return fmt.Errorf("invalid branch name: exceeds 255 characters (%d)", len(name))
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("invalid branch name: must not start with '-'")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("invalid branch name: must not contain '..'")
	}
	if !validBranchRegex.MatchString(name) {
		return fmt.Errorf("invalid branch name %q: contains characters outside [a-zA-Z0-9/_.-]", name)
	}
	return nil
}
