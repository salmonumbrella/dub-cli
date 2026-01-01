// internal/outfmt/table.go
package outfmt

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Align specifies text alignment within a column.
type Align int

const (
	AlignLeft Align = iota
	AlignRight
)

// Column defines a table column configuration.
type Column struct {
	Name  string // Header text (will be uppercased)
	Width int    // Maximum width in characters (0 means no limit)
	Align Align  // Text alignment (left or right)
}

// columnGap is the minimum spacing between columns.
const columnGap = 2

// Truncate shortens a string to maxLen characters, appending "..." if truncated.
// If maxLen is less than 4, the string is truncated without ellipsis.
// If maxLen is 0 or negative, the original string is returned unchanged.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return s
	}

	runeCount := utf8.RuneCountInString(s)
	if runeCount <= maxLen {
		return s
	}

	// If maxLen is too short for ellipsis, just truncate
	if maxLen < 4 {
		runes := []rune(s)
		return string(runes[:maxLen])
	}

	// Truncate with ellipsis
	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}

// FormatTable renders structured data as an aligned ASCII table.
// It writes column headers (uppercased) followed by data rows.
// Columns are separated by at least columnGap spaces.
func FormatTable(w io.Writer, columns []Column, rows [][]string) error {
	if len(columns) == 0 {
		return nil
	}

	// Calculate actual column widths based on content
	widths := make([]int, len(columns))
	for i, col := range columns {
		// Start with header width
		widths[i] = utf8.RuneCountInString(col.Name)

		// Check if column has a fixed max width
		if col.Width > 0 && widths[i] > col.Width {
			widths[i] = col.Width
		}
	}

	// Expand widths based on row content (up to column max width)
	for _, row := range rows {
		for i := 0; i < len(columns) && i < len(row); i++ {
			cellWidth := utf8.RuneCountInString(row[i])

			// Apply column max width constraint
			if columns[i].Width > 0 && cellWidth > columns[i].Width {
				cellWidth = columns[i].Width
			}

			if cellWidth > widths[i] {
				widths[i] = cellWidth
			}
		}
	}

	// Write header row
	if err := writeRow(w, columns, widths, headerRow(columns)); err != nil {
		return err
	}

	// Write data rows
	for _, row := range rows {
		if err := writeRow(w, columns, widths, row); err != nil {
			return err
		}
	}

	return nil
}

// headerRow creates a row of uppercase column names.
func headerRow(columns []Column) []string {
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = strings.ToUpper(col.Name)
	}
	return headers
}

// writeRow writes a single row with proper alignment and spacing.
func writeRow(w io.Writer, columns []Column, widths []int, row []string) error {
	var sb strings.Builder

	for i, col := range columns {
		var cell string
		if i < len(row) {
			cell = row[i]
		}

		// Apply truncation if column has max width
		if col.Width > 0 {
			cell = Truncate(cell, col.Width)
		}

		// Pad and align
		cellWidth := utf8.RuneCountInString(cell)
		padding := widths[i] - cellWidth

		if col.Align == AlignRight {
			sb.WriteString(strings.Repeat(" ", padding))
			sb.WriteString(cell)
		} else {
			sb.WriteString(cell)
			sb.WriteString(strings.Repeat(" ", padding))
		}

		// Add column gap (except for last column)
		if i < len(columns)-1 {
			sb.WriteString(strings.Repeat(" ", columnGap))
		}
	}

	// Trim trailing whitespace and write
	line := strings.TrimRight(sb.String(), " ")
	_, err := fmt.Fprintln(w, line)
	return err
}
