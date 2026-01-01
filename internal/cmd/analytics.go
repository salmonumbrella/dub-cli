// internal/cmd/analytics.go
package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
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

			return handleResponse(cmd, resp)
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

	return cmd
}
