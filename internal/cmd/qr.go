// internal/cmd/qr.go
package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/dub-cli/internal/api"
)

func newQRCmd() *cobra.Command {
	var (
		qrURL   string
		size    int
		level   string
		fgColor string
		bgColor string
		output  string
	)

	cmd := &cobra.Command{
		Use:   "qr",
		Short: "Generate a QR code",
		Long:  "Generate a QR code for a URL. The output is a PNG image.",
		Example: `  # Generate QR code and print to stdout
  dub qr --url https://example.com

  # Save QR code to a file
  dub qr --url https://example.com --output qr.png

  # Customize QR code appearance
  dub qr --url https://example.com --size 800 --level H --fg-color 000000 --bg-color ffffff`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if qrURL == "" {
				return fmt.Errorf("--url is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			params.Set("url", qrURL)

			if size > 0 {
				params.Set("size", fmt.Sprintf("%d", size))
			}
			if level != "" {
				params.Set("level", level)
			}
			if fgColor != "" {
				params.Set("fgColor", fgColor)
			}
			if bgColor != "" {
				params.Set("bgColor", bgColor)
			}

			path := "/qr?" + params.Encode()

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			if resp.StatusCode >= 400 {
				apiErr := api.ParseAPIError(body)
				return fmt.Errorf("%s", apiErr.Error())
			}

			// Write to file or stdout
			if output != "" {
				if err := os.WriteFile(output, body, 0o644); err != nil {
					return fmt.Errorf("failed to write file: %w", err)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "QR code saved to %s\n", output)
			} else {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), string(body))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&qrURL, "url", "", "URL to generate QR code for (required)")
	cmd.Flags().IntVar(&size, "size", 0, "Size of the QR code in pixels (default 600)")
	cmd.Flags().StringVar(&level, "level", "", "Error correction level: L, M, Q, H")
	cmd.Flags().StringVar(&fgColor, "fg-color", "", "Foreground color (hex, e.g., 000000)")
	cmd.Flags().StringVar(&bgColor, "bg-color", "", "Background color (hex, e.g., ffffff)")
	cmd.Flags().StringVarP(&output, "output", "O", "", "Output file path (default: stdout)")

	_ = cmd.MarkFlagRequired("url")

	return cmd
}
