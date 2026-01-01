// internal/cmd/partners.go
package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
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
		page      int
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
			if page > 0 {
				params.Set("page", fmt.Sprintf("%d", page))
			}

			resp, err := client.Get(cmd.Context(), "/partners?"+params.Encode())
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")

	_ = cmd.MarkFlagRequired("program-id")

	return cmd
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
		page      int
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
			if page > 0 {
				params.Set("page", fmt.Sprintf("%d", page))
			}

			resp, err := client.Get(cmd.Context(), "/partners/links?"+params.Encode())
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&programID, "program-id", "", "Program ID (required)")
	cmd.Flags().StringVar(&partnerID, "partner-id", "", "Filter by partner ID")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")

	_ = cmd.MarkFlagRequired("program-id")

	return cmd
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
