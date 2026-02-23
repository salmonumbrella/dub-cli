// internal/cmd/links.go
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/dub-cli/internal/api"
	"github.com/salmonumbrella/dub-cli/internal/outfmt"
)

func handleResponse(cmd *cobra.Command, resp *http.Response) error {
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		apiErr := api.ParseAPIError(body)
		return fmt.Errorf("%s", apiErr.Error())
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
		return nil
	}

	query := outfmt.GetQuery(cmd.Context())
	return outfmt.FormatJSON(cmd.OutOrStdout(), data, query)
}

// Link represents a Dub link from the API response.
type Link struct {
	ID          string  `json:"id"`
	Domain      string  `json:"domain"`
	Key         string  `json:"key"`
	URL         string  `json:"url"`
	Clicks      int     `json:"clicks"`
	LastClicked *string `json:"lastClicked"`
}

// handleLinksListResponse handles the response for links list command,
// formatting output as table or JSON based on the output flag.
func handleLinksListResponse(cmd *cobra.Command, resp *http.Response, output string, limit int, all bool) error {
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

	// Parse links for table output
	var links []Link
	if err := json.Unmarshal(body, &links); err != nil {
		return fmt.Errorf("failed to parse links: %w", err)
	}

	totalCount := len(links)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayLinks := links[:displayLimit]

	// Define table columns
	columns := []outfmt.Column{
		{Name: "Short Link", Width: 0, Align: outfmt.AlignLeft},
		{Name: "URL", Width: 40, Align: outfmt.AlignLeft},
		{Name: "Clicks", Width: 0, Align: outfmt.AlignRight},
		{Name: "Last Clicked", Width: 0, Align: outfmt.AlignLeft},
	}

	// Build rows
	rows := make([][]string, len(displayLinks))
	for i, link := range displayLinks {
		rows[i] = []string{
			buildShortLink(link.Domain, link.Key),
			outfmt.Truncate(link.URL, 40),
			formatClicks(link.Clicks),
			formatLastClicked(link.LastClicked),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d links. Use --limit or --all for more.\n", displayLimit, totalCount)
	}

	return nil
}

// buildShortLink combines domain and key into a short link.
func buildShortLink(domain, key string) string {
	return domain + "/" + key
}

// formatClicks formats a click count with comma separators.
func formatClicks(clicks int) string {
	if clicks == 0 {
		return "0"
	}

	s := strconv.Itoa(clicks)
	n := len(s)

	// Calculate number of commas needed
	commaCount := (n - 1) / 3
	if commaCount == 0 {
		return s
	}

	result := make([]byte, n+commaCount)
	resultIdx := len(result) - 1

	for i := n - 1; i >= 0; i-- {
		pos := n - 1 - i
		if pos > 0 && pos%3 == 0 {
			result[resultIdx] = ','
			resultIdx--
		}
		result[resultIdx] = s[i]
		resultIdx--
	}

	return string(result)
}

// formatLastClicked formats an ISO 8601 timestamp to "Jan 15, 2024" format.
// Returns "-" if the timestamp is nil or empty.
func formatLastClicked(ts *string) string {
	if ts == nil || *ts == "" {
		return "-"
	}

	t, err := time.Parse(time.RFC3339, *ts)
	if err != nil {
		// Try alternative formats
		t, err = time.Parse("2006-01-02T15:04:05Z", *ts)
		if err != nil {
			return "-"
		}
	}

	return t.Format("Jan 2, 2006")
}

func newLinksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "links",
		Short: "Manage links",
		Long:  "Create, list, update, and delete short links.",
	}

	cmd.AddCommand(newLinksCreateCmd())
	cmd.AddCommand(newLinksListCmd())
	cmd.AddCommand(newLinksGetCmd())
	cmd.AddCommand(newLinksCountCmd())
	cmd.AddCommand(newLinksUpdateCmd())
	cmd.AddCommand(newLinksUpsertCmd())
	cmd.AddCommand(newLinksDeleteCmd())
	cmd.AddCommand(newLinksBulkCmd())

	return cmd
}

