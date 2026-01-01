package cmd

import "testing"

func TestEventsCmd_Name(t *testing.T) {
	cmd := newEventsCmd()
	if cmd.Name() != "events" {
		t.Errorf("expected 'events', got %q", cmd.Name())
	}
}

func TestEventsCmd_SubCommands(t *testing.T) {
	cmd := newEventsCmd()
	subCmds := []string{"list"}
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

func TestEventsListCmd_Flags(t *testing.T) {
	cmd := newEventsListCmd()
	flags := []string{"event", "domain", "link-id", "interval", "start", "end", "country", "city", "device", "browser", "os", "referer", "page"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}
