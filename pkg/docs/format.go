package docs

import (
	"errors"
	"fmt"
	"strings"
)

// sourceLabel maps internal source identifiers to display labels.
func sourceLabel(source string) string {
	switch source {
	case "context7":
		return "Context7"
	case "scraper":
		return "WebSearch"
	case "cache":
		return "Cache"
	default:
		if len(source) == 0 {
			return source
		}
		return strings.ToUpper(source[:1]) + source[1:]
	}
}

// FormatPromptInjection formats documentation results as a prompt injection section.
// Returns an error if results is nil or empty.
func FormatPromptInjection(results []*DocResult) (string, error) {
	if len(results) == 0 {
		return "", errors.New("no documentation results to format")
	}

	var sb strings.Builder
	sb.WriteString("## Reference Documentation\n\n")
	sb.WriteString("The following documentation was fetched from Context7 for libraries used in this task.\n")

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("\n### %s (via %s)\n%s\n", r.LibraryName, sourceLabel(r.Source), r.Content))
	}

	return sb.String(), nil
}
