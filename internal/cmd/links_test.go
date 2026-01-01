// internal/cmd/links_test.go
package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestLinksCmd_SubCommands(t *testing.T) {
	cmd := newLinksCmd()

	subCmds := []string{"create", "list", "get", "count", "update", "upsert", "delete", "bulk"}
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

func TestLinksBulkCmd_SubCommands(t *testing.T) {
	linksCmd := newLinksCmd()

	var bulkCmd *cobra.Command
	for _, sub := range linksCmd.Commands() {
		if sub.Name() == "bulk" {
			bulkCmd = sub
			break
		}
	}

	if bulkCmd == nil {
		t.Fatal("expected bulk subcommand to exist")
	}

	subCmds := []string{"create", "update", "delete"}
	for _, name := range subCmds {
		found := false
		for _, sub := range bulkCmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected bulk subcommand %q to exist", name)
		}
	}
}

func TestLinksCreateCmd_RequiresURL(t *testing.T) {
	cmd := newLinksCreateCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --url is not provided")
	}
}

func TestLinksGetCmd_RequiresIDOrDomainKey(t *testing.T) {
	cmd := newLinksGetCmd()
	cmd.SetArgs([]string{})

	// Mock stdin to avoid blocking
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when neither --id nor --domain/--key are provided")
	}
}

func TestLinksUpdateCmd_RequiresID(t *testing.T) {
	cmd := newLinksUpdateCmd()
	cmd.SetArgs([]string{"--url", "https://example.com"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
}

func TestLinksDeleteCmd_RequiresID(t *testing.T) {
	cmd := newLinksDeleteCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
}

func TestLinksUpsertCmd_RequiresURL(t *testing.T) {
	cmd := newLinksUpsertCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --url is not provided")
	}
}

func TestLinksDeleteCmd_DryRun(t *testing.T) {
	cmd := newLinksDeleteCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--id", "link_abc123", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	expected := "Would delete link with ID: link_abc123\n"
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

func TestLinksDeleteCmd_DryRunFlag(t *testing.T) {
	cmd := newLinksDeleteCmd()
	if cmd.Flags().Lookup("dry-run") == nil {
		t.Error("expected flag 'dry-run' to exist")
	}
}
