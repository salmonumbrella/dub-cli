// internal/cmd/errors.go
package cmd

import (
	"errors"
	"fmt"
	"strings"
)

// UsageError represents an error caused by incorrect command usage,
// such as missing required flags, invalid flag values, or unknown commands.
// Commands returning UsageError will cause the CLI to exit with code 2.
type UsageError struct {
	Err error
}

func (e *UsageError) Error() string {
	return e.Err.Error()
}

func (e *UsageError) Unwrap() error {
	return e.Err
}

// NewUsageError wraps an error as a usage error.
func NewUsageError(err error) *UsageError {
	return &UsageError{Err: err}
}

// NewUsageErrorf creates a usage error from a formatted string.
func NewUsageErrorf(format string, args ...interface{}) *UsageError {
	return &UsageError{Err: fmt.Errorf(format, args...)}
}

// IsUsageError checks if an error is a usage error (either our custom type
// or a Cobra-generated usage error based on error message patterns).
func IsUsageError(err error) bool {
	if err == nil {
		return false
	}

	// Check for our custom UsageError type
	var usageErr *UsageError
	if errors.As(err, &usageErr) {
		return true
	}

	// Check for Cobra-generated usage errors by message patterns
	msg := err.Error()
	usagePatterns := []string{
		// Cobra-generated errors
		"unknown command",
		"unknown flag",
		"unknown shorthand flag",
		"flag needs an argument",
		"invalid argument",
		"required flag(s)",
		"accepts at most",
		"accepts between",
		"requires at least",
		"accepts at least",
		"accepts exactly",
	}

	for _, pattern := range usagePatterns {
		if strings.Contains(strings.ToLower(msg), pattern) {
			return true
		}
	}

	// Check for CLI validation patterns (missing required flags, etc.)
	if isValidationError(msg) {
		return true
	}

	return false
}

// isValidationError detects command-level validation errors that indicate
// incorrect usage (missing required flags, invalid flag combinations, etc.)
func isValidationError(msg string) bool {
	lowerMsg := strings.ToLower(msg)

	// Pattern: "--flag is required"
	if strings.Contains(msg, "--") && strings.Contains(lowerMsg, "is required") {
		return true
	}

	// Pattern: "either --X or --Y are required"
	if strings.Contains(lowerMsg, "either") && strings.Contains(msg, "--") && strings.Contains(lowerMsg, "required") {
		return true
	}

	// Pattern: "at least one of --X must be specified"
	if strings.Contains(lowerMsg, "at least one") && strings.Contains(msg, "--") && strings.Contains(lowerMsg, "must be specified") {
		return true
	}

	// Pattern: "--X requires --Y"
	if strings.Contains(msg, "--") && strings.Contains(lowerMsg, "requires") && !strings.Contains(lowerMsg, "requires at least") {
		return true
	}

	return false
}
