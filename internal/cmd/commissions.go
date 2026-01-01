// internal/cmd/commissions.go
package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
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
		page      int
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
			if page > 0 {
				params.Set("page", fmt.Sprintf("%d", page))
			}

			resp, err := client.Get(cmd.Context(), "/commissions?"+params.Encode())
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&partnerID, "partner-id", "", "Filter by partner ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (pending, approved, paid)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")

	_ = cmd.MarkFlagRequired("program-id")

	return cmd
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
