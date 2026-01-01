// internal/cmd/analytics_test.go
package cmd

import (
	"testing"
)

func TestAnalyticsCmd_Name(t *testing.T) {
	cmd := newAnalyticsCmd()
	if cmd.Name() != "analytics" {
		t.Errorf("expected command name to be 'analytics', got %q", cmd.Name())
	}
}

func TestAnalyticsCmd_Flags(t *testing.T) {
	cmd := newAnalyticsCmd()

	flags := []string{
		"event",
		"group-by",
		"domain",
		"link-id",
		"interval",
		"start",
		"end",
		"country",
		"city",
		"device",
		"browser",
		"os",
		"referer",
		"timezone",
	}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestAnalyticsCmd_HasShortDescription(t *testing.T) {
	cmd := newAnalyticsCmd()
	if cmd.Short == "" {
		t.Error("expected command to have a short description")
	}
}

func TestAnalyticsCmd_HasLongDescription(t *testing.T) {
	cmd := newAnalyticsCmd()
	if cmd.Long == "" {
		t.Error("expected command to have a long description")
	}
}

func TestAnalyticsCmd_NoSubCommands(t *testing.T) {
	cmd := newAnalyticsCmd()
	if len(cmd.Commands()) != 0 {
		t.Errorf("expected analytics to have no subcommands, got %d", len(cmd.Commands()))
	}
}
