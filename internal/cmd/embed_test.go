// internal/cmd/embed_test.go
package cmd

import (
	"testing"
)

func TestEmbedCmd_Name(t *testing.T) {
	cmd := newEmbedCmd()
	if cmd.Name() != "embed" {
		t.Errorf("expected 'embed', got %q", cmd.Name())
	}
}

func TestEmbedCmd_SubCommands(t *testing.T) {
	cmd := newEmbedCmd()
	subCmds := []string{"create-referral-token"}
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

func TestEmbedCreateReferralTokenCmd_RequiresProgramID(t *testing.T) {
	cmd := newEmbedCreateReferralTokenCmd()
	cmd.SetArgs([]string{"--partner-id", "ptr_123"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --program-id is not provided")
	}
}

func TestEmbedCreateReferralTokenCmd_RequiresPartnerID(t *testing.T) {
	cmd := newEmbedCreateReferralTokenCmd()
	cmd.SetArgs([]string{"--program-id", "prog_123"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --partner-id is not provided")
	}
}

func TestEmbedCreateReferralTokenCmd_Flags(t *testing.T) {
	cmd := newEmbedCreateReferralTokenCmd()
	flags := []string{"program-id", "partner-id"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}
