// internal/cmd/partners.go
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

func newPartnersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "partners",
		Short: "Manage partners",
		Long:  "Create, list, and manage affiliate partners.",
	}

	cmd.AddCommand(newPartnersCreateCmd())
	cmd.AddCommand(newPartnersListCmd())
	cmd.AddCommand(newPartnersBanCmd())
	cmd.AddCommand(newPartnersLinksCmd())
	cmd.AddCommand(newPartnersAnalyticsCmd())

	return cmd
}

func newPartnersCreateCmd() *cobra.Command {
	var (
		programID string
		name      string
		email     string
		image     string
		country   string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a partner",
		Long:  "Create a new affiliate partner.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if programID == "" {
				return fmt.Errorf("--program-id is required")
			}
			if email == "" {
				return fmt.Errorf("--email is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"programId": programID,
				"email":     email,
			}
			if name != "" {
				body["name"] = name
			}
			if image != "" {
				body["image"] = image
			}
			if country != "" {
				body["country"] = country
			}

			resp, err := client.Post(cmd.Context(), "/partners", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "Partner name")
	cmd.Flags().StringVar(&email, "email", "", "Partner email (required)")
	cmd.Flags().StringVar(&image, "image", "", "Partner image URL")
	cmd.Flags().StringVar(&country, "country", "", "Partner country code")

	_ = cmd.MarkFlagRequired("program-id")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}

func newPartnersListCmd() *cobra.Command {
	var (
		programID string
		search    string
		status    string
		output    string
		limit     int
		all       bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List partners",
		Long:  "List all partners in a program.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if programID == "" {
				return fmt.Errorf("--program-id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			params.Set("programId", programID)
			if search != "" {
				params.Set("search", search)
			}
			if status != "" {
				params.Set("status", status)
			}

			resp, err := client.Get(cmd.Context(), "/partners?"+params.Encode())
			if err != nil {
				return err
			}

			return handlePartnersListResponse(cmd, resp, output, limit, all)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of partners to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all partners (ignore limit)")

	_ = cmd.MarkFlagRequired("program-id")

	return cmd
}

// handlePartnersListResponse handles the response for partners list command,
// formatting output as table or JSON based on the output flag.
func handlePartnersListResponse(cmd *cobra.Command, resp *http.Response, output string, limit int, all bool) error {
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

	// Parse partners for table output
	var partners []map[string]interface{}
	if err := json.Unmarshal(body, &partners); err != nil {
		return fmt.Errorf("failed to parse partners: %w", err)
	}

	totalCount := len(partners)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayPartners := partners[:displayLimit]

	// Define table columns
	columns := []outfmt.Column{
		{Name: "Name", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Email", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Status", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Country", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Created", Width: 0, Align: outfmt.AlignLeft},
	}

	// Build rows
	rows := make([][]string, len(displayPartners))
	for i, partner := range displayPartners {
		rows[i] = []string{
			formatPartnerName(partner["name"]),
			outfmt.SafeString(partner["email"]),
			formatPartnerStatus(partner["status"]),
			formatPartnerCountry(partner["country"]),
			outfmt.FormatDate(partner["createdAt"]),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d partners. Use --limit or --all for more.\n", displayLimit, totalCount)
	}

	return nil
}

// formatPartnerName formats the partner name or returns "-" if not set.
func formatPartnerName(name interface{}) string {
	s := outfmt.SafeString(name)
	if s == "" {
		return "-"
	}
	return s
}

// formatPartnerStatus formats the partner status.
func formatPartnerStatus(status interface{}) string {
	s := outfmt.SafeString(status)
	if s == "" {
		return "-"
	}
	return s
}

// formatPartnerCountry formats the partner country code or returns "-" if not set.
func formatPartnerCountry(country interface{}) string {
	s := outfmt.SafeString(country)
	if s == "" {
		return "-"
	}
	return s
}

func newPartnersBanCmd() *cobra.Command {
	var (
		programID string
		partnerID string
		reason    string
	)

	cmd := &cobra.Command{
		Use:   "ban",
		Short: "Ban a partner",
		Long:  "Ban a partner from a program.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if programID == "" {
				return fmt.Errorf("--program-id is required")
			}
			if partnerID == "" {
				return fmt.Errorf("--partner-id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"programId": programID,
				"partnerId": partnerID,
			}
			if reason != "" {
				body["reason"] = reason
			}

			resp, err := client.Post(cmd.Context(), "/partners/ban", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&partnerID, "partner-id", "", "Partner ID (required)")
	cmd.Flags().StringVar(&reason, "reason", "", "Reason for banning")

	_ = cmd.MarkFlagRequired("program-id")
	_ = cmd.MarkFlagRequired("partner-id")

	return cmd
}

func newPartnersLinksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "links",
		Short: "Manage partner links",
		Long:  "Create and list partner referral links.",
	}

	cmd.AddCommand(newPartnersLinksCreateCmd())
	cmd.AddCommand(newPartnersLinksUpsertCmd())
	cmd.AddCommand(newPartnersLinksListCmd())

	return cmd
}

func newPartnersLinksCreateCmd() *cobra.Command {
	var (
		programID string
		partnerID string
		linkURL   string
		key       string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a partner link",
		Long:  "Create a new referral link for a partner.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if programID == "" {
				return fmt.Errorf("--program-id is required")
			}
			if partnerID == "" {
				return fmt.Errorf("--partner-id is required")
			}
			if linkURL == "" {
				return fmt.Errorf("--url is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"programId": programID,
				"partnerId": partnerID,
				"url":       linkURL,
			}
			if key != "" {
				body["key"] = key
			}

			resp, err := client.Post(cmd.Context(), "/partners/links", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&partnerID, "partner-id", "", "Partner ID (required)")
	cmd.Flags().StringVar(&linkURL, "url", "", "Destination URL (required)")
	cmd.Flags().StringVar(&key, "key", "", "Custom short key")

	_ = cmd.MarkFlagRequired("program-id")
	_ = cmd.MarkFlagRequired("partner-id")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func newPartnersLinksUpsertCmd() *cobra.Command {
	var (
		programID string
		partnerID string
		linkURL   string
		key       string
	)

	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "Create or update a partner link",
		Long:  "Create or update a referral link for a partner.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if programID == "" {
				return fmt.Errorf("--program-id is required")
			}
			if partnerID == "" {
				return fmt.Errorf("--partner-id is required")
			}
			if linkURL == "" {
				return fmt.Errorf("--url is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"programId": programID,
				"partnerId": partnerID,
				"url":       linkURL,
			}
			if key != "" {
				body["key"] = key
			}

			resp, err := client.Put(cmd.Context(), "/partners/links/upsert", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&partnerID, "partner-id", "", "Partner ID (required)")
	cmd.Flags().StringVar(&linkURL, "url", "", "Destination URL (required)")
	cmd.Flags().StringVar(&key, "key", "", "Custom short key")

	_ = cmd.MarkFlagRequired("program-id")
	_ = cmd.MarkFlagRequired("partner-id")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func newPartnersLinksListCmd() *cobra.Command {
	var (
		programID string
		partnerID string
		output    string
		limit     int
		all       bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List partner links",
		Long:  "List all referral links for a partner.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if programID == "" {
				return fmt.Errorf("--program-id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			params.Set("programId", programID)
			if partnerID != "" {
				params.Set("partnerId", partnerID)
			}

			resp, err := client.Get(cmd.Context(), "/partners/links?"+params.Encode())
			if err != nil {
				return err
			}

			return handlePartnersLinksListResponse(cmd, resp, output, limit, all)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&partnerID, "partner-id", "", "Filter by partner ID")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of links to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all links (ignore limit)")

	_ = cmd.MarkFlagRequired("program-id")

	return cmd
}

// handlePartnersLinksListResponse handles the response for partners links list command,
// formatting output as table or JSON based on the output flag.
func handlePartnersLinksListResponse(cmd *cobra.Command, resp *http.Response, output string, limit int, all bool) error {
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
	var links []map[string]interface{}
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
		{Name: "URL", Width: 50, Align: outfmt.AlignLeft},
		{Name: "Clicks", Width: 0, Align: outfmt.AlignRight},
		{Name: "Created", Width: 0, Align: outfmt.AlignLeft},
	}

	// Build rows
	rows := make([][]string, len(displayLinks))
	for i, link := range displayLinks {
		rows[i] = []string{
			buildShortLink(outfmt.SafeString(link["domain"]), outfmt.SafeString(link["key"])),
			outfmt.Truncate(outfmt.SafeString(link["url"]), 50),
			formatClicks(outfmt.SafeInt(link["clicks"])),
			outfmt.FormatDate(link["createdAt"]),
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

func newPartnersAnalyticsCmd() *cobra.Command {
	var (
		programID string
		partnerID string
		interval  string
		start     string
		end       string
		groupBy   string
	)

	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "Get partner analytics",
		Long:  "Retrieve analytics for partners.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if programID == "" {
				return fmt.Errorf("--program-id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			params.Set("programId", programID)
			if partnerID != "" {
				params.Set("partnerId", partnerID)
			}
			if interval != "" {
				params.Set("interval", interval)
			}
			if start != "" {
				params.Set("start", start)
			}
			if end != "" {
				params.Set("end", end)
			}
			if groupBy != "" {
				params.Set("groupBy", groupBy)
			}

			resp, err := client.Get(cmd.Context(), "/partners/analytics?"+params.Encode())
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&partnerID, "partner-id", "", "Filter by partner ID")
	cmd.Flags().StringVar(&interval, "interval", "", "Time interval")
	cmd.Flags().StringVar(&start, "start", "", "Start date (ISO 8601)")
	cmd.Flags().StringVar(&end, "end", "", "End date (ISO 8601)")
	cmd.Flags().StringVar(&groupBy, "group-by", "", "Property to group by")

	_ = cmd.MarkFlagRequired("program-id")

	return cmd
}
