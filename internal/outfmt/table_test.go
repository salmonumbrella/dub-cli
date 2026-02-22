// internal/outfmt/table_test.go
package outfmt

import (
	"bytes"
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "no truncation needed",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "truncate with ellipsis",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "very short maxLen without ellipsis",
			input:  "hello",
			maxLen: 3,
			want:   "hel",
		},
		{
			name:   "zero maxLen returns original",
			input:  "hello",
			maxLen: 0,
			want:   "hello",
		},
		{
			name:   "negative maxLen returns original",
			input:  "hello",
			maxLen: -1,
			want:   "hello",
		},
		{
			name:   "unicode string truncation",
			input:  "hello world",
			maxLen: 7,
			want:   "hell...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "maxLen 4 with ellipsis",
			input:  "hello world",
			maxLen: 4,
			want:   "h...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFormatTable_Basic(t *testing.T) {
	columns := []Column{
		{Name: "Name", Width: 20, Align: AlignLeft},
		{Name: "Value", Width: 10, Align: AlignLeft},
	}

	rows := [][]string{
		{"Alice", "100"},
		{"Bob", "200"},
	}

	var buf bytes.Buffer
	err := FormatTable(&buf, columns, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Errorf("expected 3 lines (1 header + 2 rows), got %d", len(lines))
	}

	// Check header is uppercase
	if !strings.Contains(lines[0], "NAME") || !strings.Contains(lines[0], "VALUE") {
		t.Errorf("expected uppercase headers, got: %s", lines[0])
	}

	// Check data rows
	if !strings.Contains(lines[1], "Alice") || !strings.Contains(lines[1], "100") {
		t.Errorf("expected first row data, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Bob") || !strings.Contains(lines[2], "200") {
		t.Errorf("expected second row data, got: %s", lines[2])
	}
}

func TestFormatTable_RightAlignment(t *testing.T) {
	columns := []Column{
		{Name: "Name", Width: 10, Align: AlignLeft},
		{Name: "Count", Width: 10, Align: AlignRight},
	}

	rows := [][]string{
		{"Alice", "1234"},
		{"Bob", "56"},
	}

	var buf bytes.Buffer
	err := FormatTable(&buf, columns, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check that "COUNT" header exists and is right-aligned
	// For right-aligned columns, the value should have leading spaces
	header := lines[0]
	if !strings.Contains(header, "COUNT") {
		t.Errorf("expected COUNT header, got: %s", header)
	}

	// Check that numeric values are right-aligned (have leading spaces before them)
	// In right-aligned column, shorter values have more leading spaces
	row1 := lines[1]
	row2 := lines[2]

	// "1234" is 4 chars, "56" is 2 chars
	// With width 10, "56" should have more leading spaces
	idx1234 := strings.Index(row1, "1234")
	idx56 := strings.Index(row2, "56")

	// The 56 value should start at a higher position (more padding) than 1234
	// Both should end at approximately the same position
	if idx56 <= idx1234 {
		t.Errorf("expected right-aligned 56 to have more padding than 1234")
	}
}

func TestFormatTable_Truncation(t *testing.T) {
	columns := []Column{
		{Name: "URL", Width: 15, Align: AlignLeft},
	}

	rows := [][]string{
		{"https://example.com/very/long/path"},
	}

	var buf bytes.Buffer
	err := FormatTable(&buf, columns, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should be truncated with ellipsis
	if !strings.Contains(output, "...") {
		t.Errorf("expected truncated URL with ellipsis, got: %s", output)
	}

	// Should not contain the full URL
	if strings.Contains(output, "very/long/path") {
		t.Errorf("expected URL to be truncated, got: %s", output)
	}
}

func TestFormatTable_EmptyRows(t *testing.T) {
	columns := []Column{
		{Name: "Name", Width: 10, Align: AlignLeft},
		{Name: "Value", Width: 10, Align: AlignLeft},
	}

	rows := [][]string{}

	var buf bytes.Buffer
	err := FormatTable(&buf, columns, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should only have header row
	if len(lines) != 1 {
		t.Errorf("expected 1 line (header only), got %d", len(lines))
	}

	if !strings.Contains(lines[0], "NAME") || !strings.Contains(lines[0], "VALUE") {
		t.Errorf("expected headers even with no data, got: %s", lines[0])
	}
}

func TestFormatTable_EmptyColumns(t *testing.T) {
	columns := []Column{}
	rows := [][]string{{"a", "b"}}

	var buf bytes.Buffer
	err := FormatTable(&buf, columns, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output for no columns, got: %q", output)
	}
}

func TestFormatTable_ColumnGap(t *testing.T) {
	columns := []Column{
		{Name: "Name", Width: 10, Align: AlignLeft},
		{Name: "Value", Width: 10, Align: AlignLeft},
	}

	rows := [][]string{
		{"X", "Y"},
	}

	var buf bytes.Buffer
	err := FormatTable(&buf, columns, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check that there's spacing between columns (at least 2 spaces)
	// Header is "NAME" (4 chars), which determines column width since content "X" is shorter
	dataRow := lines[1]

	// Find position of Y (second column)
	yIdx := strings.Index(dataRow, "Y")

	// Y should be at least 2 positions after where X column ends
	// Column 1 width = "NAME" length = 4, plus gap of 2 = 6
	// So Y should be at position >= 6
	expectedMinPos := 4 + 2
	if yIdx < expectedMinPos {
		t.Errorf("expected column gap of at least 2 spaces, Y at position %d, expected >= %d", yIdx, expectedMinPos)
	}
}

func TestFormatTable_NoWidthLimit(t *testing.T) {
	columns := []Column{
		{Name: "Text", Width: 0, Align: AlignLeft}, // 0 means no limit
	}

	longText := "This is a very long text that should not be truncated at all"
	rows := [][]string{
		{longText},
	}

	var buf bytes.Buffer
	err := FormatTable(&buf, columns, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, longText) {
		t.Errorf("expected full text without truncation, got: %s", output)
	}
}

func TestFormatTable_MissingCells(t *testing.T) {
	columns := []Column{
		{Name: "A", Width: 10, Align: AlignLeft},
		{Name: "B", Width: 10, Align: AlignLeft},
		{Name: "C", Width: 10, Align: AlignLeft},
	}

	// Row with fewer cells than columns
	rows := [][]string{
		{"1", "2"}, // missing third cell
	}

	var buf bytes.Buffer
	err := FormatTable(&buf, columns, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should not panic, should handle missing cells gracefully
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") {
		t.Errorf("expected row data, got: %s", output)
	}
}

func TestFormatTable_LinksListExample(t *testing.T) {
	// Simulate the links list table format from the task description
	columns := []Column{
		{Name: "Short Link", Width: 22, Align: AlignLeft},
		{Name: "URL", Width: 40, Align: AlignLeft},
		{Name: "Clicks", Width: 7, Align: AlignRight},
		{Name: "Last Clicked", Width: 12, Align: AlignLeft},
	}

	rows := [][]string{
		{"dub.sh/abc123", "https://example.com/very-long-path-that-exceeds-limit", "1,234", "2024-01-15"},
		{"dub.sh/xyz789", "https://other.site/page", "456", "2024-01-10"},
	}

	var buf bytes.Buffer
	err := FormatTable(&buf, columns, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 3 lines: header + 2 data rows
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Check headers are uppercase
	header := lines[0]
	expectedHeaders := []string{"SHORT LINK", "URL", "CLICKS", "LAST CLICKED"}
	for _, h := range expectedHeaders {
		if !strings.Contains(header, h) {
			t.Errorf("expected header %q in: %s", h, header)
		}
	}

	// Check first URL is truncated (it exceeds 40 chars)
	if !strings.Contains(lines[1], "...") {
		t.Errorf("expected truncated URL in first row, got: %s", lines[1])
	}

	// Check second URL is not truncated (it's under 40 chars)
	if strings.Contains(lines[2], "...") {
		t.Errorf("expected non-truncated URL in second row, got: %s", lines[2])
	}

	// Check clicks are present
	if !strings.Contains(lines[1], "1,234") {
		t.Errorf("expected clicks value 1,234 in first row, got: %s", lines[1])
	}
}
