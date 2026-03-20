package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/arch"
)

// newArchCmd는 arch 서브커맨드를 생성한다.
func newArchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "arch",
		Short: "Architecture analysis and enforcement",
	}

	cmd.AddCommand(newArchGenerateCmd())
	cmd.AddCommand(newArchEnforceCmd())
	return cmd
}

func newArchGenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [dir]",
		Short: "Generate ARCHITECTURE.md from project structure",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			archMap, err := arch.Analyze(dir)
			if err != nil {
				return fmt.Errorf("아키텍처 분석 실패: %w", err)
			}

			content, err := arch.Generate(archMap)
			if err != nil {
				return fmt.Errorf("ARCHITECTURE.md 생성 실패: %w", err)
			}

			if err := os.WriteFile("ARCHITECTURE.md", []byte(content), 0o644); err != nil {
				return fmt.Errorf("파일 저장 실패: %w", err)
			}

			fmt.Println("ARCHITECTURE.md 생성 완료")
			return nil
		},
	}
}

func newArchEnforceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enforce [dir]",
		Short: "Enforce architecture rules and report violations",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			archMap, err := arch.Analyze(dir)
			if err != nil {
				return fmt.Errorf("아키텍처 분석 실패: %w", err)
			}

			// 기본 규칙
			rules := []arch.LintRule{
				{
					Name:        "no-pkg-to-internal",
					FromLayer:   "pkg",
					ToLayer:     "internal",
					Allowed:     false,
					Remediation: "pkg 레이어에서 internal로의 직접 의존을 피하고 인터페이스를 정의하세요",
				},
			}

			violations := arch.Lint(archMap, rules)
			if len(violations) == 0 {
				fmt.Println("아키텍처 규칙 위반 없음")
				return nil
			}

			fmt.Fprintf(os.Stderr, "아키텍처 규칙 위반 %d건:\n", len(violations))
			for _, v := range violations {
				fmt.Fprintf(os.Stderr, "  [%s] %s -> %s\n    %s\n    수정: %s\n",
					v.Rule, v.From, v.To, v.Message, v.Remediation)
			}
			return fmt.Errorf("%d개의 아키텍처 규칙 위반", len(violations))
		},
	}
}
