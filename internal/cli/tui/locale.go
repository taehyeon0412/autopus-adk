package tui

import (
	"os"
	"strings"
)

// DetectSystemLang returns the best-matching language code (en, ko, ja, zh)
// based on the system locale environment variables.
// Falls back to "en" if no supported locale is detected.
func DetectSystemLang() string {
	// Check locale env vars in priority order.
	for _, key := range []string{"LANG", "LC_ALL", "LC_MESSAGES", "LANGUAGE"} {
		if val := os.Getenv(key); val != "" {
			if lang := matchLang(val); lang != "" {
				return lang
			}
		}
	}
	return "en"
}

// matchLang extracts a supported language code from a locale string
// such as "ko_KR.UTF-8", "ja_JP", "zh_CN.UTF-8", or "en_US".
func matchLang(locale string) string {
	locale = strings.ToLower(locale)

	// Strip encoding suffix (e.g., ".utf-8").
	if idx := strings.IndexByte(locale, '.'); idx > 0 {
		locale = locale[:idx]
	}

	// Match prefix against supported languages.
	switch {
	case strings.HasPrefix(locale, "ko"):
		return "ko"
	case strings.HasPrefix(locale, "ja"):
		return "ja"
	case strings.HasPrefix(locale, "zh"):
		return "zh"
	case strings.HasPrefix(locale, "en"):
		return "en"
	default:
		return ""
	}
}
