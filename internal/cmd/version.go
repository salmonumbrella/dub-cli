// internal/cmd/version.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via -ldflags
	Version = "dev"
	// Commit is set at build time via -ldflags
	Commit = "unknown"
	// Date is set at build time via -ldflags
	Date = "unknown"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print the version and build information of the Dub CLI.",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dub %s\n", Version)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "commit: %s\n", Commit)
			if Date != "unknown" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "built:  %s\n", Date)
			}
		},
	}

	return cmd
}
