// internal/cmd/commissions.go
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

func newCommissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commissions",
		Short: "Manage commissions",
		Long:  "List and update partner commissions.",
	}

	cmd.AddCommand(newCommissionsListCmd())
	cmd.AddCommand(newCommissionsUpdateCmd())

	return cmd
}

func newCommissionsListCmd() *cobra.Command {
	var (
		programID string
		partnerID string
		status    string
		output    string
		limit     int
		all       bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commissions",
		Long:  "List all commissions for a program.",
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
			if status != "" {
				params.Set("status", status)
			}

			resp, err := client.Get(cmd.Context(), "/commissions?"+params.Encode())
			if err != nil {
				return err
			}

			return handleCommissionsListResponse(cmd, resp, output, limit, all)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&partnerID, "partner-id", "", "Filter by partner ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (pending, approved, paid)")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of commissions to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all commissions (ignore limit)")

	_ = cmd.MarkFlagRequired("program-id")

	return cmd
}

// handleCommissionsListResponse handles the response for commissions list command,
// formatting output as table or JSON based on the output flag.
func handleCommissionsListResponse(cmd *cobra.Command, resp *http.Response, output string, limit int, all bool) error {
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

	// Parse commissions for table output
	var commissions []map[string]interface{}
	if err := json.Unmarshal(body, &commissions); err != nil {
		return fmt.Errorf("failed to parse commissions: %w", err)
	}

	totalCount := len(commissions)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayCommissions := commissions[:displayLimit]

	// Define table columns
	columns := []outfmt.Column{
		{Name: "ID", Width: 20, Align: outfmt.AlignLeft},
		{Name: "Partner", Width: 30, Align: outfmt.AlignLeft},
		{Name: "Amount", Width: 0, Align: outfmt.AlignRight},
		{Name: "Status", Width: 0, Align: outfmt.AlignLeft},
		{Name: "Created", Width: 0, Align: outfmt.AlignLeft},
	}

	// Build rows
	rows := make([][]string, len(displayCommissions))
	for i, commission := range displayCommissions {
		rows[i] = []string{
			outfmt.Truncate(outfmt.SafeString(commission["id"]), 20),
			formatPartner(commission),
			formatAmount(outfmt.SafeFloat(commission["amount"])),
			outfmt.SafeString(commission["status"]),
			outfmt.FormatDate(commission["createdAt"]),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d commissions. Use --limit or --all for more.\n", displayLimit, totalCount)
	}

	return nil
}

// formatPartner extracts partner name or ID from commission data.
func formatPartner(commission map[string]interface{}) string {
	// Try nested partner object first
	if partner, ok := commission["partner"].(map[string]interface{}); ok {
		if name := outfmt.SafeString(partner["name"]); name != "" {
			return outfmt.Truncate(name, 30)
		}
		if id := outfmt.SafeString(partner["id"]); id != "" {
			return outfmt.Truncate(id, 30)
		}
	}

	// Fall back to partnerId field
	if partnerID := outfmt.SafeString(commission["partnerId"]); partnerID != "" {
		return outfmt.Truncate(partnerID, 30)
	}

	return "-"
}

// formatAmount formats a currency amount with $ and commas (e.g., 1234.50 -> "$1,234.50").
func formatAmount(amount float64) string {
	if amount == 0 {
		return "$0.00"
	}

	// Handle negative amounts
	negative := amount < 0
	if negative {
		amount = -amount
	}

	// Round to 2 decimal places to avoid floating-point precision issues
	cents := int(amount*100 + 0.5)
	whole := cents / 100
	cents = cents % 100

	wholeStr := formatWithCommas(whole)

	if negative {
		return fmt.Sprintf("-$%s.%02d", wholeStr, cents)
	}
	return fmt.Sprintf("$%s.%02d", wholeStr, cents)
}

// formatWithCommas adds comma separators to an integer.
func formatWithCommas(n int) string {
	if n == 0 {
		return "0"
	}

	str := fmt.Sprintf("%d", n)
	length := len(str)

	// Calculate number of commas needed
	commaCount := (length - 1) / 3
	if commaCount == 0 {
		return str
	}

	result := make([]byte, length+commaCount)
	resultIdx := len(result) - 1

	for i := length - 1; i >= 0; i-- {
		pos := length - 1 - i
		if pos > 0 && pos%3 == 0 {
			result[resultIdx] = ','
			resultIdx--
		}
		result[resultIdx] = str[i]
		resultIdx--
	}

	return string(result)
}

func newCommissionsUpdateCmd() *cobra.Command {
	var (
		id     string
		status string
		amount float64
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a commission",
		Long:  "Update the status or amount of a commission.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{}
			if cmd.Flags().Changed("status") {
				body["status"] = status
			}
			if cmd.Flags().Changed("amount") {
				body["amount"] = amount
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one of --status or --amount must be specified")
			}

			resp, err := client.Patch(cmd.Context(), "/commissions/"+url.PathEscape(id), body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Commission ID (required)")
	cmd.Flags().StringVar(&status, "status", "", "New status (pending, approved, paid)")
	cmd.Flags().Float64Var(&amount, "amount", 0, "Commission amount")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}
