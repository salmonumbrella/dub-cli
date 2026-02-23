// internal/cmd/links_test.go
package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
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

func TestLinksUpdateCmd_RequiresIDOrDomainKey(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no identifier flags",
			args:    []string{"--url", "https://example.com"},
			wantErr: true,
		},
		{
			name:    "domain without key",
			args:    []string{"--domain", "dub.sh", "--url", "https://example.com"},
			wantErr: true,
		},
		{
			name:    "key without domain",
			args:    []string{"--key", "my-link", "--url", "https://example.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newLinksUpdateCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
		})
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

func TestLinksListCmd_Flags(t *testing.T) {
	cmd := newLinksListCmd()

	tests := []struct {
		name         string
		defaultValue string
	}{
		{"output", "table"},
		{"limit", "25"},
		{"all", "false"},
		{"search", ""},
		{"domain", ""},
	}

	for _, tt := range tests {
		flag := cmd.Flags().Lookup(tt.name)
		if flag == nil {
			t.Errorf("expected flag %q to exist", tt.name)
			continue
		}
		if flag.DefValue != tt.defaultValue {
			t.Errorf("flag %q: expected default %q, got %q", tt.name, tt.defaultValue, flag.DefValue)
		}
	}

	// Check short flag for output
	outputFlag := cmd.Flags().ShorthandLookup("o")
	if outputFlag == nil {
		t.Error("expected short flag 'o' for output")
	}
}

