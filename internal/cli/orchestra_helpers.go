package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// buildFileContents reads each file and returns formatted content string.
func buildFileContents(files []string) string {
	var sb strings.Builder
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(&sb, "--- %s (읽기 실패: %v) ---\n\n", f, err)
			continue
		}
		fmt.Fprintf(&sb, "--- %s ---\n```\n%s\n```\n\n", f, string(content))
	}
	return sb.String()
}

// buildReviewPrompt builds the review prompt, including file contents if provided.
func buildReviewPrompt(files []string) string {
	if len(files) == 0 {
		return "현재 프로젝트의 코드를 리뷰해주세요. 품질, 가독성, 잠재적 버그를 중심으로 분석하세요."
	}
	var sb strings.Builder
	sb.WriteString("다음 파일들을 코드 리뷰해주세요:\n\n")
	sb.WriteString(buildFileContents(files))
	sb.WriteString("품질, 가독성, 잠재적 버그를 중심으로 분석하세요.")
	return sb.String()
}

// buildSecurePrompt builds the security analysis prompt, including file contents if provided.
func buildSecurePrompt(files []string) string {
	if len(files) == 0 {
		return "현재 프로젝트의 보안 취약점을 분석해주세요. OWASP Top 10을 기준으로 검토하세요."
	}
	var sb strings.Builder
	sb.WriteString("다음 파일들의 보안 취약점을 분석해주세요:\n\n")
	sb.WriteString(buildFileContents(files))
	sb.WriteString("OWASP Top 10을 기준으로 검토하세요.")
	return sb.String()
}

// flagStringIfChanged returns the flag value only if the flag was explicitly set.
// Returns empty string when using default (not changed).
func flagStringIfChanged(cmd *cobra.Command, name, value string) string {
	if cmd.Flags().Changed(name) {
		return value
	}
	return ""
}

// flagStringSliceIfChanged returns the flag value only if the flag was explicitly set.
// Returns nil when using default (not changed).
func flagStringSliceIfChanged(cmd *cobra.Command, name string, value []string) []string {
	if cmd.Flags().Changed(name) {
		return value
	}
	return nil
}

// isStdoutTTY returns true if stdout is a terminal device.
// @AX:NOTE [AUTO] REQ-1 TTY detection — used by auto-detach decision; returns false in CI/pipe contexts
func isStdoutTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
