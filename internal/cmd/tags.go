// internal/cmd/tags.go
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/dub-cli/internal/api"
	"github.com/salmonumbrella/dub-cli/internal/outfmt"
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
		output string
		limit  int
		all    bool
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

			path := "/tags"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}

			return handleTagsListResponse(cmd, resp, output, limit, all)
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of tags to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all tags (ignore limit)")

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

// handleTagsListResponse handles the response for tags list command,
// formatting output as table or JSON based on the output flag.
func handleTagsListResponse(cmd *cobra.Command, resp *http.Response, output string, limit int, all bool) error {
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		apiErr := api.ParseAPIError(body)
		return fmt.Errorf("%s", apiErr.Error())
	}

	// For JSON output, use the existing handler
	if output == "json" {
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
			return nil
		}
		query := outfmt.GetQuery(cmd.Context())
		return outfmt.FormatJSON(cmd.OutOrStdout(), data, query)
	}

	// Parse tags for table output
	var tags []map[string]interface{}
	if err := json.Unmarshal(body, &tags); err != nil {
		return fmt.Errorf("failed to parse tags: %w", err)
	}

	totalCount := len(tags)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayTags := tags[:displayLimit]

	// Define table columns
	columns := []outfmt.Column{
		{Name: "Name", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Color", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Links", Width: 0, Align: outfmt.AlignRight},
	}

	// Build rows
	rows := make([][]string, len(displayTags))
	for i, tag := range displayTags {
		rows[i] = []string{
			outfmt.SafeString(tag["name"]),
			formatTagColor(tag["color"]),
			formatTagLinkCount(tag),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d tags. Use --limit or --all for more.\n", displayLimit, totalCount)
	}

	return nil
}

// formatTagColor formats the tag color or returns "-" if not set.
func formatTagColor(color interface{}) string {
	s := outfmt.SafeString(color)
	if s == "" {
		return "-"
	}
	return s
}

// formatTagLinkCount extracts the link count from tag data.
// The API returns link count in _count.links nested structure.
func formatTagLinkCount(tag map[string]interface{}) string {
	// Try _count.links nested structure first
	if countObj, ok := tag["_count"].(map[string]interface{}); ok {
		if links, ok := countObj["links"]; ok {
			return formatClicks(outfmt.SafeInt(links))
		}
	}

	// Fallback to direct links field
	if links, ok := tag["links"]; ok {
		return formatClicks(outfmt.SafeInt(links))
	}

	return "0"
}
