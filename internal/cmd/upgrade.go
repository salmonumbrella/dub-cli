// internal/cmd/upgrade.go
package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const (
	// maxDownloadSize limits the maximum size of downloaded files (100MB)
	maxDownloadSize = 100 * 1024 * 1024
	// httpTimeout is the timeout for HTTP requests
	httpTimeout = 60 * time.Second
)

const (
	repoOwner = "salmonumbrella"
	repoName  = "dub-cli"
	githubAPI = "https://api.github.com"
)

// GitHubRelease represents a release from the GitHub API
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func newUpgradeCmd() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade dub CLI to the latest version",
		Long: `Check for and install the latest version of the Dub CLI.

This command fetches the latest release from GitHub and replaces the
current binary if a newer version is available.

Examples:
  dub upgrade          # Upgrade to latest version
  dub upgrade --check  # Only check for updates, don't install`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(cmd, checkOnly)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")

	return cmd
}

func runUpgrade(cmd *cobra.Command, checkOnly bool) error {
	currentVersion := normalizeVersion(Version)

	// dev builds can't be compared
	if currentVersion == "dev" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Cannot upgrade development builds. Please install from a release.")
		return nil
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", Version)

	// Fetch latest release
	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := normalizeVersion(release.TagName)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Latest version:  %s\n", release.TagName)

	// Compare versions
	cmp := semver.Compare(currentVersion, latestVersion)
	if cmp >= 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nYou are already running the latest version.")
		return nil
	}

	if checkOnly {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nUpdate available: %s -> %s\n", Version, release.TagName)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Run 'dub upgrade' to install.")
		return nil
	}

	// Find the appropriate asset for current OS/arch
	assetName := buildAssetName(release.TagName)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (looking for %s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nDownloading %s...\n", assetName)

	// Download and install
	if err := downloadAndInstall(downloadURL); err != nil {
		return fmt.Errorf("failed to upgrade: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully upgraded to %s\n", release.TagName)
	return nil
}

func fetchLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPI, repoOwner, repoName)

	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "dub-cli/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

// normalizeVersion ensures version string has "v" prefix for semver comparison
func normalizeVersion(version string) string {
	if version == "dev" || version == "" {
		return "dev"
	}
	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

// buildAssetName constructs the expected asset filename based on OS/arch
// GoReleaser naming convention: dub-cli_VERSION_OS_ARCH.tar.gz
func buildAssetName(version string) string {
	// Strip leading "v" from version for asset naming
	ver := strings.TrimPrefix(version, "v")

	// Use runtime.GOOS and runtime.GOARCH directly - GoReleaser uses the same naming
	return fmt.Sprintf("dub-cli_%s_%s_%s.tar.gz", ver, runtime.GOOS, runtime.GOARCH)
}

func downloadAndInstall(downloadURL string) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Download the archive with timeout and User-Agent
	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "dub-cli/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Limit response body size to prevent unbounded memory usage
	limitedReader := io.LimitReader(resp.Body, maxDownloadSize)

	// Create temp file in same directory as executable to avoid cross-filesystem rename issues
	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), "dub-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }() // Clean up temp file

	// Extract binary from tar.gz
	if err := extractBinary(limitedReader, tmpFile); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	_ = tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Replace old binary with new one
	// On Unix, we can rename over the existing file
	// On Windows, we need to rename the old file first
	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		_ = os.Remove(oldPath) // Remove any previous .old file
		if err := os.Rename(execPath, oldPath); err != nil {
			return fmt.Errorf("failed to move old binary: %w", err)
		}
		if err := os.Rename(tmpPath, execPath); err != nil {
			// Try to restore old binary
			_ = os.Rename(oldPath, execPath)
			return fmt.Errorf("failed to install new binary: %w", err)
		}
		_ = os.Remove(oldPath)
	} else {
		if err := os.Rename(tmpPath, execPath); err != nil {
			return fmt.Errorf("failed to replace binary: %w", err)
		}
	}

	return nil
}

func extractBinary(r io.Reader, dst *os.File) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)

	// Look for the dub binary in the archive
	binaryName := "dub"
	if runtime.GOOS == "windows" {
		binaryName = "dub.exe"
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("binary %q not found in archive", binaryName)
		}
		if err != nil {
			return err
		}

		// Skip directories and non-matching files
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Check if this is the binary we're looking for
		name := filepath.Base(header.Name)
		if name == binaryName {
			_, err := io.Copy(dst, tr)
			return err
		}
	}
}
