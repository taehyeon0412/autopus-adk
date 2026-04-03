package routing

import (
	"strings"
	"unicode/utf8"
)

// ClassificationSignals captures the raw signals used for classification.
type ClassificationSignals struct {
	CharCount     int
	HasCodeBlocks bool
	Keywords      []string
	Score         int
}

// MessageComplexityClassifier classifies messages by complexity level.
type MessageComplexityClassifier struct {
	thresholds ClassifierThresholds
}

// NewClassifier creates a classifier with the given thresholds.
func NewClassifier(thresholds ClassifierThresholds) *MessageComplexityClassifier {
	return &MessageComplexityClassifier{thresholds: thresholds}
}

var (
	simpleKeywords  = []string{"확인", "상태", "목록", "check", "status", "list"}
	mediumKeywords  = []string{"수정", "추가", "변경", "fix", "add", "change"}
	complexKeywords = []string{"리팩토링", "분석", "아키텍처", "설계", "refactor", "analyze", "architecture", "design"}
)

// Classify determines the complexity of a message and returns the level with signals.
func (c *MessageComplexityClassifier) Classify(message string) (Complexity, ClassificationSignals) {
	signals := ClassificationSignals{
		CharCount:     utf8.RuneCountInString(message),
		HasCodeBlocks: strings.Contains(message, "```"),
	}

	// Character count score.
	switch {
	case signals.CharCount < c.thresholds.SimpleMaxChars:
		signals.Score--
	case signals.CharCount > c.thresholds.ComplexMinChars:
		signals.Score++
	}

	// Code blocks score.
	if signals.HasCodeBlocks {
		signals.Score++
	}

	// Keyword scoring.
	lower := strings.ToLower(message)
	for _, kw := range simpleKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			signals.Keywords = append(signals.Keywords, kw)
			signals.Score--
		}
	}
	for _, kw := range mediumKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			signals.Keywords = append(signals.Keywords, kw)
			// Medium keywords: neutral (score += 0).
		}
	}
	for _, kw := range complexKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			signals.Keywords = append(signals.Keywords, kw)
			signals.Score++
		}
	}

	// Determine complexity from final score.
	var level Complexity
	switch {
	case signals.Score < 0:
		level = ComplexitySimple
	case signals.Score > 0:
		level = ComplexityComplex
	default:
		level = ComplexityMedium
	}

	return level, signals
}
