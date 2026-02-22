package ui

import (
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name  string
		color string
	}{
		{"auto mode", "auto"},
		{"always mode", "always"},
		{"never mode", "never"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Reset()
			Init(tt.color)
			if ColorMode() != tt.color {
				t.Errorf("ColorMode() = %q, want %q", ColorMode(), tt.color)
			}
		})
	}
}

func TestColorFunctions(t *testing.T) {
	// Test with colors disabled to get predictable output
	Reset()
	Init("never")

	tests := []struct {
		name string
		fn   func(string) string
		text string
	}{
		{"Success", Success, "success message"},
		{"Error", Error, "error message"},
		{"Warning", Warning, "warning message"},
		{"Info", Info, "info message"},
		{"Bold", Bold, "bold text"},
		{"Dim", Dim, "dim text"},
		{"Cyan", Cyan, "cyan text"},
		{"Magenta", Magenta, "magenta text"},
		{"Underline", Underline, "underlined text"},
		{"Italic", Italic, "italic text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.text)
			// With "never" mode, text should be returned without ANSI codes
			if !strings.Contains(result, tt.text) {
				t.Errorf("%s(%q) = %q, should contain %q", tt.name, tt.text, result, tt.text)
			}
		})
	}
}

func TestColorsEnabled(t *testing.T) {
	Reset()
	Init("always")

	// With "always" mode, Success should contain ANSI escape codes
	result := Success("test")
	if !strings.Contains(result, "\x1b[") {
		t.Errorf("Success with always mode should contain ANSI codes, got %q", result)
	}
}

func TestColorsDisabled(t *testing.T) {
	Reset()
	Init("never")

	// With "never" mode, Success should NOT contain ANSI escape codes
	result := Success("test")
	if strings.Contains(result, "\x1b[") {
		t.Errorf("Success with never mode should not contain ANSI codes, got %q", result)
	}
	if result != "test" {
		t.Errorf("Success with never mode = %q, want %q", result, "test")
	}
}

func TestHasColors(t *testing.T) {
	Reset()
	Init("never")
	if HasColors() {
		t.Error("HasColors() should be false with never mode")
	}

	Reset()
	Init("always")
	if !HasColors() {
		t.Error("HasColors() should be true with always mode")
	}
}

func TestDefaultInit(t *testing.T) {
	Reset()
	// Call a color function without explicit Init
	result := Success("test")
	if !strings.Contains(result, "test") {
		t.Errorf("Success without Init should still work, got %q", result)
	}
}
