// internal/cmd/partners_test.go
package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestPartnersCmd_SubCommands(t *testing.T) {
	cmd := newPartnersCmd()

	subCmds := []string{"create", "list", "ban", "links", "analytics"}
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

func TestPartnersCmd_Name(t *testing.T) {
	cmd := newPartnersCmd()
	if cmd.Name() != "partners" {
		t.Errorf("expected command name to be 'partners', got %q", cmd.Name())
	}
}

func TestPartnersLinksCmd_SubCommands(t *testing.T) {
	partnersCmd := newPartnersCmd()

	var linksCmd *cobra.Command
	for _, sub := range partnersCmd.Commands() {
		if sub.Name() == "links" {
			linksCmd = sub
			break
		}
	}

	if linksCmd == nil {
		t.Fatal("expected links subcommand to exist")
	}

	subCmds := []string{"create", "upsert", "list"}
	for _, name := range subCmds {
		found := false
		for _, sub := range linksCmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected links subcommand %q to exist", name)
		}
	}
}

func TestPartnersCreateCmd_RequiresProgramID(t *testing.T) {
	cmd := newPartnersCreateCmd()
	cmd.SetArgs([]string{"--email", "test@example.com"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --program-id is not provided")
	}
}

func TestPartnersCreateCmd_RequiresEmail(t *testing.T) {
	cmd := newPartnersCreateCmd()
	cmd.SetArgs([]string{"--program-id", "prog_123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --email is not provided")
	}
}

func TestPartnersCreateCmd_Flags(t *testing.T) {
	cmd := newPartnersCreateCmd()

	flags := []string{"program-id", "name", "email", "image", "country"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestPartnersListCmd_RequiresProgramID(t *testing.T) {
	cmd := newPartnersListCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --program-id is not provided")
	}
}

func TestPartnersListCmd_Flags(t *testing.T) {
	cmd := newPartnersListCmd()

	flags := []string{"program-id", "search", "status", "page"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestPartnersBanCmd_RequiresProgramID(t *testing.T) {
	cmd := newPartnersBanCmd()
	cmd.SetArgs([]string{"--partner-id", "ptr_123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --program-id is not provided")
	}
}

func TestPartnersBanCmd_RequiresPartnerID(t *testing.T) {
	cmd := newPartnersBanCmd()
	cmd.SetArgs([]string{"--program-id", "prog_123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --partner-id is not provided")
	}
}

func TestPartnersLinksCreateCmd_RequiresProgramID(t *testing.T) {
	cmd := newPartnersLinksCreateCmd()
	cmd.SetArgs([]string{"--partner-id", "ptr_123", "--url", "https://example.com"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --program-id is not provided")
	}
}

func TestPartnersLinksCreateCmd_RequiresPartnerID(t *testing.T) {
	cmd := newPartnersLinksCreateCmd()
	cmd.SetArgs([]string{"--program-id", "prog_123", "--url", "https://example.com"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --partner-id is not provided")
	}
}

func TestPartnersLinksCreateCmd_RequiresURL(t *testing.T) {
	cmd := newPartnersLinksCreateCmd()
	cmd.SetArgs([]string{"--program-id", "prog_123", "--partner-id", "ptr_123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --url is not provided")
	}
}

func TestPartnersAnalyticsCmd_RequiresProgramID(t *testing.T) {
	cmd := newPartnersAnalyticsCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --program-id is not provided")
	}
}

func TestPartnersAnalyticsCmd_Flags(t *testing.T) {
	cmd := newPartnersAnalyticsCmd()

	flags := []string{"program-id", "partner-id", "interval", "start", "end", "group-by"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}
