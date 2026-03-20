// Package cli는 doctor 커맨드를 구현한다.
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/adapter/claude"
	"github.com/insajin/autopus-adk/pkg/adapter/codex"
	"github.com/insajin/autopus-adk/pkg/adapter/gemini"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/detect"
)

func newDoctorCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check autopus harness health",
		Long:  "하네스 설치 상태를 진단합니다. 누락된 파일, 의존성 상태, 플랫폼 상태를 확인합니다.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dir == "" {
				var err error
				dir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("현재 디렉터리를 가져올 수 없음: %w", err)
				}
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Autopus-ADK Doctor")
			fmt.Fprintln(out, "==================")

			// 1. 설정 파일 확인
			cfg, err := config.Load(dir)
			if err != nil {
				fmt.Fprintf(out, "[ERROR] autopus.yaml 로드 실패: %v\n", err)
				return nil
			}
			fmt.Fprintf(out, "[OK] autopus.yaml (mode: %s)\n", cfg.Mode)

			// 2. 플랫폼 파일 검증
			ctx := context.Background()
			allOK := true
			for _, p := range cfg.Platforms {
				var validationErrs []adapter.ValidationError
				var validateErr error

				switch p {
				case "claude-code":
					a := claude.NewWithRoot(dir)
					validationErrs, validateErr = a.Validate(ctx)
				case "codex":
					a := codex.NewWithRoot(dir)
					validationErrs, validateErr = a.Validate(ctx)
				case "gemini-cli":
					a := gemini.NewWithRoot(dir)
					validationErrs, validateErr = a.Validate(ctx)
				default:
					fmt.Fprintf(out, "[WARN] 알 수 없는 플랫폼: %s\n", p)
					continue
				}

				if validateErr != nil {
					fmt.Fprintf(out, "[ERROR] %s 검증 실패: %v\n", p, validateErr)
					allOK = false
					continue
				}

				if len(validationErrs) == 0 {
					fmt.Fprintf(out, "[OK] %s\n", p)
				} else {
					for _, ve := range validationErrs {
						level := strings.ToUpper(ve.Level)
						fmt.Fprintf(out, "[%s] %s: %s\n", level, p, ve.Message)
						if ve.Level == "error" {
							allOK = false
						}
					}
				}
			}

			// 3. 의존성 확인 (Full 모드)
			if cfg.IsFullMode() {
				fmt.Fprintln(out, "\nDependencies:")
				statuses := detect.CheckDependencies(detect.FullModeDeps)
				for _, s := range statuses {
					if s.Installed {
						fmt.Fprintf(out, "  [OK] %s\n", s.Name)
					} else if s.Required {
						fmt.Fprintf(out, "  [ERROR] %s not installed (install: %s)\n", s.Name, s.InstallCmd)
						allOK = false
					} else {
						fmt.Fprintf(out, "  [WARN] %s not installed (optional, install: %s)\n", s.Name, s.InstallCmd)
					}
				}
			}

			// 4. 부모 디렉터리 규칙 충돌 검사
			conflicts := detect.CheckParentRuleConflicts(dir)
			if len(conflicts) > 0 {
				fmt.Fprintln(out, "\nRule Conflicts:")
				if cfg.IsolateRules {
					fmt.Fprintln(out, "  [OK] isolate_rules: true (parent rules ignored via CLAUDE.md directive)")
				}
				for _, c := range conflicts {
					if cfg.IsolateRules {
						fmt.Fprintf(out, "  [INFO] %s/.claude/rules/%s/ (ignored)\n", c.ParentDir, c.Namespace)
					} else {
						fmt.Fprintf(out, "  [WARN] Parent rules detected: %s/.claude/rules/%s/\n", c.ParentDir, c.Namespace)
						fmt.Fprintf(out, "         Run 'auto init' or 'auto update' to configure rule isolation.\n")
						allOK = false
					}
				}
			}

			// 5. 코딩 CLI 감지 상태
			fmt.Fprintln(out, "\nInstalled CLIs:")
			detected := detect.DetectPlatforms()
			if len(detected) == 0 {
				fmt.Fprintln(out, "  [WARN] No coding CLIs detected in PATH")
			} else {
				for _, p := range detected {
					fmt.Fprintf(out, "  [OK] %s (%s)\n", p.Name, p.Version)
				}
			}

			if allOK {
				fmt.Fprintln(out, "\nDiagnosis: All checks passed OK")
			} else {
				fmt.Fprintln(out, "\nDiagnosis: Issues found, run 'auto update' to fix")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "프로젝트 루트 디렉터리 (기본값: 현재 디렉터리)")
	return cmd
}
