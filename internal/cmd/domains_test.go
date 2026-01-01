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

	flags := []string{"archived", "search", "page"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
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
