// internal/cmd/tags.go
package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

func newTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage tags",
		Long:  "Create, list, and update tags for organizing links.",
	}

	cmd.AddCommand(newTagsCreateCmd())
	cmd.AddCommand(newTagsListCmd())
	cmd.AddCommand(newTagsUpdateCmd())

	return cmd
}

func newTagsCreateCmd() *cobra.Command {
	var (
		name  string
		color string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tag",
		Long:  "Create a new tag for organizing links.",
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
			if color != "" {
				body["color"] = color
			}

			resp, err := client.Post(cmd.Context(), "/tags", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Tag name (required)")
	cmd.Flags().StringVar(&color, "color", "", "Tag color (e.g., red, blue, green)")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTagsListCmd() *cobra.Command {
	var (
		search string
		page   int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tags",
		Long:  "List all tags in your workspace.",
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

			path := "/tags"
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

func newTagsUpdateCmd() *cobra.Command {
	var (
		id    string
		name  string
		color string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a tag",
		Long:  "Update an existing tag.",
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
			if cmd.Flags().Changed("color") {
				body["color"] = color
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one of --name or --color must be specified")
			}

			resp, err := client.Patch(cmd.Context(), "/tags/"+url.PathEscape(id), body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Tag ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "New tag name")
	cmd.Flags().StringVar(&color, "color", "", "New tag color")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}
