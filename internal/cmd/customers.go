// internal/cmd/customers.go
package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
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
		page   int
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
			if page > 0 {
				params.Set("page", fmt.Sprintf("%d", page))
			}

			path := "/customers"
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
	cmd.Flags().IntVar(&page, "page", 0, "Page number")

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
