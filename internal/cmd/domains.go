// internal/cmd/domains.go
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

func newDomainsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domains",
		Short: "Manage domains",
		Long:  "Create, list, update, and delete custom domains.",
	}

	cmd.AddCommand(newDomainsCreateCmd())
	cmd.AddCommand(newDomainsListCmd())
	cmd.AddCommand(newDomainsUpdateCmd())
	cmd.AddCommand(newDomainsDeleteCmd())
	cmd.AddCommand(newDomainsRegisterCmd())
	cmd.AddCommand(newDomainsCheckCmd())

	return cmd
}

func newDomainsCreateCmd() *cobra.Command {
	var (
		slug        string
		placeholder string
		expiredURL  string
		archived    bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a domain",
		Long:  "Add a custom domain to your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if slug == "" {
				return fmt.Errorf("--slug is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"slug": slug,
			}
			if placeholder != "" {
				body["placeholder"] = placeholder
			}
			if expiredURL != "" {
				body["expiredUrl"] = expiredURL
			}
			if archived {
				body["archived"] = archived
			}

			resp, err := client.Post(cmd.Context(), "/domains", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Domain name (required)")
	cmd.Flags().StringVar(&placeholder, "placeholder", "", "Placeholder URL for root domain")
	cmd.Flags().StringVar(&expiredURL, "expired-url", "", "URL for expired links")
	cmd.Flags().BoolVar(&archived, "archived", false, "Archive the domain")

	_ = cmd.MarkFlagRequired("slug")

	return cmd
}

func newDomainsListCmd() *cobra.Command {
	var (
		archived bool
		search   string
		output   string
		limit    int
		all      bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
		Long:  "List all domains in your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			if archived {
				params.Set("archived", "true")
			}
			if search != "" {
				params.Set("search", search)
			}

			path := "/domains"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}

			return handleDomainsListResponse(cmd, resp, output, limit, all)
		},
	}

	cmd.Flags().BoolVar(&archived, "archived", false, "Include archived domains")
	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of domains to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all domains (ignore limit)")

	return cmd
}

// Domain represents a Dub domain from the API response.
type Domain struct {
	ID          string  `json:"id"`
	Slug        string  `json:"slug"`
	Verified    bool    `json:"verified"`
	Placeholder *string `json:"placeholder"`
	Links       int     `json:"_count,omitempty"`
}

// handleDomainsListResponse handles the response for domains list command,
// formatting output as table or JSON based on the output flag.
func handleDomainsListResponse(cmd *cobra.Command, resp *http.Response, output string, limit int, all bool) error {
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

	// Parse domains for table output
	var domains []map[string]interface{}
	if err := json.Unmarshal(body, &domains); err != nil {
		return fmt.Errorf("failed to parse domains: %w", err)
	}

	totalCount := len(domains)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayDomains := domains[:displayLimit]

	// Define table columns
	columns := []outfmt.Column{
		{Name: "Domain", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Verified", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Placeholder", Width: 40, Align: outfmt.AlignLeft},
		{Name: "Links", Width: 0, Align: outfmt.AlignRight},
	}

	// Build rows
	rows := make([][]string, len(displayDomains))
	for i, domain := range displayDomains {
		rows[i] = []string{
			outfmt.SafeString(domain["slug"]),
			outfmt.FormatBool(domain["verified"]),
			formatPlaceholder(domain["placeholder"]),
			formatLinkCount(domain),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d domains. Use --limit or --all for more.\n", displayLimit, totalCount)
	}

	return nil
}

// formatPlaceholder formats the placeholder URL or returns "-" if not set.
func formatPlaceholder(placeholder interface{}) string {
	s := outfmt.SafeString(placeholder)
	if s == "" {
		return "-"
	}
	return outfmt.Truncate(s, 40)
}

// formatLinkCount extracts the link count from domain data.
// The API returns link count in _count.links nested structure.
func formatLinkCount(domain map[string]interface{}) string {
	// Try _count.links nested structure first
	if countObj, ok := domain["_count"].(map[string]interface{}); ok {
		if links, ok := countObj["links"]; ok {
			return formatClicks(outfmt.SafeInt(links))
		}
	}

	// Fallback to direct links field
	if links, ok := domain["links"]; ok {
		return formatClicks(outfmt.SafeInt(links))
	}

	return "0"
}

func newDomainsUpdateCmd() *cobra.Command {
	var (
		slug        string
		placeholder string
		expiredURL  string
		archived    bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a domain",
		Long:  "Update an existing domain configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if slug == "" {
				return fmt.Errorf("--slug is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{}
			if cmd.Flags().Changed("placeholder") {
				body["placeholder"] = placeholder
			}
			if cmd.Flags().Changed("expired-url") {
				body["expiredUrl"] = expiredURL
			}
			if cmd.Flags().Changed("archived") {
				body["archived"] = archived
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			resp, err := client.Patch(cmd.Context(), "/domains/"+url.PathEscape(slug), body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Domain name (required)")
	cmd.Flags().StringVar(&placeholder, "placeholder", "", "Placeholder URL for root domain")
	cmd.Flags().StringVar(&expiredURL, "expired-url", "", "URL for expired links")
	cmd.Flags().BoolVar(&archived, "archived", false, "Archive the domain")

	_ = cmd.MarkFlagRequired("slug")

	return cmd
}

func newDomainsDeleteCmd() *cobra.Command {
	var (
		slug   string
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a domain",
		Long:  "Delete a domain from your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if slug == "" {
				return fmt.Errorf("--slug is required")
			}

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Would delete domain with slug: %s\n", slug)
				return nil
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Delete(cmd.Context(), "/domains/"+url.PathEscape(slug))
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Domain name (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")

	_ = cmd.MarkFlagRequired("slug")

	return cmd
}

func newDomainsRegisterCmd() *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a domain",
		Long:  "Register a new domain through Dub.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if domain == "" {
				return fmt.Errorf("--domain is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"domain": domain,
			}

			resp, err := client.Post(cmd.Context(), "/domains/register", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Domain name to register (required)")

	_ = cmd.MarkFlagRequired("domain")

	return cmd
}

func newDomainsCheckCmd() *cobra.Command {
	var slug string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check domain status",
		Long:  "Check the configuration status of a domain.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if slug == "" {
				return fmt.Errorf("--slug is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Get(cmd.Context(), "/domains/"+url.PathEscape(slug)+"/status")
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Domain name (required)")

	_ = cmd.MarkFlagRequired("slug")

	return cmd
}
