package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/pipeline"
)

// newPipelineCmd creates the `auto pipeline` parent command with subcommands.
func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline monitoring and management",
	}

	cmd.AddCommand(newPipelineDashboardCmd())

	return cmd
}

// newPipelineDashboardCmd creates the `auto pipeline dashboard <spec-id>` subcommand.
// It renders a one-shot pipeline dashboard to stdout (R8).
func newPipelineDashboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard <spec-id>",
		Short: "Render pipeline dashboard for a spec",
		Args:  cobra.ExactArgs(1),
		// @AX:TODO [AUTO] stub implementation — dashboard data is hardcoded to all-pending; must read actual pipeline state from checkpoint file
		RunE: func(cmd *cobra.Command, args []string) error {
			specID := args[0]

			if err := pipeline.ValidateSpecID(specID); err != nil {
				return err
			}

			data := pipeline.DashboardData{
				Phases: map[string]pipeline.PhaseStatus{
					"phase1":   pipeline.PhasePending,
					"phase1.5": pipeline.PhasePending,
					"phase2":   pipeline.PhasePending,
					"phase3":   pipeline.PhasePending,
					"phase4":   pipeline.PhasePending,
				},
				Agents: map[string]string{},
			}

			output := pipeline.RenderDashboard(data)
			fmt.Fprintf(cmd.OutOrStdout(), "SPEC: %s\n%s", specID, output)
			return nil
		},
	}
}
