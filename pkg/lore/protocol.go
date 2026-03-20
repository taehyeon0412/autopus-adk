package lore

import (
	"fmt"
	"strings"
)

// trailerKeys는 트레일러 키와 LoreEntry 필드 매핑이다.
var trailerKeys = []struct {
	key   string
	field string
}{
	{"Constraint", "Constraint"},
	{"Rejected", "Rejected"},
	{"Confidence", "Confidence"},
	{"Scope-risk", "ScopeRisk"},
	{"Reversibility", "Reversibility"},
	{"Directive", "Directive"},
	{"Tested", "Tested"},
	{"Not-tested", "NotTested"},
	{"Related", "Related"},
}

// ParseTrailers는 git commit 메시지에서 Lore 트레일러를 파싱한다.
func ParseTrailers(commitMsg string) (*LoreEntry, error) {
	entry := &LoreEntry{}

	lines := strings.Split(commitMsg, "\n")
	for _, line := range lines {
		for _, t := range trailerKeys {
			prefix := t.key + ": "
			if strings.HasPrefix(line, prefix) {
				value := strings.TrimPrefix(line, prefix)
				value = strings.TrimSpace(value)
				setField(entry, t.field, value)
				break
			}
		}
	}

	return entry, nil
}

// FormatTrailers는 LoreEntry를 트레일러 형식 문자열로 변환한다.
func FormatTrailers(entry *LoreEntry) string {
	var parts []string

	appendIfNotEmpty := func(key, value string) {
		if value != "" {
			parts = append(parts, fmt.Sprintf("%s: %s", key, value))
		}
	}

	appendIfNotEmpty("Constraint", entry.Constraint)
	appendIfNotEmpty("Rejected", entry.Rejected)
	appendIfNotEmpty("Confidence", entry.Confidence)
	appendIfNotEmpty("Scope-risk", entry.ScopeRisk)
	appendIfNotEmpty("Reversibility", entry.Reversibility)
	appendIfNotEmpty("Directive", entry.Directive)
	appendIfNotEmpty("Tested", entry.Tested)
	appendIfNotEmpty("Not-tested", entry.NotTested)
	appendIfNotEmpty("Related", entry.Related)

	return strings.Join(parts, "\n")
}

// setField는 필드명에 해당하는 LoreEntry 필드를 설정한다.
func setField(entry *LoreEntry, field, value string) {
	switch field {
	case "Constraint":
		entry.Constraint = value
	case "Rejected":
		entry.Rejected = value
	case "Confidence":
		entry.Confidence = value
	case "ScopeRisk":
		entry.ScopeRisk = value
	case "Reversibility":
		entry.Reversibility = value
	case "Directive":
		entry.Directive = value
	case "Tested":
		entry.Tested = value
	case "NotTested":
		entry.NotTested = value
	case "Related":
		entry.Related = value
	}
}
