// Package cli는 init/update 시 인터랙티브 프롬프트를 제공한다.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/detect"
)

// promptYesNo는 유저에게 yes/no 질문을 하고 결과를 반환한다.
func promptYesNo(out io.Writer, question string, defaultNo bool) bool {
	hint := "(y/N)"
	if !defaultNo {
		hint = "(Y/n)"
	}
	fmt.Fprintf(out, "  %s %s: ", question, hint)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if defaultNo {
		return answer == "y" || answer == "yes"
	}
	return answer != "n" && answer != "no"
}

// promptChoice는 유저에게 번호 선택을 요청하고 결과를 반환한다.
func promptChoice(out io.Writer, question string, options []string, defaultIdx int) int {
	fmt.Fprintf(out, "\n  %s\n", question)
	for i, opt := range options {
		marker := "  "
		if i == defaultIdx {
			marker = "* "
		}
		fmt.Fprintf(out, "    %s%d) %s\n", marker, i+1, opt)
	}
	fmt.Fprintf(out, "  Choose [%d]: ", defaultIdx+1)

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)

	if answer == "" {
		return defaultIdx
	}

	for i := range options {
		if answer == fmt.Sprintf("%d", i+1) {
			return i
		}
	}
	return defaultIdx
}

// warnParentRuleConflicts는 부모 디렉터리에 다른 하네스 규칙이 있으면 경고하고,
// 유저에게 부모 규칙 무시 여부를 선택하게 한다.
func warnParentRuleConflicts(cmd *cobra.Command, dir string, cfg *config.HarnessConfig) {
	conflicts := detect.CheckParentRuleConflicts(dir)
	if len(conflicts) == 0 {
		return
	}

	out := cmd.OutOrStdout()

	// 이미 격리 설정이 되어 있으면 알림만 출력
	if cfg.IsolateRules {
		fmt.Fprintln(out, "\n  Parent rules detected (isolated via isolate_rules: true):")
		for _, c := range conflicts {
			fmt.Fprintf(out, "    - %s/.claude/rules/%s/ (ignored)\n", c.ParentDir, c.Namespace)
		}
		return
	}

	// 충돌 경고 출력
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "  Parent rule conflicts detected:")
	for _, c := range conflicts {
		fmt.Fprintf(out, "    - %s/.claude/rules/%s/\n", c.ParentDir, c.Namespace)
	}
	fmt.Fprintln(out, "  Claude Code inherits rules from parent directories.")
	fmt.Fprintln(out, "  These rules will apply alongside autopus rules in this project.")
	fmt.Fprintln(out, "")

	if promptYesNo(out, "Ignore parent rules?", true) {
		cfg.IsolateRules = true
		if err := config.Save(dir, cfg); err != nil {
			fmt.Fprintf(out, "  [ERROR] autopus.yaml save failed: %v\n", err)
			return
		}
		fmt.Fprintln(out, "  isolate_rules: true set in autopus.yaml")
	}
}

var langCodes = []string{"en", "ko", "ja", "zh"}
var langLabels = []string{
	"English",
	"Korean (한국어)",
	"Japanese (日本語)",
	"Chinese (中文)",
}

// promptLanguageSettings는 유저에게 프로젝트 언어 설정을 물어본다.
// 이미 설정되어 있으면 프롬프트를 건너뛴다.
func promptLanguageSettings(cmd *cobra.Command, dir string, cfg *config.HarnessConfig) {
	// 이미 설정되어 있으면 스킵
	if cfg.Language.Comments != "" && cfg.Language.Commits != "" && cfg.Language.AIResponses != "" {
		return
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "\n  Language Settings:")

	// 코드 주석 언어
	if cfg.Language.Comments == "" {
		idx := promptChoice(out, "Code comments language?", langLabels, 0)
		cfg.Language.Comments = langCodes[idx]
	}

	// 커밋 메시지 언어
	if cfg.Language.Commits == "" {
		idx := promptChoice(out, "Commit message language?", langLabels, 0)
		cfg.Language.Commits = langCodes[idx]
	}

	// AI 응답 언어
	if cfg.Language.AIResponses == "" {
		idx := promptChoice(out, "AI response language?", langLabels, 0)
		cfg.Language.AIResponses = langCodes[idx]
	}

	if err := config.Save(dir, cfg); err != nil {
		fmt.Fprintf(out, "  [ERROR] autopus.yaml save failed: %v\n", err)
		return
	}

	fmt.Fprintf(out, "\n  Language configured: comments=%s, commits=%s, ai=%s\n",
		cfg.Language.Comments, cfg.Language.Commits, cfg.Language.AIResponses)
}
