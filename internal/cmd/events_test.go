package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

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
	flags := []string{"event", "domain", "link-id", "interval", "start", "end", "country", "city", "device", "browser", "os", "referer", "output", "limit", "all"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestEventsListCmd_FlagDefaults(t *testing.T) {
	cmd := newEventsListCmd()

	// Check output flag default and shorthand
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("expected 'output' flag to exist")
	}
	if outputFlag.DefValue != "table" {
		t.Errorf("expected output default 'table', got %q", outputFlag.DefValue)
	}
	if outputFlag.Shorthand != "o" {
		t.Errorf("expected output shorthand 'o', got %q", outputFlag.Shorthand)
	}

	// Check limit flag default
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Fatal("expected 'limit' flag to exist")
	}
	if limitFlag.DefValue != "25" {
		t.Errorf("expected limit default '25', got %q", limitFlag.DefValue)
	}

	// Check all flag default
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("expected 'all' flag to exist")
	}
	if allFlag.DefValue != "false" {
		t.Errorf("expected all default 'false', got %q", allFlag.DefValue)
	}
}

func TestEventsListCmd_PageFlagRemoved(t *testing.T) {
	cmd := newEventsListCmd()
	if cmd.Flags().Lookup("page") != nil {
		t.Error("expected 'page' flag to be removed")
	}
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "RFC3339 timestamp",
			input:    "2024-01-15T15:42:00Z",
			expected: "Jan 15, 3:42 PM",
		},
		{
			name:     "RFC3339Nano timestamp",
			input:    "2024-01-15T14:30:00.123456789Z",
			expected: "Jan 15, 2:30 PM",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "-",
		},
		{
			name:     "nil value",
			input:    nil,
			expected: "-",
		},
		{
			name:     "invalid timestamp",
			input:    "not-a-timestamp",
			expected: "not-a-timestamp",
		},
		{
			name:     "AM timestamp",
			input:    "2024-01-14T11:15:00Z",
			expected: "Jan 14, 11:15 AM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimestamp(tt.input)
			if result != tt.expected {
				t.Errorf("formatTimestamp(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatEventLink(t *testing.T) {
	tests := []struct {
		name     string
		event    map[string]interface{}
		expected string
	}{
		{
			name: "shortLink in link object",
			event: map[string]interface{}{
				"link": map[string]interface{}{
					"shortLink": "dub.sh/abc123",
				},
			},
			expected: "dub.sh/abc123",
		},
		{
			name: "domain and key in link object",
			event: map[string]interface{}{
				"link": map[string]interface{}{
					"domain": "dub.sh",
					"key":    "xyz789",
				},
			},
			expected: "dub.sh/xyz789",
		},
		{
			name: "id in link object",
			event: map[string]interface{}{
				"link": map[string]interface{}{
					"id": "link_abc123def456",
				},
			},
			expected: "link_abc123def456",
		},
		{
			name: "linkId at top level",
			event: map[string]interface{}{
				"linkId": "link_toplevel123",
			},
			expected: "link_toplevel123",
		},
		{
			name:     "no link information",
			event:    map[string]interface{}{},
			expected: "-",
		},
		{
			name: "truncates long shortLink",
			event: map[string]interface{}{
				"link": map[string]interface{}{
					"shortLink": "dub.sh/verylongkeyname12345",
				},
			},
			expected: "dub.sh/verylongke...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEventLink(tt.event)
			if result != tt.expected {
				t.Errorf("formatEventLink() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatEventField(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string value",
			input:    "US",
			expected: "US",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "-",
		},
		{
			name:     "nil value",
			input:    nil,
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEventField(tt.input)
			if result != tt.expected {
				t.Errorf("formatEventField(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHandleEventsListResponse_TableOutput(t *testing.T) {
	jsonBody := `[
		{
			"timestamp": "2024-01-15T15:42:00Z",
			"event": "click",
			"link": {"shortLink": "dub.sh/abc123"},
			"country": "US",
			"device": "desktop",
			"browser": "Chrome"
		},
		{
			"timestamp": "2024-01-15T14:30:00Z",
			"event": "lead",
			"link": {"shortLink": "dub.sh/xyz789"},
			"country": "DE",
			"device": "mobile",
			"browser": "Safari"
		}
	]`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(jsonBody)),
	}

	cmd := newEventsListCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := handleEventsListResponse(cmd, resp, "table", 25, false)
	if err != nil {
		t.Fatalf("handleEventsListResponse() error = %v", err)
	}

	output := buf.String()

	// Check header
	if !bytes.Contains([]byte(output), []byte("TIMESTAMP")) {
		t.Error("expected output to contain 'TIMESTAMP' header")
	}
	if !bytes.Contains([]byte(output), []byte("EVENT")) {
		t.Error("expected output to contain 'EVENT' header")
	}
	if !bytes.Contains([]byte(output), []byte("LINK")) {
		t.Error("expected output to contain 'LINK' header")
	}
	if !bytes.Contains([]byte(output), []byte("COUNTRY")) {
		t.Error("expected output to contain 'COUNTRY' header")
	}
	if !bytes.Contains([]byte(output), []byte("DEVICE")) {
		t.Error("expected output to contain 'DEVICE' header")
	}
	if !bytes.Contains([]byte(output), []byte("BROWSER")) {
		t.Error("expected output to contain 'BROWSER' header")
	}

	// Check data
	if !bytes.Contains([]byte(output), []byte("click")) {
		t.Error("expected output to contain 'click' event")
	}
	if !bytes.Contains([]byte(output), []byte("dub.sh/abc123")) {
		t.Error("expected output to contain 'dub.sh/abc123' link")
	}
}

func TestHandleEventsListResponse_JSONOutput(t *testing.T) {
	jsonBody := `[{"event": "click"}]`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(jsonBody)),
	}

	cmd := newEventsListCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetContext(context.Background())

	err := handleEventsListResponse(cmd, resp, "json", 25, false)
	if err != nil {
		t.Fatalf("handleEventsListResponse() error = %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte(`"event"`)) {
		t.Error("expected JSON output to contain event field")
	}
}

func TestHandleEventsListResponse_Pagination(t *testing.T) {
	// Create 30 events
	jsonBody := `[`
	for i := 0; i < 30; i++ {
		if i > 0 {
			jsonBody += ","
		}
		jsonBody += `{"timestamp": "2024-01-15T15:42:00Z", "event": "click", "country": "US", "device": "desktop", "browser": "Chrome"}`
	}
	jsonBody += `]`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(jsonBody)),
	}

	cmd := newEventsListCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := handleEventsListResponse(cmd, resp, "table", 25, false)
	if err != nil {
		t.Fatalf("handleEventsListResponse() error = %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Showing 25 of 30 events")) {
		t.Error("expected pagination message 'Showing 25 of 30 events'")
	}
}

func TestHandleEventsListResponse_AllFlag(t *testing.T) {
	// Create 30 events
	jsonBody := `[`
	for i := 0; i < 30; i++ {
		if i > 0 {
			jsonBody += ","
		}
		jsonBody += `{"timestamp": "2024-01-15T15:42:00Z", "event": "click", "country": "US", "device": "desktop", "browser": "Chrome"}`
	}
	jsonBody += `]`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(jsonBody)),
	}

	cmd := newEventsListCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := handleEventsListResponse(cmd, resp, "table", 25, true) // all=true
	if err != nil {
		t.Fatalf("handleEventsListResponse() error = %v", err)
	}

	output := buf.String()
	// Should NOT show pagination message when --all is used
	if bytes.Contains([]byte(output), []byte("Showing")) {
		t.Error("expected no pagination message when --all is used")
	}
}

func TestHandleEventsListResponse_APIError(t *testing.T) {
	jsonBody := `{"error": {"code": "unauthorized", "message": "Invalid API key"}}`

	resp := &http.Response{
		StatusCode: 401,
		Body:       io.NopCloser(bytes.NewBufferString(jsonBody)),
	}

	cmd := newEventsListCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := handleEventsListResponse(cmd, resp, "table", 25, false)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("unauthorized")) {
		t.Errorf("expected error to contain 'unauthorized', got: %v", err)
	}
}
