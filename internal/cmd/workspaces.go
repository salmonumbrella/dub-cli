// internal/cmd/workspaces.go
package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

func newWorkspacesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspaces",
		Short: "Manage workspaces",
		Long:  "Get and update workspace settings.",
	}

	cmd.AddCommand(newWorkspacesGetCmd())
	cmd.AddCommand(newWorkspacesUpdateCmd())

	return cmd
}

func newWorkspacesGetCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get workspace info",
		Long:  "Get details of a specific workspace by ID or slug.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Get(cmd.Context(), "/workspaces/"+url.PathEscape(id))
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Workspace ID or slug (required)")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newWorkspacesUpdateCmd() *cobra.Command {
	var (
		id   string
		name string
		slug string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a workspace",
		Long:  "Update an existing workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{}
			if cmd.Flags().Changed("name") {
				body["name"] = name
			}
			if cmd.Flags().Changed("slug") {
				body["slug"] = slug
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one of --name or --slug must be specified")
			}

			resp, err := client.Patch(cmd.Context(), "/workspaces/"+url.PathEscape(id), body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&name, "name", "", "New workspace name")
	cmd.Flags().StringVar(&slug, "slug", "", "New workspace slug")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}
