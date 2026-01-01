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
	flags := []string{"program-id", "partner-id", "status", "output", "limit", "all"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestCommissionsListCmd_OutputShorthand(t *testing.T) {
	cmd := newCommissionsListCmd()
	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected 'output' flag to exist")
	}
	if flag.Shorthand != "o" {
		t.Errorf("expected 'output' shorthand to be 'o', got %q", flag.Shorthand)
	}
}

func TestCommissionsListCmd_DefaultValues(t *testing.T) {
	cmd := newCommissionsListCmd()

	// Check limit default
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Fatal("expected 'limit' flag to exist")
	}
	if limitFlag.DefValue != "25" {
		t.Errorf("expected 'limit' default to be '25', got %q", limitFlag.DefValue)
	}

	// Check output default
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("expected 'output' flag to exist")
	}
	if outputFlag.DefValue != "table" {
		t.Errorf("expected 'output' default to be 'table', got %q", outputFlag.DefValue)
	}

	// Check all default
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("expected 'all' flag to exist")
	}
	if allFlag.DefValue != "false" {
		t.Errorf("expected 'all' default to be 'false', got %q", allFlag.DefValue)
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "$0.00"},
		{1.00, "$1.00"},
		{12.34, "$12.34"},
		{123.45, "$123.45"},
		{1234.56, "$1,234.56"},
		{12345.67, "$12,345.67"},
		{123456.78, "$123,456.78"},
		{1234567.89, "$1,234,567.89"},
		{0.50, "$0.50"},
		{0.05, "$0.05"},
		{1000000.00, "$1,000,000.00"},
	}

	for _, tt := range tests {
		result := formatAmount(tt.input)
		if result != tt.expected {
			t.Errorf("formatAmount(%v): expected %q, got %q", tt.input, tt.expected, result)
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
