package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/setup"
)

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Generate and manage project documentation for AI agents",
	}

	cmd.AddCommand(newSetupGenerateCmd())
	cmd.AddCommand(newSetupUpdateCmd())
	cmd.AddCommand(newSetupValidateCmd())
	cmd.AddCommand(newSetupStatusCmd())
	return cmd
}

func newSetupGenerateCmd() *cobra.Command {
	var (
		force     bool
		outputDir string
	)

	cmd := &cobra.Command{
		Use:   "generate [dir]",
		Short: "Generate project documentation in .autopus/docs/",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := resolveDirFromArgs(args)
			if err != nil {
				return err
			}

			opts := &setup.GenerateOptions{
				OutputDir: outputDir,
				Force:     force,
			}

			_, genErr := setup.Generate(dir, opts)
			if genErr != nil {
				return genErr
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Documentation generated successfully in .autopus/docs/")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing documentation")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (default: .autopus/docs/)")
	return cmd
}

func newSetupUpdateCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "update [dir]",
		Short: "Update changed documentation files",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := resolveDirFromArgs(args)
			if err != nil {
				return err
			}

			updated, updateErr := setup.Update(dir, outputDir)
			if updateErr != nil {
				return updateErr
			}

			if len(updated) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "All documents are up to date.")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Updated %d document(s):\n", len(updated))
				for _, f := range updated {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", f)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "", "Documentation directory (default: .autopus/docs/)")
	return cmd
}

func newSetupValidateCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "validate [dir]",
		Short: "Validate documentation against current code state",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := resolveDirFromArgs(args)
			if err != nil {
				return err
			}

			docsDir := resolveOutputDir(dir, outputDir)
			report, valErr := setup.Validate(docsDir, dir)
			if valErr != nil {
				return valErr
			}

			// Also check command validity
			cmdWarnings := setup.ValidateCommands(docsDir, dir)
			report.Warnings = append(report.Warnings, cmdWarnings...)
			if len(cmdWarnings) > 0 {
				report.Valid = false
			}

			if report.Valid && len(report.Warnings) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "All documents are up to date.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Validation issues (%d):\n", len(report.Warnings))
			for _, w := range report.Warnings {
				loc := w.File
				if w.Line > 0 {
					loc = fmt.Sprintf("%s:%d", w.File, w.Line)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s: %s\n", w.Type, loc, w.Message)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nDrift score: %.1f%%\n", report.DriftScore*100)

			return fmt.Errorf("%d validation issue(s) found", len(report.Warnings))
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "", "Documentation directory (default: .autopus/docs/)")
	return cmd
}

func newSetupStatusCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "status [dir]",
		Short: "Show documentation status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := resolveDirFromArgs(args)
			if err != nil {
				return err
			}

			status, statusErr := setup.GetStatus(dir, outputDir)
			if statusErr != nil {
				return statusErr
			}

			if !status.Exists {
				fmt.Fprintln(cmd.OutOrStdout(), "No documentation found. Run `auto setup generate` to create.")
				return nil
			}

			w := cmd.OutOrStdout()
			fmt.Fprintln(w, "Documentation Status")
			fmt.Fprintln(w, strings.Repeat("-", 40))

			if !status.GeneratedAt.IsZero() {
				ago := time.Since(status.GeneratedAt).Round(time.Hour)
				fmt.Fprintf(w, "Last generated: %s (%s ago)\n",
					status.GeneratedAt.Format("2006-01-02 15:04"),
					ago)
			}
			fmt.Fprintf(w, "Drift score:    %.1f%%\n\n", status.DriftScore*100)

			fmt.Fprintln(w, "Files:")
			for fileName, fs := range status.FileStatuses {
				state := "missing"
				if fs.Exists {
					if fs.Fresh {
						state = "fresh"
					} else {
						state = "stale"
					}
				}
				fmt.Fprintf(w, "  %-20s %s\n", fileName, state)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "", "Documentation directory (default: .autopus/docs/)")
	return cmd
}

func resolveDirFromArgs(args []string) (string, error) {
	dir := ""
	if len(args) > 0 {
		dir = args[0]
	}
	return resolveDir(dir)
}

func resolveOutputDir(projectDir, outputDir string) string {
	if outputDir != "" {
		return outputDir
	}
	return projectDir + "/.autopus/docs"
}
