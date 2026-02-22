// internal/cmd/domains_test.go
package cmd

import (
	"bytes"
	"testing"
)

func TestDomainsCmd_SubCommands(t *testing.T) {
	cmd := newDomainsCmd()

	subCmds := []string{"create", "list", "update", "delete", "register", "check"}
	for _, name := range subCmds {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q to exist", name)
		}
	}
}

func TestDomainsCmd_Name(t *testing.T) {
	cmd := newDomainsCmd()
	if cmd.Name() != "domains" {
		t.Errorf("expected command name to be 'domains', got %q", cmd.Name())
	}
}

func TestDomainsCreateCmd_RequiresSlug(t *testing.T) {
	cmd := newDomainsCreateCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --slug is not provided")
	}
}

func TestDomainsCreateCmd_Flags(t *testing.T) {
	cmd := newDomainsCreateCmd()

	flags := []string{"slug", "placeholder", "expired-url", "archived"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestDomainsUpdateCmd_RequiresSlug(t *testing.T) {
	cmd := newDomainsUpdateCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --slug is not provided")
	}
}

func TestDomainsUpdateCmd_Flags(t *testing.T) {
	cmd := newDomainsUpdateCmd()

	flags := []string{"slug", "placeholder", "expired-url", "archived"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestDomainsDeleteCmd_RequiresSlug(t *testing.T) {
	cmd := newDomainsDeleteCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --slug is not provided")
	}
}

func TestDomainsRegisterCmd_RequiresDomain(t *testing.T) {
	cmd := newDomainsRegisterCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --domain is not provided")
	}
}

func TestDomainsCheckCmd_RequiresSlug(t *testing.T) {
	cmd := newDomainsCheckCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --slug is not provided")
	}
}

func TestDomainsListCmd_Flags(t *testing.T) {
	cmd := newDomainsListCmd()

	flags := []string{"archived", "search", "output", "limit", "all"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestDomainsListCmd_OutputFlagShorthand(t *testing.T) {
	cmd := newDomainsListCmd()

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected flag 'output' to exist")
	}
	if flag.Shorthand != "o" {
		t.Errorf("expected output flag shorthand to be 'o', got %q", flag.Shorthand)
	}
}

func TestDomainsListCmd_DefaultLimit(t *testing.T) {
	cmd := newDomainsListCmd()

	flag := cmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("expected flag 'limit' to exist")
	}
	if flag.DefValue != "25" {
		t.Errorf("expected limit default to be '25', got %q", flag.DefValue)
	}
}

func TestDomainsListCmd_DefaultOutput(t *testing.T) {
	cmd := newDomainsListCmd()

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected flag 'output' to exist")
	}
	if flag.DefValue != "table" {
		t.Errorf("expected output default to be 'table', got %q", flag.DefValue)
	}
}

func TestFormatPlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, "-"},
		{"empty string", "", "-"},
		{"short URL", "https://dub.co", "https://dub.co"},
		{"long URL truncated", "https://example.com/very/long/path/that/exceeds/forty/characters/limit", "https://example.com/very/long/path/th..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPlaceholder(tt.input)
			if result != tt.expected {
				t.Errorf("formatPlaceholder(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatLinkCount(t *testing.T) {
	tests := []struct {
		name     string
		domain   map[string]interface{}
		expected string
	}{
		{
			name:     "no links field",
			domain:   map[string]interface{}{"slug": "example.com"},
			expected: "0",
		},
		{
			name:     "links in _count.links",
			domain:   map[string]interface{}{"_count": map[string]interface{}{"links": float64(142)}},
			expected: "142",
		},
		{
			name:     "links as direct field",
			domain:   map[string]interface{}{"links": float64(5)},
			expected: "5",
		},
		{
			name:     "links with comma formatting",
			domain:   map[string]interface{}{"_count": map[string]interface{}{"links": float64(1234)}},
			expected: "1,234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatLinkCount(tt.domain)
			if result != tt.expected {
				t.Errorf("formatLinkCount() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDomainsDeleteCmd_DryRun(t *testing.T) {
	cmd := newDomainsDeleteCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--slug", "example.com", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	expected := "Would delete domain with slug: example.com\n"
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

func TestDomainsDeleteCmd_DryRunFlag(t *testing.T) {
	cmd := newDomainsDeleteCmd()
	if cmd.Flags().Lookup("dry-run") == nil {
		t.Error("expected flag 'dry-run' to exist")
	}
}
