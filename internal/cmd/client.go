// internal/cmd/client.go
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/salmonumbrella/dub-cli/internal/api"
	"github.com/salmonumbrella/dub-cli/internal/config"
	"github.com/salmonumbrella/dub-cli/internal/secrets"
)

// storeOpener allows injecting a mock store for testing
var storeOpener = func() (secrets.Store, error) {
	return secrets.OpenDefault()
}

// defaultWorkspaceGetter allows injecting a mock for testing
var defaultWorkspaceGetter = config.GetDefaultWorkspace

// getClient returns an API client using stored credentials.
// Credential resolution priority:
// 1. DUB_API_KEY environment variable (for CI/testing)
// 2. --workspace / -w flag (via context)
// 3. DUB_WORKSPACE environment variable (already folded into flag default)
// 4. Default workspace from config (set via `dub auth switch`)
// 5. If only one workspace configured, use it automatically
// 6. If multiple workspaces configured, return error asking user to specify
func getClient(ctx context.Context) (*api.Client, error) {
	// Check for API key environment variable first (useful for CI/testing)
	if apiKey := os.Getenv("DUB_API_KEY"); apiKey != "" {
		return api.NewClient(apiKey), nil
	}

	store, err := storeOpener()
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	return getClientWithStore(ctx, store)
}

// getClientWithStore is the core logic, separated for testing
func getClientWithStore(ctx context.Context, store secrets.Store) (*api.Client, error) {
	// Check for workspace flag (includes DUB_WORKSPACE via flag default)
	workspace := GetWorkspace(ctx)
	if workspace != "" {
		creds, err := store.Get(workspace)
		if err != nil {
			return nil, fmt.Errorf("workspace %q not found. Run: dub auth list", workspace)
		}
		return api.NewClient(creds.APIKey), nil
	}

	// Check for default workspace from config
	defaultWs, err := defaultWorkspaceGetter()
	if err == nil && defaultWs != "" {
		creds, err := store.Get(defaultWs)
		if err == nil {
			return api.NewClient(creds.APIKey), nil
		}
		// Default workspace no longer exists - continue to fallback logic
	}
	// If err != nil && !errors.Is(err, config.ErrNoDefaultWorkspace):
	// Unexpected error reading config - log but continue
	// (don't fail the command just because config is unreadable)

	// No workspace specified - use first available or error if multiple
	creds, err := store.List()
	if err != nil {
		return nil, err
	}

	switch len(creds) {
	case 0:
		return nil, fmt.Errorf("not authenticated. Run: dub auth login")
	case 1:
		return api.NewClient(creds[0].APIKey), nil
	default:
		names := make([]string, len(creds))
		for i, c := range creds {
			names[i] = c.Name
		}
		return nil, fmt.Errorf("multiple workspaces configured: %s\nSpecify with --workspace <name>, set DUB_WORKSPACE, or use: dub auth switch <name>", strings.Join(names, ", "))
	}
}
