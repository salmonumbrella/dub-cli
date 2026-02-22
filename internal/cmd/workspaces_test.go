// internal/cmd/workspaces_test.go
package cmd

import "testing"

func TestWorkspacesCmd_Name(t *testing.T) {
	cmd := newWorkspacesCmd()
	if cmd.Name() != "workspaces" {
		t.Errorf("expected command name to be 'workspaces', got %q", cmd.Name())
	}
}

func TestWorkspacesCmd_SubCommands(t *testing.T) {
	cmd := newWorkspacesCmd()

	subCmds := []string{"get", "update"}
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

func TestWorkspacesGetCmd_RequiresID(t *testing.T) {
	cmd := newWorkspacesGetCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
}

func TestWorkspacesGetCmd_Flags(t *testing.T) {
	cmd := newWorkspacesGetCmd()

	flags := []string{"id"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestWorkspacesUpdateCmd_RequiresID(t *testing.T) {
	cmd := newWorkspacesUpdateCmd()
	cmd.SetArgs([]string{"--name", "test"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
}

func TestWorkspacesUpdateCmd_RequiresNameOrSlug(t *testing.T) {
	cmd := newWorkspacesUpdateCmd()
	cmd.SetArgs([]string{"--id", "ws_123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when neither --name nor --slug is provided")
	}
}

func TestWorkspacesUpdateCmd_Flags(t *testing.T) {
	cmd := newWorkspacesUpdateCmd()

	flags := []string{"id", "name", "slug"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}
