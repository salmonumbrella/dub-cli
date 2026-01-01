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
	flags := []string{"search", "output", "limit", "all"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestFoldersListCmd_OutputFlagShorthand(t *testing.T) {
	cmd := newFoldersListCmd()

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected flag 'output' to exist")
	}
	if flag.Shorthand != "o" {
		t.Errorf("expected output flag shorthand to be 'o', got %q", flag.Shorthand)
	}
}

func TestFoldersListCmd_DefaultLimit(t *testing.T) {
	cmd := newFoldersListCmd()

	flag := cmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("expected flag 'limit' to exist")
	}
	if flag.DefValue != "25" {
		t.Errorf("expected limit default to be '25', got %q", flag.DefValue)
	}
}

func TestFoldersListCmd_DefaultOutput(t *testing.T) {
	cmd := newFoldersListCmd()

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected flag 'output' to exist")
	}
	if flag.DefValue != "table" {
		t.Errorf("expected output default to be 'table', got %q", flag.DefValue)
	}
}

func TestFormatFolderType(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, "-"},
		{"empty string", "", "-"},
		{"default type", "default", "default"},
		{"custom type", "project", "project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFolderType(tt.input)
			if result != tt.expected {
				t.Errorf("formatFolderType(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatAccessLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, "-"},
		{"empty string", "", "-"},
		{"write access", "write", "write"},
		{"read access", "read", "read"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAccessLevel(tt.input)
			if result != tt.expected {
				t.Errorf("formatAccessLevel(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatFolderLinkCount(t *testing.T) {
	tests := []struct {
		name     string
		folder   map[string]interface{}
		expected string
	}{
		{
			name:     "no links field",
			folder:   map[string]interface{}{"name": "Marketing"},
			expected: "0",
		},
		{
			name:     "links in _count.links",
			folder:   map[string]interface{}{"_count": map[string]interface{}{"links": float64(45)}},
			expected: "45",
		},
		{
			name:     "links as direct field",
			folder:   map[string]interface{}{"links": float64(12)},
			expected: "12",
		},
		{
			name:     "links with comma formatting",
			folder:   map[string]interface{}{"_count": map[string]interface{}{"links": float64(1234)}},
			expected: "1,234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFolderLinkCount(tt.folder)
			if result != tt.expected {
				t.Errorf("formatFolderLinkCount() = %q, want %q", result, tt.expected)
			}
		})
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
