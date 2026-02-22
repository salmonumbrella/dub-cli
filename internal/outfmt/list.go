// internal/outfmt/list.go
package outfmt

import (
	"fmt"
	"io"
	"time"
)

// RowMapper converts a single item from the API response into table row values.
type RowMapper func(item map[string]interface{}) []string

// ListConfig configures list output behavior.
type ListConfig struct {
	Columns   []Column
	RowMapper RowMapper
	Limit     int    // 0 means no limit
	All       bool   // if true, ignore limit
	Output    string // "table" or "json"
	Query     string // jq query for JSON output
}

// HandleListResponse processes a list API response and formats it as table or JSON.
// The data parameter should be a slice of items from the API response.
// The total parameter is the total count of items available (for pagination message).
func HandleListResponse(w io.Writer, data []interface{}, total int, cfg ListConfig) error {
	if cfg.Output == "json" {
		return FormatJSON(w, data, cfg.Query)
	}

	// Table output
	displayData := data
	limited := false

	// Apply limit unless --all flag is set
	if !cfg.All && cfg.Limit > 0 && len(data) > cfg.Limit {
		displayData = data[:cfg.Limit]
		limited = true
	}

	// Convert items to table rows
	rows := make([][]string, 0, len(displayData))
	for _, item := range displayData {
		if m, ok := item.(map[string]interface{}); ok {
			rows = append(rows, cfg.RowMapper(m))
		}
	}

	// Format and write table
	if err := FormatTable(w, cfg.Columns, rows); err != nil {
		return err
	}

	// Show pagination message if limited
	if limited {
		showing := len(displayData)
		available := len(data)
		if total > available {
			available = total
		}
		if _, err := fmt.Fprintf(w, "\nShowing %d of %d items. Use --limit or --all for more.\n", showing, available); err != nil {
			return err
		}
	}

	return nil
}

// FormatDate converts a timestamp interface to a human-readable date string.
// Handles *string, string, and nil. Returns "-" for nil or empty values.
// Attempts to parse RFC3339 format and returns "Jan 15, 2024" format.
func FormatDate(ts interface{}) string {
	var s string

	switch v := ts.(type) {
	case string:
		s = v
	case *string:
		if v == nil {
			return "-"
		}
		s = *v
	case nil:
		return "-"
	default:
		return "-"
	}

	if s == "" {
		return "-"
	}

	// Try parsing RFC3339 format
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try RFC3339Nano
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			// Return original string if parsing fails
			return s
		}
	}

	return t.Format("Jan 2, 2006")
}

// FormatBool converts a boolean interface to "Yes" or "No".
// Handles bool, *bool, and nil. Returns "-" for nil values.
func FormatBool(b interface{}) string {
	switch v := b.(type) {
	case bool:
		if v {
			return "Yes"
		}
		return "No"
	case *bool:
		if v == nil {
			return "-"
		}
		if *v {
			return "Yes"
		}
		return "No"
	case nil:
		return "-"
	default:
		return "-"
	}
}

// SafeString safely extracts a string from an interface{}.
// Returns empty string for nil or non-string types.
func SafeString(v interface{}) string {
	if v == nil {
		return ""
	}

	switch s := v.(type) {
	case string:
		return s
	case *string:
		if s == nil {
			return ""
		}
		return *s
	default:
		// Try to convert using fmt for other types (numbers, etc.)
		return fmt.Sprintf("%v", v)
	}
}

// SafeInt safely extracts an integer from an interface{}.
// Returns 0 for nil or values that cannot be converted to int.
func SafeInt(v interface{}) int {
	if v == nil {
		return 0
	}

	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case *int:
		if n == nil {
			return 0
		}
		return *n
	default:
		return 0
	}
}

// SafeFloat safely extracts a float64 from an interface{}.
// Returns 0 for nil or values that cannot be converted to float64.
func SafeFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}

	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case *float64:
		if n == nil {
			return 0
		}
		return *n
	default:
		return 0
	}
}
