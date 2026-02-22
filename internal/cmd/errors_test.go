// internal/cmd/errors_test.go
package cmd

import (
	"errors"
	"testing"
)

func TestUsageError(t *testing.T) {
	t.Run("wraps error", func(t *testing.T) {
		inner := errors.New("test error")
		usageErr := NewUsageError(inner)

		if usageErr.Error() != "test error" {
			t.Errorf("expected error message 'test error', got %q", usageErr.Error())
		}

		if !errors.Is(usageErr, inner) {
			t.Error("expected UsageError to wrap inner error")
		}
	})

	t.Run("NewUsageErrorf formats correctly", func(t *testing.T) {
		usageErr := NewUsageErrorf("--%s is required", "url")

		if usageErr.Error() != "--url is required" {
			t.Errorf("expected error message '--url is required', got %q", usageErr.Error())
		}
	})
}

func TestIsUsageError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "wrapped UsageError",
			err:      NewUsageError(errors.New("test")),
			expected: true,
		},
		{
			name:     "unknown command",
			err:      errors.New("unknown command \"foo\" for \"bar\""),
			expected: true,
		},
		{
			name:     "unknown flag",
			err:      errors.New("unknown flag: --invalid"),
			expected: true,
		},
		{
			name:     "unknown shorthand flag",
			err:      errors.New("unknown shorthand flag: 'x' in -x"),
			expected: true,
		},
		{
			name:     "flag needs argument",
			err:      errors.New("flag needs an argument: --config"),
			expected: true,
		},
		{
			name:     "required flag",
			err:      errors.New("--url is required"),
			expected: true,
		},
		{
			name:     "either/or required",
			err:      errors.New("either --id or both --domain and --key are required"),
			expected: true,
		},
		{
			name:     "at least one must be specified",
			err:      errors.New("at least one of --url or --key must be specified"),
			expected: true,
		},
		{
			name:     "flag requires another",
			err:      errors.New("--desc requires --sort-by to be specified"),
			expected: true,
		},
		{
			name:     "accepts at most N args",
			err:      errors.New("accepts at most 1 arg(s), received 2"),
			expected: true,
		},
		{
			name:     "requires at least N args",
			err:      errors.New("requires at least 1 arg(s), only received 0"),
			expected: true,
		},
		{
			name:     "API error - not a usage error",
			err:      errors.New("not_found: Link not found"),
			expected: false,
		},
		{
			name:     "network error - not a usage error",
			err:      errors.New("connection refused"),
			expected: false,
		},
		{
			name:     "auth error - not a usage error",
			err:      errors.New("unauthorized: Invalid API key"),
			expected: false,
		},
		{
			name:     "invalid JSON input - not a usage error",
			err:      errors.New("invalid JSON input: unexpected end of JSON input"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUsageError(tt.err)
			if result != tt.expected {
				t.Errorf("IsUsageError(%q) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
