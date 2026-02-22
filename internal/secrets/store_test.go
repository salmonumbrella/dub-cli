// internal/secrets/store_test.go
package secrets

import (
	"testing"
	"time"
)

func TestCredentials_Fields(t *testing.T) {
	creds := Credentials{
		Name:      "test-workspace",
		APIKey:    "dub_test123",
		CreatedAt: time.Now(),
	}

	if creds.Name != "test-workspace" {
		t.Errorf("expected name 'test-workspace', got %q", creds.Name)
	}
	if creds.APIKey != "dub_test123" {
		t.Errorf("expected api key 'dub_test123', got %q", creds.APIKey)
	}
}

func TestParseCredentialKey(t *testing.T) {
	tests := []struct {
		key      string
		wantName string
		wantOK   bool
	}{
		{"workspace:production", "production", true},
		{"workspace:dev", "dev", true},
		{"other:key", "", false},
		{"workspace:", "", false},
	}

	for _, tt := range tests {
		name, ok := ParseCredentialKey(tt.key)
		if ok != tt.wantOK || name != tt.wantName {
			t.Errorf("ParseCredentialKey(%q) = (%q, %v), want (%q, %v)",
				tt.key, name, ok, tt.wantName, tt.wantOK)
		}
	}
}
