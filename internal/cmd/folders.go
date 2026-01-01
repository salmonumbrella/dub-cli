// internal/cmd/folders.go
package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

func newFoldersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folders",
		Short: "Manage folders",
		Long:  "Create, list, update, and delete folders for organizing links.",
	}

	cmd.AddCommand(newFoldersCreateCmd())
	cmd.AddCommand(newFoldersListCmd())
	cmd.AddCommand(newFoldersUpdateCmd())
	cmd.AddCommand(newFoldersDeleteCmd())

	return cmd
}

func newFoldersCreateCmd() *cobra.Command {
	var (
		name     string
		parentID string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a folder",
		Long:  "Create a new folder for organizing links.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"name": name,
			}
			if parentID != "" {
				body["parentId"] = parentID
			}

			resp, err := client.Post(cmd.Context(), "/folders", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Folder name (required)")
	cmd.Flags().StringVar(&parentID, "parent-id", "", "Parent folder ID (for nested folders)")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newFoldersListCmd() *cobra.Command {
	var (
		search string
		page   int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List folders",
		Long:  "List all folders in your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			if search != "" {
				params.Set("search", search)
			}
			if page > 0 {
				params.Set("page", fmt.Sprintf("%d", page))
			}

			path := "/folders"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")

	return cmd
}

func newFoldersUpdateCmd() *cobra.Command {
	var (
		id       string
		name     string
		parentID string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a folder",
		Long:  "Update an existing folder.",
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
			if cmd.Flags().Changed("parent-id") {
				body["parentId"] = parentID
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one of --name or --parent-id must be specified")
			}

			resp, err := client.Patch(cmd.Context(), "/folders/"+url.PathEscape(id), body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Folder ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "New folder name")
	cmd.Flags().StringVar(&parentID, "parent-id", "", "New parent folder ID")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newFoldersDeleteCmd() *cobra.Command {
	var (
		id     string
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a folder",
		Long:  "Delete a folder from your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Would delete folder with ID: %s\n", id)
				return nil
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Delete(cmd.Context(), "/folders/"+url.PathEscape(id))
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Folder ID (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}