func TestFormatClicks(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{12, "12"},
		{123, "123"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{12345678, "12,345,678"},
		{123456789, "123,456,789"},
		{1000, "1,000"},
		{1000000, "1,000,000"},
	}

	for _, tt := range tests {
		result := formatClicks(tt.input)
		if result != tt.expected {
			t.Errorf("formatClicks(%d): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}

func TestFormatLastClicked(t *testing.T) {
	tests := []struct {
		input    *string
		expected string
	}{
		{nil, "-"},
		{strPtr(""), "-"},
		{strPtr("2024-01-15T10:30:00Z"), "Jan 15, 2024"},
		{strPtr("2023-12-10T14:45:00.000Z"), "Dec 10, 2023"},
		{strPtr("2024-07-04T00:00:00+00:00"), "Jul 4, 2024"},
		{strPtr("invalid-date"), "-"},
	}

	for _, tt := range tests {
		result := formatLastClicked(tt.input)
		if result != tt.expected {
			input := "nil"
			if tt.input != nil {
				input = *tt.input
			}
			t.Errorf("formatLastClicked(%q): expected %q, got %q", input, tt.expected, result)
		}
	}
}

func TestBuildShortLink(t *testing.T) {
	tests := []struct {
		domain   string
		key      string
		expected string
	}{
		{"dub.sh", "abc123", "dub.sh/abc123"},
		{"spn.sh", "demo", "spn.sh/demo"},
		{"example.com", "test-key", "example.com/test-key"},
	}

	for _, tt := range tests {
		result := buildShortLink(tt.domain, tt.key)
		if result != tt.expected {
			t.Errorf("buildShortLink(%q, %q): expected %q, got %q", tt.domain, tt.key, tt.expected, result)
		}
	}
}

func strPtr(s string) *string {
	return &s
}

func TestHandleLinksListResponse_TableOutput(t *testing.T) {
	jsonBody := `[
		{"id": "1", "domain": "dub.sh", "key": "abc123", "url": "https://example.com/very-long-path-that-should-be-truncated", "clicks": 1234, "lastClicked": "2024-01-15T10:30:00Z"},
		{"id": "2", "domain": "dub.sh", "key": "xyz789", "url": "https://other.site/page", "clicks": 456, "lastClicked": "2023-12-10T14:45:00Z"},
		{"id": "3", "domain": "spn.sh", "key": "demo", "url": "https://demo.example.com", "clicks": 0, "lastClicked": null}
	]`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(jsonBody)),
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := handleLinksListResponse(cmd, resp, "table", 25, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check header is present
	if !strings.Contains(output, "SHORT LINK") {
		t.Error("expected output to contain 'SHORT LINK' header")
	}
	if !strings.Contains(output, "URL") {
		t.Error("expected output to contain 'URL' header")
	}
	if !strings.Contains(output, "CLICKS") {
		t.Error("expected output to contain 'CLICKS' header")
	}
	if !strings.Contains(output, "LAST CLICKED") {
		t.Error("expected output to contain 'LAST CLICKED' header")
	}

	// Check data is present
	if !strings.Contains(output, "dub.sh/abc123") {
		t.Error("expected output to contain 'dub.sh/abc123'")
	}
	if !strings.Contains(output, "1,234") {
		t.Error("expected output to contain '1,234' (formatted clicks)")
	}
	if !strings.Contains(output, "Jan 15, 2024") {
		t.Error("expected output to contain 'Jan 15, 2024'")
	}
	if !strings.Contains(output, "spn.sh/demo") {
		t.Error("expected output to contain 'spn.sh/demo'")
	}
}

func TestHandleLinksListResponse_JSONOutput(t *testing.T) {
	jsonBody := `[{"id": "1", "domain": "dub.sh", "key": "abc123", "url": "https://example.com", "clicks": 100}]`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(jsonBody)),
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := handleLinksListResponse(cmd, resp, "json", 25, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// JSON output should contain the raw JSON structure
	if !strings.Contains(output, `"id"`) {
		t.Error("expected JSON output to contain 'id' field")
	}
	if !strings.Contains(output, `"dub.sh"`) {
		t.Error("expected JSON output to contain domain value")
	}
}

func TestHandleLinksListResponse_Limit(t *testing.T) {
	// Create 5 links
	jsonBody := `[
		{"id": "1", "domain": "dub.sh", "key": "link1", "url": "https://example.com/1", "clicks": 1},
		{"id": "2", "domain": "dub.sh", "key": "link2", "url": "https://example.com/2", "clicks": 2},
		{"id": "3", "domain": "dub.sh", "key": "link3", "url": "https://example.com/3", "clicks": 3},
		{"id": "4", "domain": "dub.sh", "key": "link4", "url": "https://example.com/4", "clicks": 4},
		{"id": "5", "domain": "dub.sh", "key": "link5", "url": "https://example.com/5", "clicks": 5}
	]`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(jsonBody)),
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Limit to 2
	err := handleLinksListResponse(cmd, resp, "table", 2, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain first 2 links
	if !strings.Contains(output, "dub.sh/link1") {
		t.Error("expected output to contain 'dub.sh/link1'")
	}
	if !strings.Contains(output, "dub.sh/link2") {
		t.Error("expected output to contain 'dub.sh/link2'")
	}

	// Should NOT contain link3, link4, link5
	if strings.Contains(output, "dub.sh/link3") {
		t.Error("expected output NOT to contain 'dub.sh/link3'")
	}

	// Should show pagination message
	if !strings.Contains(output, "Showing 2 of 5 links") {
		t.Error("expected output to contain pagination message")
	}
}

func TestHandleLinksListResponse_AllFlag(t *testing.T) {
	// Create 3 links
	jsonBody := `[
		{"id": "1", "domain": "dub.sh", "key": "link1", "url": "https://example.com/1", "clicks": 1},
		{"id": "2", "domain": "dub.sh", "key": "link2", "url": "https://example.com/2", "clicks": 2},
		{"id": "3", "domain": "dub.sh", "key": "link3", "url": "https://example.com/3", "clicks": 3}
	]`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(jsonBody)),
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// With --all flag, should show all links even with limit=1
	err := handleLinksListResponse(cmd, resp, "table", 1, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain all links
	if !strings.Contains(output, "dub.sh/link1") {
		t.Error("expected output to contain 'dub.sh/link1'")
	}
	if !strings.Contains(output, "dub.sh/link2") {
		t.Error("expected output to contain 'dub.sh/link2'")
	}
	if !strings.Contains(output, "dub.sh/link3") {
		t.Error("expected output to contain 'dub.sh/link3'")
	}

	// Should NOT show pagination message
	if strings.Contains(output, "Showing") {
		t.Error("expected output NOT to contain pagination message when --all is used")
	}
}
