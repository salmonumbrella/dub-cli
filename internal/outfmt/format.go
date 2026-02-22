// internal/outfmt/format.go
package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/itchyny/gojq"
)

type contextKey string

const (
	formatKey contextKey = "format"
	queryKey  contextKey = "query"
	yesKey    contextKey = "yes"
	limitKey  contextKey = "limit"
	sortByKey contextKey = "sortBy"
	descKey   contextKey = "desc"
)

func WithFormat(ctx context.Context, format string) context.Context {
	return context.WithValue(ctx, formatKey, format)
}

func GetFormat(ctx context.Context) string {
	if v, ok := ctx.Value(formatKey).(string); ok {
		return v
	}
	return "text"
}

func WithQuery(ctx context.Context, query string) context.Context {
	return context.WithValue(ctx, queryKey, query)
}

func GetQuery(ctx context.Context) string {
	if v, ok := ctx.Value(queryKey).(string); ok {
		return v
	}
	return ""
}

func WithYes(ctx context.Context, yes bool) context.Context {
	return context.WithValue(ctx, yesKey, yes)
}

func GetYes(ctx context.Context) bool {
	if v, ok := ctx.Value(yesKey).(bool); ok {
		return v
	}
	return false
}

func WithLimit(ctx context.Context, limit int) context.Context {
	return context.WithValue(ctx, limitKey, limit)
}

func GetLimit(ctx context.Context) int {
	if v, ok := ctx.Value(limitKey).(int); ok {
		return v
	}
	return 0
}

func WithSortBy(ctx context.Context, sortBy string) context.Context {
	return context.WithValue(ctx, sortByKey, sortBy)
}

func GetSortBy(ctx context.Context) string {
	if v, ok := ctx.Value(sortByKey).(string); ok {
		return v
	}
	return ""
}

func WithDesc(ctx context.Context, desc bool) context.Context {
	return context.WithValue(ctx, descKey, desc)
}

func GetDesc(ctx context.Context) bool {
	if v, ok := ctx.Value(descKey).(bool); ok {
		return v
	}
	return false
}

func FormatJSON(w io.Writer, data interface{}, query string) error {
	if query == "" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	// Apply jq query
	q, err := gojq.Parse(query)
	if err != nil {
		return fmt.Errorf("invalid query: %w", err)
	}

	// Normalize data through JSON round-trip for gojq compatibility
	// gojq requires map[string]interface{} not map[string]string
	normalized, err := normalizeForJQ(data)
	if err != nil {
		return fmt.Errorf("failed to normalize data: %w", err)
	}

	iter := q.Run(normalized)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}
		out, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(w, string(out))
	}
	return nil
}

// normalizeForJQ converts arbitrary Go types to JSON-compatible types
// that gojq can process (map[string]interface{}, []interface{}, etc.)
func normalizeForJQ(data interface{}) (interface{}, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var normalized interface{}
	if err := json.Unmarshal(jsonBytes, &normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}
