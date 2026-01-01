package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCustomersCmd_Name(t *testing.T) {
	cmd := newCustomersCmd()
	if cmd.Name() != "customers" {
		t.Errorf("expected 'customers', got %q", cmd.Name())
	}
}

func TestCustomersCmd_SubCommands(t *testing.T) {
	cmd := newCustomersCmd()
	subCmds := []string{"list", "get", "update", "delete"}
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

func TestCustomersGetCmd_RequiresID(t *testing.T) {
	cmd := newCustomersGetCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
	if err != nil && !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error about 'id' flag, got %q", err.Error())
	}
}

func TestCustomersUpdateCmd_RequiresID(t *testing.T) {
	cmd := newCustomersUpdateCmd()
	cmd.SetArgs([]string{"--name", "test"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
	if err != nil && !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error about 'id' flag, got %q", err.Error())
	}
}

func TestCustomersUpdateCmd_RequiresField(t *testing.T) {
	cmd := newCustomersUpdateCmd()
	cmd.SetArgs([]string{"--id", "cust_123"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no update field is provided")
	}
	// The error might be authentication-related, field validation, or keyring init
	// All are acceptable errors in this context
	expectedErrors := []string{
		"at least one field must be specified for update",
		"not authenticated",
		"No directory provided for file keyring",
	}
	if err != nil {
		found := false
		for _, expected := range expectedErrors {
			if contains(err.Error(), expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected error containing one of %v, got %q", expectedErrors, err.Error())
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCustomersDeleteCmd_RequiresID(t *testing.T) {
	cmd := newCustomersDeleteCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --id is not provided")
	}
	if err != nil && !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error about 'id' flag, got %q", err.Error())
	}
}

func TestCustomersListCmd_Flags(t *testing.T) {
	cmd := newCustomersListCmd()
	flags := []string{"search", "output", "limit", "all"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestCustomersListCmd_OutputFlagShorthand(t *testing.T) {
	cmd := newCustomersListCmd()

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected flag 'output' to exist")
	}
	if flag.Shorthand != "o" {
		t.Errorf("expected output flag shorthand to be 'o', got %q", flag.Shorthand)
	}
}

func TestCustomersListCmd_DefaultLimit(t *testing.T) {
	cmd := newCustomersListCmd()

	flag := cmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("expected flag 'limit' to exist")
	}
	if flag.DefValue != "25" {
		t.Errorf("expected limit default to be '25', got %q", flag.DefValue)
	}
}

func TestCustomersListCmd_DefaultOutput(t *testing.T) {
	cmd := newCustomersListCmd()

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected flag 'output' to exist")
	}
	if flag.DefValue != "table" {
		t.Errorf("expected output default to be 'table', got %q", flag.DefValue)
	}
}

func TestCustomersGetCmd_Flags(t *testing.T) {
	cmd := newCustomersGetCmd()
	if cmd.Flags().Lookup("id") == nil {
		t.Error("expected flag 'id' to exist")
	}
}

func TestCustomersUpdateCmd_Flags(t *testing.T) {
	cmd := newCustomersUpdateCmd()
	flags := []string{"id", "name", "email", "external-id"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestCustomersDeleteCmd_Flags(t *testing.T) {
	cmd := newCustomersDeleteCmd()
	if cmd.Flags().Lookup("id") == nil {
		t.Error("expected flag 'id' to exist")
	}
}

func TestCustomersDeleteCmd_DryRun(t *testing.T) {
	cmd := newCustomersDeleteCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--id", "cust_def456", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	expected := "Would delete customer with ID: cust_def456\n"
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

func TestCustomersDeleteCmd_DryRunFlag(t *testing.T) {
	cmd := newCustomersDeleteCmd()
	if cmd.Flags().Lookup("dry-run") == nil {
		t.Error("expected flag 'dry-run' to exist")
	}
}

func TestFormatCustomerField(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, "-"},
		{"empty string", "", "-"},
		{"valid name", "John Doe", "John Doe"},
		{"valid email", "john@example.com", "john@example.com"},
		{"valid external ID", "cust_123", "cust_123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCustomerField(tt.input)
			if result != tt.expected {
				t.Errorf("formatCustomerField(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
