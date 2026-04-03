package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/worker/setup"
)

const defaultBackendURL = "https://api.autopus.co"

// runWorkerSetup executes the 3-step setup wizard:
//
//	Step 1: Device Auth OAuth (PKCE) → Autopus server login
//	Step 2: Workspace selection
//	Step 3: Provider auth check (claude/codex/gemini)
func runWorkerSetup(cmd *cobra.Command, backendURL string) error {
	out := cmd.OutOrStdout()
	if backendURL == "" {
		backendURL = defaultBackendURL
	}

	fmt.Fprintln(out, "🐙 Autopus Worker Setup")
	fmt.Fprintln(out, "───────────────────────────────")
	fmt.Fprintln(out)

	// Step 1: Device Auth
	token, err := stepDeviceAuth(cmd, backendURL)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Step 2: Workspace selection
	workspace, err := stepSelectWorkspace(cmd, backendURL, token)
	if err != nil {
		return fmt.Errorf("workspace selection failed: %w", err)
	}

	// Step 3: Save config and check providers
	if err := stepSaveAndCheckProviders(cmd, backendURL, token, workspace); err != nil {
		return fmt.Errorf("config save failed: %w", err)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "🐙 Setup complete!")
	fmt.Fprintf(out, "   Worker config: %s\n", setup.DefaultWorkerConfigPath())
	fmt.Fprintf(out, "   MCP config:    %s\n", setup.DefaultMCPConfigPath())
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Next: `auto worker start` to run the worker.")
	return nil
}

// stepDeviceAuth performs Device Authorization with PKCE.
func stepDeviceAuth(cmd *cobra.Command, backendURL string) (string, error) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Step 1/3: Autopus 서버 인증")

	_ = setup.SaveProgress(1)

	verifier, _, err := setup.GeneratePKCE()
	if err != nil {
		return "", fmt.Errorf("generate PKCE: %w", err)
	}

	dc, err := setup.RequestDeviceCode(backendURL, verifier)
	if err != nil {
		return "", fmt.Errorf("request device code: %w", err)
	}

	// Use the complete URI (includes code) if available, otherwise the base URI.
	verifyURL := dc.VerificationURIComplete
	if verifyURL == "" {
		verifyURL = dc.VerificationURI
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "  브라우저에서 아래 URL을 열고 코드를 입력하세요:\n")
	fmt.Fprintf(out, "  URL:  %s\n", verifyURL)
	fmt.Fprintf(out, "  Code: %s\n", dc.UserCode)
	fmt.Fprintln(out)

	// Try to open browser automatically.
	if verifyURL == "" {
		fmt.Fprintf(out, "  ⚠ 인증 URL이 비어있습니다. 백엔드 연결을 확인하세요.\n")
		return "", fmt.Errorf("empty verification URI from backend")
	}
	if err := setup.OpenBrowser(verifyURL); err != nil {
		fmt.Fprintf(out, "  브라우저를 수동으로 열어주세요.\n")
	} else {
		fmt.Fprintf(out, "  브라우저가 열렸습니다. 인증을 완료해주세요.\n")
	}

	fmt.Fprintln(out, "  인증 대기 중...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dc.ExpiresIn)*time.Second)
	defer cancel()

	tokenResp, err := setup.PollForToken(ctx, backendURL, dc.DeviceCode, verifier, dc.Interval)
	if err != nil {
		return "", fmt.Errorf("token polling: %w", err)
	}

	// Save credentials.
	creds := map[string]any{
		"access_token":  tokenResp.AccessToken,
		"refresh_token": tokenResp.RefreshToken,
		"backend_url":   backendURL,
		"created_at":    time.Now().Format(time.RFC3339),
	}
	if err := setup.SaveCredentials(creds); err != nil {
		return "", fmt.Errorf("save credentials: %w", err)
	}

	fmt.Fprintln(out, "  ✓ 인증 성공")
	return tokenResp.AccessToken, nil
}

// stepSelectWorkspace fetches workspaces and lets the user pick one.
func stepSelectWorkspace(cmd *cobra.Command, backendURL, token string) (*setup.Workspace, error) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Step 2/3: 워크스페이스 선택")

	_ = setup.SaveProgress(2)

	workspaces, err := setup.FetchWorkspaces(backendURL, token)
	if err != nil {
		return nil, fmt.Errorf("fetch workspaces: %w", err)
	}

	ws, err := setup.SelectWorkspace(workspaces)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(out, "  ✓ 워크스페이스: %s (%s)\n", ws.Name, ws.ID)
	return ws, nil
}

// stepSaveAndCheckProviders saves worker/MCP config and reports provider status.
func stepSaveAndCheckProviders(cmd *cobra.Command, backendURL, token string, ws *setup.Workspace) error {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Step 3/3: 설정 저장 및 프로바이더 확인")

	_ = setup.SaveProgress(3)

	// Save worker config.
	workerCfg := setup.WorkerConfig{
		BackendURL:  backendURL,
		WorkspaceID: ws.ID,
		Concurrency: 3,
	}

	// Detect installed providers and add to config.
	providers := setup.DetectProviders()
	for _, p := range providers {
		if p.Installed {
			workerCfg.Providers = append(workerCfg.Providers, p.Name)
		}
	}

	if err := setup.SaveWorkerConfig(workerCfg); err != nil {
		return fmt.Errorf("save worker config: %w", err)
	}
	fmt.Fprintf(out, "  ✓ Worker config 저장: %s\n", setup.DefaultWorkerConfigPath())

	// Generate and save MCP config.
	mcpCfg, err := setup.GenerateMCPConfig(setup.MCPConfigOptions{
		BackendURL:  backendURL,
		AuthToken:   token,
		WorkspaceID: ws.ID,
		OutputPath:  setup.DefaultMCPConfigPath(),
	})
	if err != nil {
		return fmt.Errorf("generate MCP config: %w", err)
	}

	if err := setup.WriteMCPConfig(mcpCfg, setup.DefaultMCPConfigPath()); err != nil {
		return fmt.Errorf("write MCP config: %w", err)
	}
	fmt.Fprintf(out, "  ✓ MCP config 저장: %s\n", setup.DefaultMCPConfigPath())

	// Report provider status.
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  프로바이더 상태:")
	for _, p := range providers {
		status := "❌ 미설치"
		if p.Installed {
			authed, guide := setup.CheckProviderAuth(p.Name)
			if authed {
				status = fmt.Sprintf("✅ %s (인증됨)", p.Version)
			} else {
				status = fmt.Sprintf("⚠️  %s (인증 필요: %s)", p.Version, guide)
			}
		}
		fmt.Fprintf(out, "    %-10s %s\n", p.Name, status)
	}

	_ = setup.ClearProgress()
	return nil
}
