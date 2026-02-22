// internal/cmd/auth_test.go
package cmd

import (
	"testing"
)

func TestAuthCmd_SubCommands(t *testing.T) {
	cmd := newAuthCmd()

	subCmds := []string{"login", "logout", "list", "switch", "status"}
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
