// internal/cmd/links.go
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

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

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().StringVar(&domain, "domain", "", "Filter by domain")

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

func newLinksUpdateCmd() *cobra.Command {
	var (
		id      string
		linkURL string
		key     string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a link",
		Long:  "Update an existing link by ID.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{}
			if linkURL != "" {
				body["url"] = linkURL
			}
			if key != "" {
				body["key"] = key
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one of --url or --key must be specified")
			}

			resp, err := client.Patch(cmd.Context(), "/links/"+url.PathEscape(id), body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Link ID (required)")
	cmd.Flags().StringVar(&linkURL, "url", "", "New destination URL")
	cmd.Flags().StringVar(&key, "key", "", "New short key")

	_ = cmd.MarkFlagRequired("id")

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
