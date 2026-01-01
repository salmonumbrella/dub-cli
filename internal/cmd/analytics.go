// internal/cmd/analytics.go
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/dub-cli/internal/api"
	"github.com/salmonumbrella/dub-cli/internal/outfmt"
)

func newAnalyticsCmd() *cobra.Command {
	var (
		event    string
		groupBy  string
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
		timezone string
		output   string
		limit    int
		all      bool
	)

	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "Retrieve analytics",
		Long:  "Retrieve analytics for links, including clicks, leads, and sales.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			if event != "" {
				params.Set("event", event)
			}
			if groupBy != "" {
				params.Set("groupBy", groupBy)
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
			if timezone != "" {
				params.Set("timezone", timezone)
			}

			path := "/analytics"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}

			return handleAnalyticsResponse(cmd, resp, groupBy, output, limit, all)
		},
	}

	cmd.Flags().StringVar(&event, "event", "", "Event type: clicks, leads, or sales")
	cmd.Flags().StringVar(&groupBy, "group-by", "", "Property to group by: count, timeseries, countries, cities, devices, browsers, os, referers")
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
	cmd.Flags().StringVar(&timezone, "timezone", "", "Timezone for results")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of rows to show (for grouped results)")
	cmd.Flags().BoolVar(&all, "all", false, "Show all rows (ignore limit)")

	return cmd
}

// handleAnalyticsResponse handles the response for analytics command,
// formatting output as table or JSON based on the output flag and group-by value.
func handleAnalyticsResponse(cmd *cobra.Command, resp *http.Response, groupBy, output string, limit int, all bool) error {
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

	// Determine table format based on group-by value
	switch groupBy {
	case "", "count":
		return formatAnalyticsCount(cmd, body)
	case "timeseries":
		return formatAnalyticsTimeseries(cmd, body, limit, all)
	case "countries", "cities", "devices", "browsers", "os", "referers":
		return formatAnalyticsGrouped(cmd, body, groupBy, limit, all)
	default:
		// Unknown group-by, fall back to JSON
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
			return nil
		}
		return outfmt.FormatJSON(cmd.OutOrStdout(), data, "")
	}
}

// formatAnalyticsCount formats simple count/stats output as a vertical table.
func formatAnalyticsCount(cmd *cobra.Command, body []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// If not an object, print raw
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
		return nil
	}

	// Define table columns for vertical layout
	columns := []outfmt.Column{
		{Name: "Metric", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Value", Width: 0, Align: outfmt.AlignRight},
	}

	// Build rows from the data
	rows := [][]string{}
	metricOrder := []string{"clicks", "leads", "sales", "saleAmount"}
	metricLabels := map[string]string{
		"clicks":     "Clicks",
		"leads":      "Leads",
		"sales":      "Sales",
		"saleAmount": "Sale Amount",
	}

	for _, key := range metricOrder {
		if val, ok := data[key]; ok {
			label := metricLabels[key]
			rows = append(rows, []string{label, formatMetricValue(val)})
		}
	}

	// Add any other fields not in metricOrder
	for key, val := range data {
		if _, found := metricLabels[key]; !found {
			label := strings.Title(key) //nolint:staticcheck // strings.Title is fine for simple capitalization
			rows = append(rows, []string{label, formatMetricValue(val)})
		}
	}

	return outfmt.FormatTable(cmd.OutOrStdout(), columns, rows)
}

// formatAnalyticsTimeseries formats timeseries data as a table with date column.
func formatAnalyticsTimeseries(cmd *cobra.Command, body []byte, limit int, all bool) error {
	var data []map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
		return nil
	}

	totalCount := len(data)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayData := data[:displayLimit]

	// Define table columns
	columns := []outfmt.Column{
		{Name: "Date", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Clicks", Width: 0, Align: outfmt.AlignRight},
		{Name: "Leads", Width: 0, Align: outfmt.AlignRight},
		{Name: "Sales", Width: 0, Align: outfmt.AlignRight},
	}

	// Build rows
	rows := make([][]string, len(displayData))
	for i, item := range displayData {
		rows[i] = []string{
			outfmt.FormatDate(item["start"]),
			formatMetricValue(item["clicks"]),
			formatMetricValue(item["leads"]),
			formatMetricValue(item["sales"]),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d dates. Use --limit or --all for more.\n", displayLimit, totalCount)
	}

	return nil
}

// formatAnalyticsGrouped formats grouped analytics data (countries, cities, etc.).
func formatAnalyticsGrouped(cmd *cobra.Command, body []byte, groupBy string, limit int, all bool) error {
	var data []map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
		return nil
	}

	totalCount := len(data)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayData := data[:displayLimit]

	// Get column name and key based on group-by type
	columnName, dataKey := getGroupByColumn(groupBy)

	// Define table columns
	columns := []outfmt.Column{
		{Name: columnName, Width: 0, Align: outfmt.AlignLeft},
		{Name: "Clicks", Width: 0, Align: outfmt.AlignRight},
		{Name: "Leads", Width: 0, Align: outfmt.AlignRight},
		{Name: "Sales", Width: 0, Align: outfmt.AlignRight},
	}

	// Build rows
	rows := make([][]string, len(displayData))
	for i, item := range displayData {
		rows[i] = []string{
			outfmt.SafeString(item[dataKey]),
			formatMetricValue(item["clicks"]),
			formatMetricValue(item["leads"]),
			formatMetricValue(item["sales"]),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		noun := getGroupByNoun(groupBy)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d %s. Use --limit or --all for more.\n", displayLimit, totalCount, noun)
	}

	return nil
}

// getGroupByColumn returns the column header name and data key for a group-by type.
func getGroupByColumn(groupBy string) (columnName, dataKey string) {
	switch groupBy {
	case "countries":
		return "Country", "country"
	case "cities":
		return "City", "city"
	case "devices":
		return "Device", "device"
	case "browsers":
		return "Browser", "browser"
	case "os":
		return "OS", "os"
	case "referers":
		return "Referer", "referer"
	default:
		return "Value", groupBy
	}
}

// getGroupByNoun returns the plural noun for pagination message.
func getGroupByNoun(groupBy string) string {
	switch groupBy {
	case "countries":
		return "countries"
	case "cities":
		return "cities"
	case "devices":
		return "devices"
	case "browsers":
		return "browsers"
	case "os":
		return "operating systems"
	case "referers":
		return "referers"
	default:
		return "items"
	}
}

// formatMetricValue formats a numeric value with comma separators.
func formatMetricValue(val interface{}) string {
	n := outfmt.SafeInt(val)
	return formatClicks(n)
}
