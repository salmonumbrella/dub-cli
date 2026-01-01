// internal/cmd/customers.go
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

func newCustomersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customers",
		Short: "Manage customers",
		Long:  "List, get, update, and delete customers.",
	}

	cmd.AddCommand(newCustomersListCmd())
	cmd.AddCommand(newCustomersGetCmd())
	cmd.AddCommand(newCustomersUpdateCmd())
	cmd.AddCommand(newCustomersDeleteCmd())

	return cmd
}

func newCustomersListCmd() *cobra.Command {
	var (
		search string
		output string
		limit  int
		all    bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customers",
		Long:  "List all customers in your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			if search != "" {
				params.Set("search", search)
			}

			path := "/customers"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}

			return handleCustomersListResponse(cmd, resp, output, limit, all)
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of customers to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all customers (ignore limit)")

	return cmd
}

func newCustomersGetCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a customer",
		Long:  "Get details of a specific customer.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Get(cmd.Context(), "/customers/"+url.PathEscape(id))
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Customer ID (required)")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newCustomersUpdateCmd() *cobra.Command {
	var (
		id         string
		name       string
		email      string
		externalID string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a customer",
		Long:  "Update an existing customer.",
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
			if cmd.Flags().Changed("email") {
				body["email"] = email
			}
			if cmd.Flags().Changed("external-id") {
				body["externalId"] = externalID
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			resp, err := client.Patch(cmd.Context(), "/customers/"+url.PathEscape(id), body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Customer ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "Customer name")
	cmd.Flags().StringVar(&email, "email", "", "Customer email")
	cmd.Flags().StringVar(&externalID, "external-id", "", "External customer ID")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newCustomersDeleteCmd() *cobra.Command {
	var (
		id     string
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a customer",
		Long:  "Delete a customer from your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required")
			}

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Would delete customer with ID: %s\n", id)
				return nil
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Delete(cmd.Context(), "/customers/"+url.PathEscape(id))
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Customer ID (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// handleCustomersListResponse handles the response for customers list command,
// formatting output as table or JSON based on the output flag.
func handleCustomersListResponse(cmd *cobra.Command, resp *http.Response, output string, limit int, all bool) error {
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

	// Parse customers for table output
	var customers []map[string]interface{}
	if err := json.Unmarshal(body, &customers); err != nil {
		return fmt.Errorf("failed to parse customers: %w", err)
	}

	totalCount := len(customers)

	// Apply limit unless --all is set
	displayLimit := limit
	if all {
		displayLimit = totalCount
	}
	if displayLimit > totalCount {
		displayLimit = totalCount
	}

	displayCustomers := customers[:displayLimit]

	// Define table columns
	columns := []outfmt.Column{
		{Name: "NAME", Width: 0, Align: outfmt.AlignLeft},
		{Name: "EMAIL", Width: 0, Align: outfmt.AlignLeft},
		{Name: "EXTERNAL ID", Width: 0, Align: outfmt.AlignLeft},
		{Name: "CREATED", Width: 0, Align: outfmt.AlignLeft},
	}

	// Build rows
	rows := make([][]string, len(displayCustomers))
	for i, customer := range displayCustomers {
		rows[i] = []string{
			formatCustomerField(customer["name"]),
			formatCustomerField(customer["email"]),
			formatCustomerField(customer["externalId"]),
			outfmt.FormatDate(customer["createdAt"]),
		}
	}

	// Write table
	if err := outfmt.FormatTable(cmd.OutOrStdout(), columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if displayLimit < totalCount {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d of %d customers. Use --limit or --all for more.\n", displayLimit, totalCount)
	}

	return nil
}

// formatCustomerField formats a customer field or returns "-" if not set.
func formatCustomerField(field interface{}) string {
	s := outfmt.SafeString(field)
	if s == "" {
		return "-"
	}
	return s
}