func newLinksCreateCmd() *cobra.Command {
	var (
		linkURL string
		key     string
		domain  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new short link",
		Long:  "Create a new short link with the specified URL.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if linkURL == "" {
				return fmt.Errorf("--url is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"url": linkURL,
			}
			if key != "" {
				body["key"] = key
			}
			if domain != "" {
				body["domain"] = domain
			}

			resp, err := client.Post(cmd.Context(), "/links", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&linkURL, "url", "", "Destination URL (required)")
	cmd.Flags().StringVar(&key, "key", "", "Custom short key (optional)")
	cmd.Flags().StringVar(&domain, "domain", "", "Domain for the short link (optional)")

	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func newLinksListCmd() *cobra.Command {
	var (
		search string
		domain string
		output string
		limit  int
		all    bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List links",
		Long:  "List all links in the workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			if search != "" {
				params.Set("search", search)
			}
			if domain != "" {
				params.Set("domain", domain)
			}

			path := "/links"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}

			return handleLinksListResponse(cmd, resp, output, limit, all)
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().StringVar(&domain, "domain", "", "Filter by domain")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of links to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all links (ignore limit)")

	return cmd
}

func newLinksGetCmd() *cobra.Command {
	var (
		id     string
		domain string
		key    string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a link",
		Long:  "Get a link by ID or by domain and key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate flags first before auth
			if id == "" && (domain == "" || key == "") {
				return fmt.Errorf("either --id or both --domain and --key are required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			var path string
			if id != "" {
				path = "/links/" + url.PathEscape(id)
			} else {
				params := url.Values{}
				params.Set("domain", domain)
				params.Set("key", key)
				path = "/links/info?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Link ID")
	cmd.Flags().StringVar(&domain, "domain", "", "Domain (used with --key)")
	cmd.Flags().StringVar(&key, "key", "", "Short key (used with --domain)")

	return cmd
}

func newLinksCountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Count links",
		Long:  "Get the total count of links in the workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Get(cmd.Context(), "/links/count")
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	return cmd
}

// resolveLink looks up a link by domain and key, returning the link ID.
func resolveLink(ctx context.Context, client *api.Client, domain, key string) (string, error) {
	params := url.Values{}
	params.Set("domain", domain)
	params.Set("key", key)

	resp, err := client.Get(ctx, "/links/info?"+params.Encode())
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		apiErr := api.ParseAPIError(body)
		return "", fmt.Errorf("failed to resolve link %s/%s: %s", domain, key, apiErr.Error())
	}

	var link struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &link); err != nil {
		return "", fmt.Errorf("failed to parse link info: %w", err)
	}

	if link.ID == "" {
		return "", fmt.Errorf("link %s/%s not found", domain, key)
	}

	return link.ID, nil
}

func newLinksUpdateCmd() *cobra.Command {
	var (
		id      string
		domain  string
		linkURL string
		key     string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a link",
		Long:  "Update an existing link by ID or by domain and key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" && (domain == "" || key == "") {
				return fmt.Errorf("either --id or both --domain and --key are required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Resolve link ID if using domain+key lookup
			linkID := id
			if linkID == "" {
				resolved, err := resolveLink(cmd.Context(), client, domain, key)
				if err != nil {
					return err
				}
				linkID = resolved
			}

			body := map[string]interface{}{}
			if linkURL != "" {
				body["url"] = linkURL
			}
			// key is only a field to update when identifying by --id
			if id != "" && key != "" {
				body["key"] = key
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one update field (--url) must be specified")
			}

			resp, err := client.Patch(cmd.Context(), "/links/"+url.PathEscape(linkID), body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Link ID")
	cmd.Flags().StringVar(&domain, "domain", "", "Domain (used with --key to identify link)")
	cmd.Flags().StringVar(&linkURL, "url", "", "New destination URL")
	cmd.Flags().StringVar(&key, "key", "", "Short key (used with --domain to identify link, or with --id to rename)")

	return cmd
}

func newLinksUpsertCmd() *cobra.Command {
	var (
		linkURL string
		key     string
		domain  string
	)

	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "Create or update a link",
		Long:  "Create a new link or update an existing one if it matches.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if linkURL == "" {
				return fmt.Errorf("--url is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"url": linkURL,
			}
			if key != "" {
				body["key"] = key
			}
			if domain != "" {
				body["domain"] = domain
			}

			resp, err := client.Put(cmd.Context(), "/links/upsert", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&linkURL, "url", "", "Destination URL (required)")
	cmd.Flags().StringVar(&key, "key", "", "Custom short key (optional)")
	cmd.Flags().StringVar(&domain, "domain", "", "Domain for the short link (optional)")

	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func newLinksDeleteCmd() *cobra.Command {
	var (
		id     string
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a link",
		Long:  "Delete a link by ID.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Would delete link with ID: %s\n", id)
				return nil
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Delete(cmd.Context(), "/links/"+url.PathEscape(id))
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Link ID (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newLinksBulkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk link operations",
		Long:  "Perform bulk operations on links (create, update, delete).",
	}

	cmd.AddCommand(newLinksBulkCreateCmd())
	cmd.AddCommand(newLinksBulkUpdateCmd())
	cmd.AddCommand(newLinksBulkDeleteCmd())

	return cmd
}

func newLinksBulkCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Bulk create links",
		Long:  "Create multiple links from JSON input (reads from stdin).",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}

			var body interface{}
			if err := json.Unmarshal(input, &body); err != nil {
				return fmt.Errorf("invalid JSON input: %w", err)
			}

			resp, err := client.Post(cmd.Context(), "/links/bulk", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	return cmd
}

func newLinksBulkUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Bulk update links",
		Long:  "Update multiple links from JSON input (reads from stdin).",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}

			var body interface{}
			if err := json.Unmarshal(input, &body); err != nil {
				return fmt.Errorf("invalid JSON input: %w", err)
			}

			resp, err := client.Patch(cmd.Context(), "/links/bulk", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	return cmd
}

func newLinksBulkDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Bulk delete links",
		Long:  "Delete multiple links from JSON input (reads from stdin).",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}

			var body interface{}
			if err := json.Unmarshal(input, &body); err != nil {
				return fmt.Errorf("invalid JSON input: %w", err)
			}

			resp, err := client.DeleteWithBody(cmd.Context(), "/links/bulk", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	return cmd
}
