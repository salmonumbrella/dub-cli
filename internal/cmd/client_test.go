// internal/cmd/client_test.go
package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/dub-cli/internal/secrets"
)

// mockStore implements secrets.Store for testing
type mockStore struct {
	creds map[string]secrets.Credentials
}

func newMockStore() *mockStore {
	return &mockStore{creds: make(map[string]secrets.Credentials)}
}

func (m *mockStore) Keys() ([]string, error) {
	var keys []string
	for k := range m.creds {
		keys = append(keys, "workspace:"+k)
	}
	return keys, nil
}

func (m *mockStore) Set(name string, creds secrets.Credentials) error {
	m.creds[name] = creds
	return nil
}

func (m *mockStore) Get(name string) (secrets.Credentials, error) {
	if creds, ok := m.creds[name]; ok {
		return creds, nil
	}
	return secrets.Credentials{}, errors.New("not found")
}

func (m *mockStore) Delete(name string) error {
	delete(m.creds, name)
	return nil
}

func (m *mockStore) List() ([]secrets.Credentials, error) {
	var out []secrets.Credentials
	for _, c := range m.creds {
		out = append(out, c)
	}
	return out, nil
}

func TestGetClientWithStore_NoCredentials(t *testing.T) {
	store := newMockStore()
	ctx := context.Background()

	_, err := getClientWithStore(ctx, store)
	if err == nil {
		t.Fatal("expected error for no credentials")
	}
	if got := err.Error(); got != "not authenticated. Run: dub auth login" {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestGetClientWithStore_SingleWorkspace(t *testing.T) {
	store := newMockStore()
	_ = store.Set("production", secrets.Credentials{
		Name:      "production",
		APIKey:    "dub_prod123",
		CreatedAt: time.Now(),
	})

	ctx := context.Background()
	client, err := getClientWithStore(ctx, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
}

func TestGetClientWithStore_MultipleWorkspaces_NoSelection(t *testing.T) {
	store := newMockStore()
	_ = store.Set("production", secrets.Credentials{
		Name:      "production",
		APIKey:    "dub_prod123",
		CreatedAt: time.Now(),
	})
	_ = store.Set("staging", secrets.Credentials{
		Name:      "staging",
		APIKey:    "dub_staging456",
		CreatedAt: time.Now(),
	})

	// Ensure no default workspace is set
	origGetter := defaultWorkspaceGetter
	defaultWorkspaceGetter = func() (string, error) {
		return "", errors.New("no default workspace configured")
	}
	defer func() { defaultWorkspaceGetter = origGetter }()

	ctx := context.Background()
	_, err := getClientWithStore(ctx, store)
	if err == nil {
		t.Fatal("expected error for multiple workspaces without selection")
	}
	errMsg := err.Error()
	if !containsAll(errMsg, "multiple workspaces", "--workspace", "DUB_WORKSPACE", "dub auth switch") {
		t.Errorf("error message should guide user, got: %s", errMsg)
	}
}

func TestGetClientWithStore_MultipleWorkspaces_WithSelection(t *testing.T) {
	store := newMockStore()
	_ = store.Set("production", secrets.Credentials{
		Name:      "production",
		APIKey:    "dub_prod123",
		CreatedAt: time.Now(),
	})
	_ = store.Set("staging", secrets.Credentials{
		Name:      "staging",
		APIKey:    "dub_staging456",
		CreatedAt: time.Now(),
	})

	// Create context with workspace selection
	ctx := context.WithValue(context.Background(), workspaceKey, "production")

	client, err := getClientWithStore(ctx, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	// Verify the correct workspace's API key is used
	if got := client.APIKey(); got != "dub_prod123" {
		t.Errorf("expected production API key, got: %s", got)
	}
}

func TestGetClientWithStore_WorkspaceNotFound(t *testing.T) {
	store := newMockStore()
	_ = store.Set("production", secrets.Credentials{
		Name:      "production",
		APIKey:    "dub_prod123",
		CreatedAt: time.Now(),
	})

	// Request non-existent workspace
	ctx := context.WithValue(context.Background(), workspaceKey, "nonexistent")

	_, err := getClientWithStore(ctx, store)
	if err == nil {
		t.Fatal("expected error for non-existent workspace")
	}
	errMsg := err.Error()
	if !containsAll(errMsg, "nonexistent", "not found", "dub auth list") {
		t.Errorf("error message should be helpful, got: %s", errMsg)
	}
}

func TestGetClientWithStore_ExplicitWorkspace_OverridesDefault(t *testing.T) {
	store := newMockStore()
	_ = store.Set("production", secrets.Credentials{
		Name:      "production",
		APIKey:    "dub_prod123",
		CreatedAt: time.Now(),
	})
	_ = store.Set("staging", secrets.Credentials{
		Name:      "staging",
		APIKey:    "dub_staging456",
		CreatedAt: time.Now(),
	})

	// Even with multiple workspaces, explicit selection should work
	ctx := context.WithValue(context.Background(), workspaceKey, "staging")

	client, err := getClientWithStore(ctx, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	// Verify the staging workspace's API key is used (not production)
	if got := client.APIKey(); got != "dub_staging456" {
		t.Errorf("expected staging API key, got: %s", got)
	}
}

func TestGetClientWithStore_DefaultWorkspace(t *testing.T) {
	store := newMockStore()
	_ = store.Set("production", secrets.Credentials{
		Name:      "production",
		APIKey:    "dub_prod123",
		CreatedAt: time.Now(),
	})
	_ = store.Set("staging", secrets.Credentials{
		Name:      "staging",
		APIKey:    "dub_staging456",
		CreatedAt: time.Now(),
	})

	// Set up mock default workspace getter
	origGetter := defaultWorkspaceGetter
	defaultWorkspaceGetter = func() (string, error) {
		return "staging", nil
	}
	defer func() { defaultWorkspaceGetter = origGetter }()

	// Without explicit workspace, should use default
	ctx := context.Background()
	client, err := getClientWithStore(ctx, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	// Verify the default workspace's API key is used
	if got := client.APIKey(); got != "dub_staging456" {
		t.Errorf("expected staging API key from default, got: %s", got)
	}
}

func TestGetClientWithStore_ExplicitOverridesDefault(t *testing.T) {
	store := newMockStore()
	_ = store.Set("production", secrets.Credentials{
		Name:      "production",
		APIKey:    "dub_prod123",
		CreatedAt: time.Now(),
	})
	_ = store.Set("staging", secrets.Credentials{
		Name:      "staging",
		APIKey:    "dub_staging456",
		CreatedAt: time.Now(),
	})

	// Set up mock default workspace getter
	origGetter := defaultWorkspaceGetter
	defaultWorkspaceGetter = func() (string, error) {
		return "staging", nil
	}
	defer func() { defaultWorkspaceGetter = origGetter }()

	// Explicit workspace should override default
	ctx := context.WithValue(context.Background(), workspaceKey, "production")
	client, err := getClientWithStore(ctx, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := client.APIKey(); got != "dub_prod123" {
		t.Errorf("expected production API key (explicit), got: %s", got)
	}
}

func TestGetClientWithStore_DefaultWorkspaceNotFound(t *testing.T) {
	store := newMockStore()
	_ = store.Set("production", secrets.Credentials{
		Name:      "production",
		APIKey:    "dub_prod123",
		CreatedAt: time.Now(),
	})

	// Set up mock default workspace getter pointing to non-existent workspace
	origGetter := defaultWorkspaceGetter
	defaultWorkspaceGetter = func() (string, error) {
		return "deleted-workspace", nil
	}
	defer func() { defaultWorkspaceGetter = origGetter }()

	// Should fall back to single workspace
	ctx := context.Background()
	client, err := getClientWithStore(ctx, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := client.APIKey(); got != "dub_prod123" {
		t.Errorf("expected fallback to production, got: %s", got)
	}
}

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
