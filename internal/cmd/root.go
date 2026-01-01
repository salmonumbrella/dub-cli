// internal/cmd/root.go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/salmonumbrella/dub-cli/internal/debug"
	"github.com/salmonumbrella/dub-cli/internal/outfmt"
	"github.com/salmonumbrella/dub-cli/internal/ui"
	"github.com/spf13/cobra"
)

type rootFlags struct {
	Workspace string
	Output    string
	Query     string
	Yes       bool
	Debug     bool
	Limit     int
	SortBy    string
	Desc      bool
	Color     string
}

type contextKey string

const workspaceKey contextKey = "workspace"

// GetWorkspace returns the workspace name from context
func GetWorkspace(ctx context.Context) string {
	if v, ok := ctx.Value(workspaceKey).(string); ok {
		return v
	}
	return ""
}

func NewRootCmd() *cobra.Command {
	// flags is local to this function to avoid package-level mutable state
	// that could cause issues with parallel tests
	var flags rootFlags

	cmd := &cobra.Command{
		Use:          "dub",
		Short:        "Dub CLI - manage your Dub links from the terminal",
		Long:         "dub - A command-line interface for the Dub API. Manage links, analytics, domains, and more.",
		Version:      Version,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize debug logging based on --debug flag
			debug.Init(flags.Debug)

			// Initialize UI color output based on --color flag
			ui.Init(flags.Color)

			if flags.Desc && flags.SortBy == "" {
				return fmt.Errorf("--desc requires --sort-by to be specified")
			}

			// Wire global flags to context
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = outfmt.WithFormat(ctx, flags.Output)
			ctx = outfmt.WithQuery(ctx, flags.Query)
			ctx = outfmt.WithYes(ctx, flags.Yes)
			ctx = outfmt.WithLimit(ctx, flags.Limit)
			ctx = outfmt.WithSortBy(ctx, flags.SortBy)
			ctx = outfmt.WithDesc(ctx, flags.Desc)
			ctx = context.WithValue(ctx, workspaceKey, flags.Workspace)
			cmd.SetContext(ctx)

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&flags.Workspace, "workspace", "w", os.Getenv("DUB_WORKSPACE"), "Workspace name (or DUB_WORKSPACE env)")
	cmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", getEnvOrDefault("DUB_OUTPUT", "text"), "Output format: text|json")
	cmd.PersistentFlags().StringVar(&flags.Query, "query", "", "JQ filter expression for JSON output")
	cmd.PersistentFlags().BoolVarP(&flags.Yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.PersistentFlags().BoolVar(&flags.Yes, "force", false, "Skip confirmation prompts (alias for --yes)")
	cmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Enable debug output")
	cmd.PersistentFlags().IntVar(&flags.Limit, "limit", 0, "Limit number of results (0 = no limit)")
	cmd.PersistentFlags().StringVar(&flags.SortBy, "sort-by", "", "Field name to sort by")
	cmd.PersistentFlags().BoolVar(&flags.Desc, "desc", false, "Sort descending (requires --sort-by)")
	cmd.PersistentFlags().StringVar(&flags.Color, "color", "auto", "Color output: auto|always|never")

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newLinksCmd())
	cmd.AddCommand(newAnalyticsCmd())
	cmd.AddCommand(newEventsCmd())
	cmd.AddCommand(newDomainsCmd())
	cmd.AddCommand(newPartnersCmd())
	cmd.AddCommand(newCustomersCmd())
	cmd.AddCommand(newCommissionsCmd())
	cmd.AddCommand(newTrackCmd())
	cmd.AddCommand(newTagsCmd())
	cmd.AddCommand(newFoldersCmd())
	cmd.AddCommand(newWorkspacesCmd())
	cmd.AddCommand(newQRCmd())
	cmd.AddCommand(newEmbedCmd())
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newUpgradeCmd())
	cmd.AddCommand(newCompletionCmd())

	return cmd
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Execute(args []string) error {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	return cmd.Execute()
}

func ExecuteContext(ctx context.Context, args []string) error {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}
