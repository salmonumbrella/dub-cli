// internal/cmd/embed.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newEmbedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "embed",
		Short: "Manage embed tokens",
		Long:  "Create embed tokens for widgets.",
	}

	cmd.AddCommand(newEmbedCreateReferralTokenCmd())

	return cmd
}

func newEmbedCreateReferralTokenCmd() *cobra.Command {
	var (
		programID string
		partnerID string
	)

	cmd := &cobra.Command{
		Use:   "create-referral-token",
		Short: "Create a referral embed token",
		Long:  "Create an embed token for referral widgets.",
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

			resp, err := client.Post(cmd.Context(), "/embed-tokens/referrals", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Partner program ID (required)")
	cmd.Flags().StringVar(&partnerID, "partner-id", "", "Partner ID (required)")

	_ = cmd.MarkFlagRequired("program-id")
	_ = cmd.MarkFlagRequired("partner-id")

	return cmd
}
