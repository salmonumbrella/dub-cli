// internal/cmd/folders.go
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
		output string
		limit  int
		all    bool
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

			path := "/folders"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}

			return handleFoldersListResponse(cmd, resp, output, limit, all)
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of folders to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all folders (ignore limit)")

	return cmd
}

// handleFoldersListResponse handles the response for folders list command,
// formatting output as table or JSON based on the output flag.
func handleFoldersListResponse(cmd *cobra.Command, resp *http.Response, output string, limit int, all bool) error {
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

	// Parse folders for table output
	var folders []map[string]interface{}
	if err := json.Unmarshal(body, &folders); err != nil {
		return fmt.Errorf("failed to parse folders: %w", err)
	}

	totalCount := len(folders)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayFolders := folders[:displayLimit]

	// Define table columns
	columns := []outfmt.Column{
		{Name: "Name", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Type", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Access Level", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Links", Width: 0, Align: outfmt.AlignRight},
	}

	// Build rows
	rows := make([][]string, len(displayFolders))
	for i, folder := range displayFolders {
		rows[i] = []string{
			outfmt.SafeString(folder["name"]),
			formatFolderType(folder["type"]),
			formatAccessLevel(folder["accessLevel"]),
			formatFolderLinkCount(folder),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d folders. Use --limit or --all for more.\n", displayLimit, totalCount)
	}

	return nil
}

// formatFolderType formats the folder type or returns "-" if not set.
func formatFolderType(folderType interface{}) string {
	s := outfmt.SafeString(folderType)
	if s == "" {
		return "-"
	}
	return s
}

// formatAccessLevel formats the access level or returns "-" if not set.
func formatAccessLevel(accessLevel interface{}) string {
	s := outfmt.SafeString(accessLevel)
	if s == "" {
		return "-"
	}
	return s
}

// formatFolderLinkCount extracts the link count from folder data.
// The API returns link count in _count.links nested structure.
func formatFolderLinkCount(folder map[string]interface{}) string {
	// Try _count.links nested structure first
	if countObj, ok := folder["_count"].(map[string]interface{}); ok {
		if links, ok := countObj["links"]; ok {
			return formatClicks(outfmt.SafeInt(links))
		}
	}

	// Fallback to direct links field
	if links, ok := folder["links"]; ok {
		return formatClicks(outfmt.SafeInt(links))
	}

	return "0"
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
