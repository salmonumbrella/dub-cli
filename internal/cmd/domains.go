// internal/cmd/domains.go
package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

func newDomainsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domains",
		Short: "Manage domains",
		Long:  "Create, list, update, and delete custom domains.",
	}

	cmd.AddCommand(newDomainsCreateCmd())
	cmd.AddCommand(newDomainsListCmd())
	cmd.AddCommand(newDomainsUpdateCmd())
	cmd.AddCommand(newDomainsDeleteCmd())
	cmd.AddCommand(newDomainsRegisterCmd())
	cmd.AddCommand(newDomainsCheckCmd())

	return cmd
}

func newDomainsCreateCmd() *cobra.Command {
	var (
		slug        string
		placeholder string
		expiredURL  string
		archived    bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a domain",
		Long:  "Add a custom domain to your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if slug == "" {
				return fmt.Errorf("--slug is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"slug": slug,
			}
			if placeholder != "" {
				body["placeholder"] = placeholder
			}
			if expiredURL != "" {
				body["expiredUrl"] = expiredURL
			}
			if archived {
				body["archived"] = archived
			}

			resp, err := client.Post(cmd.Context(), "/domains", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Domain name (required)")
	cmd.Flags().StringVar(&placeholder, "placeholder", "", "Placeholder URL for root domain")
	cmd.Flags().StringVar(&expiredURL, "expired-url", "", "URL for expired links")
	cmd.Flags().BoolVar(&archived, "archived", false, "Archive the domain")

	_ = cmd.MarkFlagRequired("slug")

	return cmd
}

func newDomainsListCmd() *cobra.Command {
	var (
		archived bool
		search   string
		page     int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
		Long:  "List all domains in your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			if archived {
				params.Set("archived", "true")
			}
			if search != "" {
				params.Set("search", search)
			}
			if page > 0 {
				params.Set("page", fmt.Sprintf("%d", page))
			}

			path := "/domains"
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

	cmd.Flags().BoolVar(&archived, "archived", false, "Include archived domains")
	cmd.Flags().StringVar(&search, "search", "", "Search query")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")

	return cmd
}

func newDomainsUpdateCmd() *cobra.Command {
	var (
		slug        string
		placeholder string
		expiredURL  string
		archived    bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a domain",
		Long:  "Update an existing domain configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if slug == "" {
				return fmt.Errorf("--slug is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{}
			if cmd.Flags().Changed("placeholder") {
				body["placeholder"] = placeholder
			}
			if cmd.Flags().Changed("expired-url") {
				body["expiredUrl"] = expiredURL
			}
			if cmd.Flags().Changed("archived") {
				body["archived"] = archived
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			resp, err := client.Patch(cmd.Context(), "/domains/"+url.PathEscape(slug), body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Domain name (required)")
	cmd.Flags().StringVar(&placeholder, "placeholder", "", "Placeholder URL for root domain")
	cmd.Flags().StringVar(&expiredURL, "expired-url", "", "URL for expired links")
	cmd.Flags().BoolVar(&archived, "archived", false, "Archive the domain")

	_ = cmd.MarkFlagRequired("slug")

	return cmd
}

func newDomainsDeleteCmd() *cobra.Command {
	var (
		slug   string
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a domain",
		Long:  "Delete a domain from your workspace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if slug == "" {
				return fmt.Errorf("--slug is required")
			}

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Would delete domain with slug: %s\n", slug)
				return nil
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Delete(cmd.Context(), "/domains/"+url.PathEscape(slug))
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Domain name (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")

	_ = cmd.MarkFlagRequired("slug")

	return cmd
}

func newDomainsRegisterCmd() *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a domain",
		Long:  "Register a new domain through Dub.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if domain == "" {
				return fmt.Errorf("--domain is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"domain": domain,
			}

			resp, err := client.Post(cmd.Context(), "/domains/register", body)
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Domain name to register (required)")

	_ = cmd.MarkFlagRequired("domain")

	return cmd
}

func newDomainsCheckCmd() *cobra.Command {
	var slug string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check domain status",
		Long:  "Check the configuration status of a domain.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if slug == "" {
				return fmt.Errorf("--slug is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Get(cmd.Context(), "/domains/"+url.PathEscape(slug)+"/status")
			if err != nil {
				return err
			}

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Domain name (required)")

	_ = cmd.MarkFlagRequired("slug")

	return cmd
}
