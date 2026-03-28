package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/detect"
)

// @AX:NOTE [AUTO] subcommand registration point for "auto permission" — extend here to add new permission subcommands
func newPermissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permission",
		Short: "Manage permission detection",
	}
	cmd.AddCommand(newPermissionDetectCmd())
	return cmd
}

func newPermissionDetectCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect parent process permission mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			result := detect.DetectPermissionMode()
			if jsonOutput {
				data, err := json.Marshal(result)
				if err != nil {
					return fmt.Errorf("JSON marshal failed: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), result.Mode)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
