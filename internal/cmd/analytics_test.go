// internal/cmd/analytics_test.go
package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
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
		"output",
		"limit",
		"all",
	}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to exist", name)
		}
	}
}

func TestAnalyticsCmd_OutputFlagShorthand(t *testing.T) {
	cmd := newAnalyticsCmd()

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected flag 'output' to exist")
	}
	if flag.Shorthand != "o" {
		t.Errorf("expected output flag shorthand to be 'o', got %q", flag.Shorthand)
	}
}

func TestAnalyticsCmd_DefaultOutput(t *testing.T) {
	cmd := newAnalyticsCmd()

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("expected flag 'output' to exist")
	}
	if flag.DefValue != "table" {
		t.Errorf("expected output default to be 'table', got %q", flag.DefValue)
	}
}

func TestAnalyticsCmd_DefaultLimit(t *testing.T) {
	cmd := newAnalyticsCmd()

	flag := cmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("expected flag 'limit' to exist")
	}
	if flag.DefValue != "25" {
		t.Errorf("expected limit default to be '25', got %q", flag.DefValue)
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

func TestFormatMetricValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"zero", float64(0), "0"},
		{"small number", float64(42), "42"},
		{"hundreds", float64(999), "999"},
		{"thousands", float64(1234), "1,234"},
		{"tens of thousands", float64(12345), "12,345"},
		{"hundreds of thousands", float64(123456), "123,456"},
		{"millions", float64(1234567), "1,234,567"},
		{"nil value", nil, "0"},
		{"integer", 5432, "5,432"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMetricValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatMetricValue(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetGroupByColumn(t *testing.T) {
	tests := []struct {
		groupBy      string
		expectedName string
		expectedKey  string
	}{
		{"countries", "Country", "country"},
		{"cities", "City", "city"},
		{"devices", "Device", "device"},
		{"browsers", "Browser", "browser"},
		{"os", "OS", "os"},
		{"referers", "Referer", "referer"},
		{"unknown", "Value", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.groupBy, func(t *testing.T) {
			name, key := getGroupByColumn(tt.groupBy)
			if name != tt.expectedName {
				t.Errorf("getGroupByColumn(%q) name = %q, want %q", tt.groupBy, name, tt.expectedName)
			}
			if key != tt.expectedKey {
				t.Errorf("getGroupByColumn(%q) key = %q, want %q", tt.groupBy, key, tt.expectedKey)
			}
		})
	}
}

func TestGetGroupByNoun(t *testing.T) {
	tests := []struct {
		groupBy  string
		expected string
	}{
		{"countries", "countries"},
		{"cities", "cities"},
		{"devices", "devices"},
		{"browsers", "browsers"},
		{"os", "operating systems"},
		{"referers", "referers"},
		{"unknown", "items"},
	}

	for _, tt := range tests {
		t.Run(tt.groupBy, func(t *testing.T) {
			result := getGroupByNoun(tt.groupBy)
			if result != tt.expected {
				t.Errorf("getGroupByNoun(%q) = %q, want %q", tt.groupBy, result, tt.expected)
			}
		})
	}
}

// mockReadCloser implements io.ReadCloser for testing
type mockReadCloser struct {
	io.Reader
}

func (m mockReadCloser) Close() error {
	return nil
}

func TestHandleAnalyticsResponse_CountFormat(t *testing.T) {
	cmd := newAnalyticsCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	body := `{"clicks": 1234, "leads": 45, "sales": 12}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       mockReadCloser{strings.NewReader(body)},
	}

	err := handleAnalyticsResponse(cmd, resp, "", "table", 25, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "METRIC") {
		t.Error("expected output to contain 'METRIC' header")
	}
	if !strings.Contains(output, "Clicks") {
		t.Error("expected output to contain 'Clicks'")
	}
	if !strings.Contains(output, "1,234") {
		t.Error("expected output to contain formatted click count '1,234'")
	}
}

func TestHandleAnalyticsResponse_TimeseriesFormat(t *testing.T) {
	cmd := newAnalyticsCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	body := `[
		{"start": "2024-01-15T00:00:00Z", "clicks": 1234, "leads": 45, "sales": 12},
		{"start": "2024-01-14T00:00:00Z", "clicks": 987, "leads": 32, "sales": 8}
	]`
	resp := &http.Response{
		StatusCode: 200,
		Body:       mockReadCloser{strings.NewReader(body)},
	}

	err := handleAnalyticsResponse(cmd, resp, "timeseries", "table", 25, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "DATE") {
		t.Error("expected output to contain 'DATE' header")
	}
	if !strings.Contains(output, "CLICKS") {
		t.Error("expected output to contain 'CLICKS' header")
	}
	if !strings.Contains(output, "Jan 15, 2024") {
		t.Error("expected output to contain formatted date 'Jan 15, 2024'")
	}
}

func TestHandleAnalyticsResponse_CountriesFormat(t *testing.T) {
	cmd := newAnalyticsCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	body := `[
		{"country": "US", "clicks": 5432, "leads": 123, "sales": 45},
		{"country": "DE", "clicks": 2341, "leads": 67, "sales": 23},
		{"country": "UK", "clicks": 1234, "leads": 34, "sales": 12}
	]`
	resp := &http.Response{
		StatusCode: 200,
		Body:       mockReadCloser{strings.NewReader(body)},
	}

	err := handleAnalyticsResponse(cmd, resp, "countries", "table", 25, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "COUNTRY") {
		t.Error("expected output to contain 'COUNTRY' header")
	}
	if !strings.Contains(output, "US") {
		t.Error("expected output to contain 'US'")
	}
	if !strings.Contains(output, "5,432") {
		t.Error("expected output to contain formatted count '5,432'")
	}
}

func TestHandleAnalyticsResponse_LimitApplied(t *testing.T) {
	cmd := newAnalyticsCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Create 5 countries but limit to 2
	body := `[
		{"country": "US", "clicks": 100, "leads": 10, "sales": 5},
		{"country": "DE", "clicks": 90, "leads": 9, "sales": 4},
		{"country": "UK", "clicks": 80, "leads": 8, "sales": 3},
		{"country": "FR", "clicks": 70, "leads": 7, "sales": 2},
		{"country": "ES", "clicks": 60, "leads": 6, "sales": 1}
	]`
	resp := &http.Response{
		StatusCode: 200,
		Body:       mockReadCloser{strings.NewReader(body)},
	}

	err := handleAnalyticsResponse(cmd, resp, "countries", "table", 2, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "US") {
		t.Error("expected output to contain 'US'")
	}
	if !strings.Contains(output, "DE") {
		t.Error("expected output to contain 'DE'")
	}
	if strings.Contains(output, "UK") {
		t.Error("expected output to NOT contain 'UK' (should be limited)")
	}
	if !strings.Contains(output, "Showing 2 of 5 countries") {
		t.Error("expected output to contain pagination message")
	}
}

func TestHandleAnalyticsResponse_AllFlag(t *testing.T) {
	cmd := newAnalyticsCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Create 3 countries with limit 2 but --all flag
	body := `[
		{"country": "US", "clicks": 100, "leads": 10, "sales": 5},
		{"country": "DE", "clicks": 90, "leads": 9, "sales": 4},
		{"country": "UK", "clicks": 80, "leads": 8, "sales": 3}
	]`
	resp := &http.Response{
		StatusCode: 200,
		Body:       mockReadCloser{strings.NewReader(body)},
	}

	err := handleAnalyticsResponse(cmd, resp, "countries", "table", 2, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "UK") {
		t.Error("expected output to contain 'UK' (--all flag should show all)")
	}
	if strings.Contains(output, "Showing") {
		t.Error("expected output to NOT contain pagination message when --all is used")
	}
}

func TestHandleAnalyticsResponse_JSONOutput(t *testing.T) {
	cmd := newAnalyticsCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetContext(context.Background())

	body := `{"clicks": 1234, "leads": 45}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       mockReadCloser{strings.NewReader(body)},
	}

	err := handleAnalyticsResponse(cmd, resp, "", "json", 25, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// JSON output should preserve the raw format
	if !strings.Contains(output, `"clicks"`) {
		t.Error("expected JSON output to contain 'clicks' key")
	}
	if !strings.Contains(output, "1234") {
		t.Error("expected JSON output to contain raw number 1234")
	}
}

func TestHandleAnalyticsResponse_ErrorStatus(t *testing.T) {
	cmd := newAnalyticsCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	body := `{"error": {"code": "not_found", "message": "Link not found"}}`
	resp := &http.Response{
		StatusCode: 404,
		Body:       mockReadCloser{strings.NewReader(body)},
	}

	err := handleAnalyticsResponse(cmd, resp, "", "table", 25, false)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}
