package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchLang(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{"Korean UTF-8", "ko_KR.UTF-8", "ko"},
		{"Korean bare", "ko_KR", "ko"},
		{"Korean short", "ko", "ko"},
		{"Japanese UTF-8", "ja_JP.UTF-8", "ja"},
		{"Chinese UTF-8", "zh_CN.UTF-8", "zh"},
		{"Chinese Taiwan", "zh_TW.UTF-8", "zh"},
		{"English US", "en_US.UTF-8", "en"},
		{"English bare", "en", "en"},
		{"French", "fr_FR.UTF-8", ""},
		{"Empty", "", ""},
		{"C locale", "C", ""},
		{"POSIX", "POSIX", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchLang(tt.locale)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Not parallel: uses t.Setenv.
func TestDetectSystemLang_Fallback(t *testing.T) {
	for _, key := range []string{"LANG", "LC_ALL", "LC_MESSAGES", "LANGUAGE"} {
		t.Setenv(key, "")
	}

	got := DetectSystemLang()
	assert.Equal(t, "en", got)
}

// Not parallel: uses t.Setenv.
func TestDetectSystemLang_Korean(t *testing.T) {
	t.Setenv("LANG", "ko_KR.UTF-8")
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANGUAGE", "")

	got := DetectSystemLang()
	assert.Equal(t, "ko", got)
}

// Not parallel: uses t.Setenv.
func TestDetectSystemLang_LCAllOverride(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "ja_JP.UTF-8")

	got := DetectSystemLang()
	assert.Equal(t, "en", got, "LANG has higher priority than LC_ALL")
}
