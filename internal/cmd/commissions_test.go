// internal/cmd/commissions_test.go
package cmd

import (
	"testing"
)

func TestCommissionsCmd_Name(t *testing.T) {
	cmd := newCommissionsCmd()
	if cmd.Name() != "commissions" {
		t.Errorf("expected 'commissions', got %q", cmd.Name())
	}
}

func TestCommissionsCmd_SubCommands(t *testing.T) {
	cmd := newCommissionsCmd()
	subCmds := []string{"list", "update"}
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

func TestCommissionsListCmd_RequiresProgramID(t *testing.T) {
	cmd := newCommissionsListCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --program-id is not provided")
	}
}

func TestCommissionsListCmd_Flags(t *testing.T) {
	cmd := newCommissionsListCmd()
	flags := []string{"program-id", "partner-id", "status", "page"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestCommissionsUpdateCmd_RequiresID(t *testing.T) {
	cmd := newCommissionsUpdateCmd()
	cmd.SetArgs([]string{"--status", "approved"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
}

func TestCommissionsUpdateCmd_RequiresStatusOrAmount(t *testing.T) {
	cmd := newCommissionsUpdateCmd()
	cmd.SetArgs([]string{"--id", "comm_123"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when neither --status nor --amount is provided")
	}
}

func TestCommissionsUpdateCmd_Flags(t *testing.T) {
	cmd := newCommissionsUpdateCmd()
	flags := []string{"id", "status", "amount"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}
