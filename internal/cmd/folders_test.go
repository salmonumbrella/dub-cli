// internal/cmd/folders_test.go
package cmd

import (
	"bytes"
	"testing"
)

func TestFoldersCmd_Name(t *testing.T) {
	cmd := newFoldersCmd()
	if cmd.Name() != "folders" {
		t.Errorf("expected 'folders', got %q", cmd.Name())
	}
}

func TestFoldersCmd_SubCommands(t *testing.T) {
	cmd := newFoldersCmd()
	subCmds := []string{"create", "list", "update", "delete"}
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

func TestFoldersCreateCmd_RequiresName(t *testing.T) {
	cmd := newFoldersCreateCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --name is not provided")
	}
}

func TestFoldersCreateCmd_Flags(t *testing.T) {
	cmd := newFoldersCreateCmd()
	flags := []string{"name", "parent-id"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestFoldersListCmd_Flags(t *testing.T) {
	cmd := newFoldersListCmd()
	flags := []string{"search", "page"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestFoldersUpdateCmd_RequiresID(t *testing.T) {
	cmd := newFoldersUpdateCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
}

func TestFoldersUpdateCmd_RequiresNameOrParentID(t *testing.T) {
	cmd := newFoldersUpdateCmd()
	cmd.SetArgs([]string{"--id", "fld_123"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when neither --name nor --parent-id is provided")
	}
}

func TestFoldersUpdateCmd_Flags(t *testing.T) {
	cmd := newFoldersUpdateCmd()
	flags := []string{"id", "name", "parent-id"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestFoldersDeleteCmd_RequiresID(t *testing.T) {
	cmd := newFoldersDeleteCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
}

func TestFoldersDeleteCmd_Flags(t *testing.T) {
	cmd := newFoldersDeleteCmd()
	if cmd.Flags().Lookup("id") == nil {
		t.Error("expected flag 'id' to exist")
	}
}

func TestFoldersDeleteCmd_DryRun(t *testing.T) {
	cmd := newFoldersDeleteCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--id", "fld_xyz789", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	expected := "Would delete folder with ID: fld_xyz789\n"
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

func TestFoldersDeleteCmd_DryRunFlag(t *testing.T) {
	cmd := newFoldersDeleteCmd()
	if cmd.Flags().Lookup("dry-run") == nil {
		t.Error("expected flag 'dry-run' to exist")
	}
}
