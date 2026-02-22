// internal/cmd/events.go
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/dub-cli/internal/api"
	"github.com/salmonumbrella/dub-cli/internal/outfmt"
)

func newEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Manage events",
		Long:  "List and manage click, lead, and sale events.",
	}

	cmd.AddCommand(newEventsListCmd())

	return cmd
}

func newEventsListCmd() *cobra.Command {
	var (
		event    string
		domain   string
		linkID   string
		interval string
		start    string
		end      string
		country  string
		city     string
		device   string
		browser  string
		os       string
		referer  string
		output   string
		limit    int
		all      bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events",
		Long:  "List click, lead, and sale events.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			if event != "" {
				params.Set("event", event)
			}
			if domain != "" {
				params.Set("domain", domain)
			}
			if linkID != "" {
				params.Set("linkId", linkID)
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
			if country != "" {
				params.Set("country", country)
			}
			if city != "" {
				params.Set("city", city)
			}
			if device != "" {
				params.Set("device", device)
			}
			if browser != "" {
				params.Set("browser", browser)
			}
			if os != "" {
				params.Set("os", os)
			}
			if referer != "" {
				params.Set("referer", referer)
			}

			path := "/events"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}

			return handleEventsListResponse(cmd, resp, output, limit, all)
		},
	}

	cmd.Flags().StringVar(&event, "event", "", "Event type: clicks, leads, or sales")
	cmd.Flags().StringVar(&domain, "domain", "", "Filter by domain")
	cmd.Flags().StringVar(&linkID, "link-id", "", "Filter by link ID")
	cmd.Flags().StringVar(&interval, "interval", "", "Time interval: 1h, 24h, 7d, 30d, 90d, all")
	cmd.Flags().StringVar(&start, "start", "", "Start date (ISO 8601)")
	cmd.Flags().StringVar(&end, "end", "", "End date (ISO 8601)")
	cmd.Flags().StringVar(&country, "country", "", "Filter by country code")
	cmd.Flags().StringVar(&city, "city", "", "Filter by city")
	cmd.Flags().StringVar(&device, "device", "", "Filter by device type")
	cmd.Flags().StringVar(&browser, "browser", "", "Filter by browser")
	cmd.Flags().StringVar(&os, "os", "", "Filter by operating system")
	cmd.Flags().StringVar(&referer, "referer", "", "Filter by referer")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of events to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all events (ignore limit)")

	return cmd
}

// handleEventsListResponse handles the response for events list command,
// formatting output as table or JSON based on the output flag.
func handleEventsListResponse(cmd *cobra.Command, resp *http.Response, output string, limit int, all bool) error {
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

	// Parse events for table output
	var events []map[string]interface{}
	if err := json.Unmarshal(body, &events); err != nil {
		return fmt.Errorf("failed to parse events: %w", err)
	}

	totalCount := len(events)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayEvents := events[:displayLimit]

	// Define table columns
	columns := []outfmt.Column{
		{Name: "Timestamp", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Event", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Link", Width: 20, Align: outfmt.AlignLeft},
		{Name: "Country", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Device", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Browser", Width: 0, Align: outfmt.AlignLeft},
	}

	// Build rows
	rows := make([][]string, len(displayEvents))
	for i, event := range displayEvents {
		rows[i] = []string{
			formatTimestamp(event["timestamp"]),
			outfmt.SafeString(event["event"]),
			formatEventLink(event),
			formatEventField(event["country"]),
			formatEventField(event["device"]),
			formatEventField(event["browser"]),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d events. Use --limit or --all for more.\n", displayLimit, totalCount)
	}

	return nil
}

// formatTimestamp formats an ISO timestamp to "Jan 15, 3:42 PM" format.
func formatTimestamp(ts interface{}) string {
	s := outfmt.SafeString(ts)
	if s == "" {
		return "-"
	}

	// Try parsing RFC3339 format
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try RFC3339Nano
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			// Return original string if parsing fails
			return s
		}
	}

	return t.Format("Jan 2, 3:04 PM")
}

// formatEventLink extracts and formats the link from event data.
// Returns the short link or truncated link ID.
func formatEventLink(event map[string]interface{}) string {
	// Try to get the link object
	if linkObj, ok := event["link"].(map[string]interface{}); ok {
		// Try shortLink first
		if shortLink := outfmt.SafeString(linkObj["shortLink"]); shortLink != "" {
			return outfmt.Truncate(shortLink, 20)
		}
		// Fall back to domain/key
		domain := outfmt.SafeString(linkObj["domain"])
		key := outfmt.SafeString(linkObj["key"])
		if domain != "" && key != "" {
			return outfmt.Truncate(domain+"/"+key, 20)
		}
		// Fall back to link ID
		if id := outfmt.SafeString(linkObj["id"]); id != "" {
			return outfmt.Truncate(id, 20)
		}
	}

	// Fall back to linkId at top level
	if linkID := outfmt.SafeString(event["linkId"]); linkID != "" {
		return outfmt.Truncate(linkID, 20)
	}

	return "-"
}

// formatEventField formats an event field value, returning "-" for empty/nil values.
func formatEventField(v interface{}) string {
	s := outfmt.SafeString(v)
	if s == "" {
		return "-"
	}
	return s
}
