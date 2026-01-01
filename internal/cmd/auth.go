// internal/cmd/auth.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/dub-cli/internal/auth"
	"github.com/salmonumbrella/dub-cli/internal/config"
	"github.com/salmonumbrella/dub-cli/internal/secrets"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Login, logout, and manage workspace credentials.",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthListCmd())
	cmd.AddCommand(newAuthSwitchCmd())
	cmd.AddCommand(newAuthStatusCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Dub",
		Long:  "Opens a browser to enter your Dub API key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := secrets.OpenDefault()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			server, err := auth.NewSetupServer(store)
			if err != nil {
				return err
			}

			result, err := server.Start(cmd.Context())
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully authenticated workspace: %s\n", result.WorkspaceName)
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	var workspace string

	cmd := &cobra.Command{
		Use:   "logout [workspace]",
		Short: "Remove workspace credentials",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				workspace = args[0]
			}
			if workspace == "" {
				return fmt.Errorf("workspace name required")
			}

			store, err := secrets.OpenDefault()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			if err := store.Delete(workspace); err != nil {
				return fmt.Errorf("failed to remove workspace: %w", err)
			}

			// Clear default if this was the default workspace
			defaultWs, _ := config.GetDefaultWorkspace()
			if defaultWs == workspace {
				_ = config.ClearDefaultWorkspace() // Best-effort cleanup
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed workspace: %s\n", workspace)
			return nil
		},
	}

	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace to remove")

	return cmd
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := secrets.OpenDefault()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			creds, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list workspaces: %w", err)
			}

			if len(creds) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No workspaces configured. Run: dub auth login")
				return nil
			}

			for _, c := range creds {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s (added %s)\n", c.Name, c.CreatedAt.Format("2006-01-02"))
			}
			return nil
		},
	}
}

func newAuthSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <workspace>",
		Short: "Set default workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := args[0]

			// Verify workspace exists in keyring
			store, err := secrets.OpenDefault()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			if _, err := store.Get(workspace); err != nil {
				return fmt.Errorf("workspace %q not found. Run: dub auth list", workspace)
			}

			// Set as default workspace
			if err := config.SetDefaultWorkspace(workspace); err != nil {
				return fmt.Errorf("failed to set default workspace: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Switched to workspace: %s\n", workspace)
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for environment variable authentication
			if apiKey := os.Getenv("DUB_API_KEY"); apiKey != "" {
				masked := apiKey[:7] + "..." + apiKey[len(apiKey)-4:]
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Authenticated via DUB_API_KEY environment variable\n")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "API Key: %s\n", masked)
				return nil
			}

			store, err := secrets.OpenDefault()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			creds, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list workspaces: %w", err)
			}

			if len(creds) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Not authenticated. Run: dub auth login")
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Authenticated with %d workspace(s)\n", len(creds))

			// Show default workspace if configured
			defaultWs, err := config.GetDefaultWorkspace()
			if err == nil && defaultWs != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Default workspace: %s\n", defaultWs)
			}

			return nil
		},
	}
}
