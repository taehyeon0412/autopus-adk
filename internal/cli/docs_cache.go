package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/docs"
)

// newDocsCacheCmd creates the `auto docs cache` subcommand with list and clear subcommands.
func newDocsCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the documentation cache",
	}
	cmd.AddCommand(newDocsCacheListCmd())
	cmd.AddCommand(newDocsCacheClearCmd())
	return cmd
}

// newDocsCacheListCmd creates the `auto docs cache list` subcommand.
func newDocsCacheListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List cached documentation entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cache := docs.NewCache(docsCacheDir(), 24*time.Hour)
			entries, err := cache.List()
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no cached entries")
				return nil
			}
			for _, e := range entries {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s  (expires: %s)\n", e.Key, e.ExpiresAt.Format(time.RFC3339))
			}
			return nil
		},
	}
}

// newDocsCacheClearCmd creates the `auto docs cache clear` subcommand.
func newDocsCacheClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all cached documentation entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cache := docs.NewCache(docsCacheDir(), 24*time.Hour)
			if err := cache.Clear(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "cache clear: done")
			return nil
		},
	}
}
