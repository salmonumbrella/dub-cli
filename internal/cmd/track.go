// internal/cmd/track.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTrackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track",
		Short: "Track conversions",
		Long:  "Track lead and sale conversion events.",
	}

	cmd.AddCommand(newTrackLeadCmd())
	cmd.AddCommand(newTrackSaleCmd())

	return cmd
}

func newTrackLeadCmd() *cobra.Command {
	var (
		clickID    string
		eventName  string
		externalID string
		customerID string
		name       string
		email      string
	)

	cmd := &cobra.Command{
		Use:   "lead",
		Short: "Track a lead",
		Long:  "Track a lead conversion event.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if clickID == "" {
				return fmt.Errorf("--click-id is required")
			}
			if eventName == "" {
				return fmt.Errorf("--event-name is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"clickId":   clickID,
				"eventName": eventName,
			}
			if externalID != "" {
				body["externalId"] = externalID
			}
			if customerID != "" {
				body["customerId"] = customerID
			}
			if name != "" {
				body["customerName"] = name
			}
			if email != "" {
				body["customerEmail"] = email
			}

			resp, err := client.Post(cmd.Context(), "/track/lead", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&clickID, "click-id", "", "Click ID from the link (required)")
	cmd.Flags().StringVar(&eventName, "event-name", "", "Name of the lead event (required)")
	cmd.Flags().StringVar(&externalID, "external-id", "", "External ID for the customer")
	cmd.Flags().StringVar(&customerID, "customer-id", "", "Customer ID")
	cmd.Flags().StringVar(&name, "name", "", "Customer name")
	cmd.Flags().StringVar(&email, "email", "", "Customer email")

	_ = cmd.MarkFlagRequired("click-id")
	_ = cmd.MarkFlagRequired("event-name")

	return cmd
}

func newTrackSaleCmd() *cobra.Command {
	var (
		clickID     string
		eventName   string
		externalID  string
		customerID  string
		amount      float64
		currency    string
		paymentProc string
		invoiceID   string
	)

	cmd := &cobra.Command{
		Use:   "sale",
		Short: "Track a sale",
		Long:  "Track a sale conversion event.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if clickID == "" {
				return fmt.Errorf("--click-id is required")
			}
			if eventName == "" {
				return fmt.Errorf("--event-name is required")
			}
			if amount <= 0 {
				return fmt.Errorf("--amount is required and must be positive")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"clickId":   clickID,
				"eventName": eventName,
				"amount":    amount,
			}
			if externalID != "" {
				body["externalId"] = externalID
			}
			if customerID != "" {
				body["customerId"] = customerID
			}
			if currency != "" {
				body["currency"] = currency
			}
			if paymentProc != "" {
				body["paymentProcessor"] = paymentProc
			}
			if invoiceID != "" {
				body["invoiceId"] = invoiceID
			}

			resp, err := client.Post(cmd.Context(), "/track/sale", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&clickID, "click-id", "", "Click ID from the link (required)")
	cmd.Flags().StringVar(&eventName, "event-name", "", "Name of the sale event (required)")
	cmd.Flags().Float64Var(&amount, "amount", 0, "Sale amount in cents (required)")
	cmd.Flags().StringVar(&externalID, "external-id", "", "External ID for the customer")
	cmd.Flags().StringVar(&customerID, "customer-id", "", "Customer ID")
	cmd.Flags().StringVar(&currency, "currency", "", "Currency code (default: USD)")
	cmd.Flags().StringVar(&paymentProc, "payment-processor", "", "Payment processor (stripe, shopify, etc)")
	cmd.Flags().StringVar(&invoiceID, "invoice-id", "", "Invoice ID")

	_ = cmd.MarkFlagRequired("click-id")
	_ = cmd.MarkFlagRequired("event-name")
	// Note: --amount has custom validation (must be positive) so we keep manual check in RunE

	return cmd
}
