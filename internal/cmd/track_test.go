// internal/cmd/track_test.go
package cmd

import (
	"strings"
	"testing"
)

func TestTrackCmd_Name(t *testing.T) {
	cmd := newTrackCmd()
	if cmd.Name() != "track" {
		t.Errorf("expected 'track', got %q", cmd.Name())
	}
}

func TestTrackCmd_SubCommands(t *testing.T) {
	cmd := newTrackCmd()
	subCmds := []string{"lead", "sale"}
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

func TestTrackLeadCmd_RequiresClickID(t *testing.T) {
	cmd := newTrackLeadCmd()
	cmd.SetArgs([]string{"--event-name", "signup"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --click-id is not provided")
	}
	if err != nil && !strings.Contains(err.Error(), "click-id") {
		t.Errorf("expected error about 'click-id' flag, got %q", err.Error())
	}
}

func TestTrackLeadCmd_RequiresEventName(t *testing.T) {
	cmd := newTrackLeadCmd()
	cmd.SetArgs([]string{"--click-id", "click_123"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --event-name is not provided")
	}
	if err != nil && !strings.Contains(err.Error(), "event-name") {
		t.Errorf("expected error about 'event-name' flag, got %q", err.Error())
	}
}

func TestTrackSaleCmd_RequiresClickID(t *testing.T) {
	cmd := newTrackSaleCmd()
	cmd.SetArgs([]string{"--event-name", "purchase", "--amount", "100"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --click-id is not provided")
	}
	if err != nil && !strings.Contains(err.Error(), "click-id") {
		t.Errorf("expected error about 'click-id' flag, got %q", err.Error())
	}
}

func TestTrackSaleCmd_RequiresEventName(t *testing.T) {
	cmd := newTrackSaleCmd()
	cmd.SetArgs([]string{"--click-id", "click_123", "--amount", "100"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --event-name is not provided")
	}
	if err != nil && !strings.Contains(err.Error(), "event-name") {
		t.Errorf("expected error about 'event-name' flag, got %q", err.Error())
	}
}

func TestTrackSaleCmd_RequiresPositiveAmount(t *testing.T) {
	cmd := newTrackSaleCmd()
	cmd.SetArgs([]string{"--click-id", "click_123", "--event-name", "purchase", "--amount", "0"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --amount is not positive")
	}
	if err != nil && err.Error() != "--amount is required and must be positive" {
		t.Errorf("expected '--amount is required and must be positive', got %q", err.Error())
	}
}

func TestTrackSaleCmd_RequiresAmountWhenMissing(t *testing.T) {
	cmd := newTrackSaleCmd()
	cmd.SetArgs([]string{"--click-id", "click_123", "--event-name", "purchase"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --amount is not provided")
	}
	if err != nil && err.Error() != "--amount is required and must be positive" {
		t.Errorf("expected '--amount is required and must be positive', got %q", err.Error())
	}
}

func TestTrackLeadCmd_Flags(t *testing.T) {
	cmd := newTrackLeadCmd()
	flags := []string{"click-id", "event-name", "external-id", "customer-id", "name", "email"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestTrackSaleCmd_Flags(t *testing.T) {
	cmd := newTrackSaleCmd()
	flags := []string{"click-id", "event-name", "amount", "external-id", "customer-id", "currency", "payment-processor", "invoice-id"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}
