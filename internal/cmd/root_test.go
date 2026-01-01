// internal/cmd/root_test.go
package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand_Help(t *testing.T) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("dub")) {
		t.Errorf("expected output to contain 'dub', got: %s", output)
	}
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	cmd := NewRootCmd()

	// Check persistent flags exist
	flags := []string{"workspace", "output", "query", "yes", "debug", "limit", "sort-by", "desc"}
	for _, name := range flags {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("expected persistent flag %q to exist", name)
		}
	}
}
