// internal/outfmt/list_test.go
package outfmt

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestHandleListResponse_TableOutput(t *testing.T) {
	columns := []Column{
		{Name: "Name", Width: 20, Align: AlignLeft},
		{Name: "Slug", Width: 15, Align: AlignLeft},
	}

	mapper := func(item map[string]interface{}) []string {
		return []string{
			SafeString(item["name"]),
			SafeString(item["slug"]),
		}
	}

	data := []interface{}{
		map[string]interface{}{"name": "Example", "slug": "example"},
		map[string]interface{}{"name": "Test Site", "slug": "test-site"},
	}

	var buf bytes.Buffer
	cfg := ListConfig{
		Columns:   columns,
		RowMapper: mapper,
		Limit:     10,
		All:       false,
		Output:    "table",
	}

	err := HandleListResponse(&buf, data, 2, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 3 lines: header + 2 data rows
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %s", len(lines), output)
	}

	// Check headers
	if !strings.Contains(lines[0], "NAME") || !strings.Contains(lines[0], "SLUG") {
		t.Errorf("expected headers, got: %s", lines[0])
	}

	// Check data
	if !strings.Contains(lines[1], "Example") {
		t.Errorf("expected Example in first row, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Test Site") {
		t.Errorf("expected Test Site in second row, got: %s", lines[2])
	}
}

func TestHandleListResponse_TableWithLimit(t *testing.T) {
	columns := []Column{
		{Name: "Name", Width: 20, Align: AlignLeft},
	}

	mapper := func(item map[string]interface{}) []string {
		return []string{SafeString(item["name"])}
	}

	data := []interface{}{
		map[string]interface{}{"name": "Item 1"},
		map[string]interface{}{"name": "Item 2"},
		map[string]interface{}{"name": "Item 3"},
		map[string]interface{}{"name": "Item 4"},
		map[string]interface{}{"name": "Item 5"},
	}

	var buf bytes.Buffer
	cfg := ListConfig{
		Columns:   columns,
		RowMapper: mapper,
		Limit:     3,
		All:       false,
		Output:    "table",
	}

	err := HandleListResponse(&buf, data, 10, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 4 lines: header + 3 data rows (limited)
	// The pagination message is on a separate line
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines, got %d: %s", len(lines), output)
	}

	// Check pagination message
	if !strings.Contains(output, "Showing 3 of 10 items") {
		t.Errorf("expected pagination message, got: %s", output)
	}

	if !strings.Contains(output, "--limit") || !strings.Contains(output, "--all") {
		t.Errorf("expected pagination hint about --limit and --all, got: %s", output)
	}

	// Should not contain Item 4 or Item 5
	if strings.Contains(output, "Item 4") || strings.Contains(output, "Item 5") {
		t.Errorf("expected items to be limited, got: %s", output)
	}
}

func TestHandleListResponse_AllFlag(t *testing.T) {
	columns := []Column{
		{Name: "Name", Width: 20, Align: AlignLeft},
	}

	mapper := func(item map[string]interface{}) []string {
		return []string{SafeString(item["name"])}
	}

	data := []interface{}{
		map[string]interface{}{"name": "Item 1"},
		map[string]interface{}{"name": "Item 2"},
		map[string]interface{}{"name": "Item 3"},
		map[string]interface{}{"name": "Item 4"},
		map[string]interface{}{"name": "Item 5"},
	}

	var buf bytes.Buffer
	cfg := ListConfig{
		Columns:   columns,
		RowMapper: mapper,
		Limit:     3,
		All:       true, // --all flag set
		Output:    "table",
	}

	err := HandleListResponse(&buf, data, 5, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain all items
	if !strings.Contains(output, "Item 4") || !strings.Contains(output, "Item 5") {
		t.Errorf("expected all items with --all flag, got: %s", output)
	}

	// Should not show pagination message
	if strings.Contains(output, "Showing") {
		t.Errorf("expected no pagination message with --all, got: %s", output)
	}
}

func TestHandleListResponse_JSONOutput(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Example", "slug": "example"},
		map[string]interface{}{"name": "Test Site", "slug": "test-site"},
	}

	var buf bytes.Buffer
	cfg := ListConfig{
		Output: "json",
	}

	err := HandleListResponse(&buf, data, 2, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON, got: %s", buf.String())
	}

	if len(result) != 2 {
		t.Errorf("expected 2 items in JSON, got %d", len(result))
	}

	if result[0]["name"] != "Example" {
		t.Errorf("expected first item name to be Example, got: %v", result[0]["name"])
	}
}

func TestHandleListResponse_JSONWithQuery(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Example", "slug": "example"},
		map[string]interface{}{"name": "Test Site", "slug": "test-site"},
	}

	var buf bytes.Buffer
	cfg := ListConfig{
		Output: "json",
		Query:  ".[0].name",
	}

	err := HandleListResponse(&buf, data, 2, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != `"Example"` {
		t.Errorf("expected \"Example\", got: %s", output)
	}
}

func TestHandleListResponse_NoLimitSet(t *testing.T) {
	columns := []Column{
		{Name: "Name", Width: 20, Align: AlignLeft},
	}

	mapper := func(item map[string]interface{}) []string {
		return []string{SafeString(item["name"])}
	}

	data := []interface{}{
		map[string]interface{}{"name": "Item 1"},
		map[string]interface{}{"name": "Item 2"},
	}

	var buf bytes.Buffer
	cfg := ListConfig{
		Columns:   columns,
		RowMapper: mapper,
		Limit:     0, // No limit
		All:       false,
		Output:    "table",
	}

	err := HandleListResponse(&buf, data, 2, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should not show pagination message when limit is 0
	if strings.Contains(output, "Showing") {
		t.Errorf("expected no pagination message when limit is 0, got: %s", output)
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "RFC3339 string",
			input: "2024-01-15T10:30:00Z",
			want:  "Jan 15, 2024",
		},
		{
			name:  "RFC3339Nano string",
			input: "2024-12-25T14:30:00.123456789Z",
			want:  "Dec 25, 2024",
		},
		{
			name:  "nil value",
			input: nil,
			want:  "-",
		},
		{
			name:  "empty string",
			input: "",
			want:  "-",
		},
		{
			name:  "nil pointer",
			input: (*string)(nil),
			want:  "-",
		},
		{
			name:  "pointer to string",
			input: ptrString("2024-06-20T08:00:00Z"),
			want:  "Jun 20, 2024",
		},
		{
			name:  "unparseable string returns original",
			input: "not-a-date",
			want:  "not-a-date",
		},
		{
			name:  "non-string type",
			input: 12345,
			want:  "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDate(tt.input)
			if got != tt.want {
				t.Errorf("FormatDate(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatBool(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "true",
			input: true,
			want:  "Yes",
		},
		{
			name:  "false",
			input: false,
			want:  "No",
		},
		{
			name:  "nil",
			input: nil,
			want:  "-",
		},
		{
			name:  "pointer to true",
			input: ptrBool(true),
			want:  "Yes",
		},
		{
			name:  "pointer to false",
			input: ptrBool(false),
			want:  "No",
		},
		{
			name:  "nil bool pointer",
			input: (*bool)(nil),
			want:  "-",
		},
		{
			name:  "non-bool type",
			input: "true",
			want:  "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBool(tt.input)
			if got != tt.want {
				t.Errorf("FormatBool(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSafeString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "string",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "nil",
			input: nil,
			want:  "",
		},
		{
			name:  "pointer to string",
			input: ptrString("world"),
			want:  "world",
		},
		{
			name:  "nil string pointer",
			input: (*string)(nil),
			want:  "",
		},
		{
			name:  "integer",
			input: 42,
			want:  "42",
		},
		{
			name:  "float",
			input: 3.14,
			want:  "3.14",
		},
		{
			name:  "bool",
			input: true,
			want:  "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeString(tt.input)
			if got != tt.want {
				t.Errorf("SafeString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSafeInt(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int
	}{
		{
			name:  "int",
			input: 42,
			want:  42,
		},
		{
			name:  "int64",
			input: int64(100),
			want:  100,
		},
		{
			name:  "float64",
			input: float64(99.9),
			want:  99,
		},
		{
			name:  "nil",
			input: nil,
			want:  0,
		},
		{
			name:  "pointer to int",
			input: ptrInt(55),
			want:  55,
		},
		{
			name:  "nil int pointer",
			input: (*int)(nil),
			want:  0,
		},
		{
			name:  "string returns 0",
			input: "42",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeInt(tt.input)
			if got != tt.want {
				t.Errorf("SafeInt(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestSafeFloat(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{
			name:  "float64",
			input: 3.14,
			want:  3.14,
		},
		{
			name:  "float32",
			input: float32(2.5),
			want:  2.5,
		},
		{
			name:  "int",
			input: 42,
			want:  42.0,
		},
		{
			name:  "int64",
			input: int64(100),
			want:  100.0,
		},
		{
			name:  "nil",
			input: nil,
			want:  0,
		},
		{
			name:  "pointer to float64",
			input: ptrFloat64(9.99),
			want:  9.99,
		},
		{
			name:  "nil float64 pointer",
			input: (*float64)(nil),
			want:  0,
		},
		{
			name:  "string returns 0",
			input: "3.14",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeFloat(tt.input)
			if got != tt.want {
				t.Errorf("SafeFloat(%v) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

func TestHandleListResponse_EmptyData(t *testing.T) {
	columns := []Column{
		{Name: "Name", Width: 20, Align: AlignLeft},
	}

	mapper := func(item map[string]interface{}) []string {
		return []string{SafeString(item["name"])}
	}

	data := []interface{}{}

	var buf bytes.Buffer
	cfg := ListConfig{
		Columns:   columns,
		RowMapper: mapper,
		Limit:     10,
		Output:    "table",
	}

	err := HandleListResponse(&buf, data, 0, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should only have header row
	if len(lines) != 1 {
		t.Errorf("expected 1 line (header only), got %d", len(lines))
	}

	if !strings.Contains(lines[0], "NAME") {
		t.Errorf("expected header, got: %s", lines[0])
	}
}

// Helper functions for creating pointers
func ptrString(s string) *string {
	return &s
}

func ptrBool(b bool) *bool {
	return &b
}

func ptrInt(i int) *int {
	return &i
}

func ptrFloat64(f float64) *float64 {
	return &f
}
