// internal/cmd/qr_test.go
package cmd

import (
	"testing"
)

func TestQRCmd_Name(t *testing.T) {
	cmd := newQRCmd()
	if cmd.Name() != "qr" {
		t.Errorf("expected command name to be 'qr', got %q", cmd.Name())
	}
}

func TestQRCmd_RequiresURL(t *testing.T) {
	cmd := newQRCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --url is not provided")
	}
}

func TestQRCmd_Flags(t *testing.T) {
	cmd := newQRCmd()

	tests := []struct {
		name string
		flag string
	}{
		{name: "url flag", flag: "url"},
		{name: "size flag", flag: "size"},
		{name: "level flag", flag: "level"},
		{name: "fg-color flag", flag: "fg-color"},
		{name: "bg-color flag", flag: "bg-color"},
		{name: "output flag", flag: "output"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if cmd.Flags().Lookup(tt.flag) == nil {
				t.Errorf("expected flag %q to exist", tt.flag)
			}
		})
	}
}

func TestQRCmd_URLFlagIsRequired(t *testing.T) {
	cmd := newQRCmd()

	flag := cmd.Flags().Lookup("url")
	if flag == nil {
		t.Fatal("expected --url flag to exist")
	}

	// Check that flag has usage text indicating it's required
	if flag.Usage == "" {
		t.Error("expected --url flag to have usage text")
	}
}

func TestQRCmd_SizeFlagIsInt(t *testing.T) {
	cmd := newQRCmd()

	flag := cmd.Flags().Lookup("size")
	if flag == nil {
		t.Fatal("expected --size flag to exist")
	}

	// Verify flag type is int
	if flag.Value.Type() != "int" {
		t.Errorf("expected --size flag to be int type, got %s", flag.Value.Type())
	}
}

func TestQRCmd_OutputFlagShorthand(t *testing.T) {
	cmd := newQRCmd()

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected --output flag to exist")
	}

	// Check that output flag has shorthand -O
	if flag.Shorthand != "O" {
		t.Errorf("expected --output flag to have shorthand 'O', got %q", flag.Shorthand)
	}
}
