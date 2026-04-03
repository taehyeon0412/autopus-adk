package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/internal/cli/tui"
	"github.com/insajin/autopus-adk/pkg/connect"
)

// @AX:NOTE [AUTO] @AX:REASON: hardcoded production server URL — overridable via --server flag
const (
	defaultServerURL = "https://api.autopus.co"
	// @AX:NOTE [AUTO] @AX:REASON: 5-minute timeout covers full 3-step wizard (server auth + workspace + OAuth)
	connectTimeout = 5 * time.Minute
)

func newConnectCmd() *cobra.Command {
	var (
		serverURL   string
		workspaceID string
	)

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect an AI provider via local OAuth flow",
		Long:  "3-step wizard: (1) Autopus server auth, (2) workspace selection, (3) OpenAI PKCE OAuth.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), connectTimeout)
			defer cancel()

			out := cmd.OutOrStdout()
			tui.SectionHeader(out, "Step 1: Autopus Server Authentication")

			authResult, err := stepServerAuth(ctx, serverURL)
			if err != nil {
				return fmt.Errorf("server authentication failed: %w", err)
			}
			tui.Success(out, "Authenticated with Autopus server")

			client := connect.NewClient(authResult.Token).WithServerURL(serverURL)

			wsID, wsName, err := stepWorkspaceSelect(ctx, client, workspaceID)
			if err != nil {
				return fmt.Errorf("workspace selection failed: %w", err)
			}
			tui.Successf(out, "Selected workspace: %s", wsName)

			tui.SectionHeader(out, "Step 3: OpenAI OAuth")

			oauthResult, err := stepOAuth(ctx)
			if err != nil {
				return fmt.Errorf("OpenAI OAuth failed: %w", err)
			}
			tui.Success(out, "OpenAI OAuth token obtained")

			err = stepSubmitToken(ctx, client, wsID, oauthResult)
			if err != nil {
				return fmt.Errorf("token submission failed: %w", err)
			}

			tui.Successf(out, "Connected OpenAI to workspace %q", wsName)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", defaultServerURL, "Autopus server URL")
	cmd.Flags().StringVar(&workspaceID, "workspace", "", "Skip workspace selection and use this ID")
	return cmd
}

func stepServerAuth(ctx context.Context, serverURL string) (*connect.AuthResult, error) {
	cfg := connect.ServerAuthConfig{ServerURL: serverURL}
	return connect.AuthenticateServer(ctx, cfg, nil)
}

func stepWorkspaceSelect(ctx context.Context, client *connect.Client, preselected string) (id, name string, err error) {
	workspaces, err := client.ListWorkspaces(ctx)
	if err != nil {
		return "", "", err
	}
	if len(workspaces) == 0 {
		// @AX:NOTE [AUTO] @AX:REASON: hardcoded app URL in error message — update if frontend domain changes
		return "", "", fmt.Errorf("no workspaces found — create one at https://app.autopus.co first")
	}

	// If --workspace flag provided, find the matching workspace.
	if preselected != "" {
		for _, ws := range workspaces {
			if ws.ID == preselected {
				return ws.ID, ws.Name, nil
			}
		}
		return "", "", fmt.Errorf("workspace %q not found", preselected)
	}

	// Single workspace: auto-select without prompting.
	if len(workspaces) == 1 {
		return workspaces[0].ID, workspaces[0].Name, nil
	}

	return promptWorkspaceSelect(workspaces)
}

func promptWorkspaceSelect(workspaces []connect.Workspace) (id, name string, err error) {
	options := make([]huh.Option[string], len(workspaces))
	for i, ws := range workspaces {
		label := ws.Name
		if ws.Description != "" {
			label = fmt.Sprintf("%s — %s", ws.Name, ws.Description)
		}
		options[i] = huh.NewOption(label, ws.ID)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a workspace").
				Options(options...).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		return "", "", fmt.Errorf("workspace selection cancelled: %w", err)
	}

	for _, ws := range workspaces {
		if ws.ID == selected {
			return ws.ID, ws.Name, nil
		}
	}
	return selected, selected, nil
}

func stepOAuth(ctx context.Context) (*connect.OAuthResult, error) {
	cfg := connect.OAuthConfig{ClientID: connect.DefaultClientID()}
	return connect.WaitForCallback(ctx, cfg)
}

func stepSubmitToken(ctx context.Context, client *connect.Client, wsID string, oauth *connect.OAuthResult) error {
	req := connect.SubmitTokenRequest{
		ProviderToken: oauth.AccessToken,
		WorkspaceID:   wsID,
		Provider:      "openai",
	}
	return client.SubmitToken(ctx, req)
}
