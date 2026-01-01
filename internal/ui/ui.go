// Package ui provides terminal color output using termenv.
// It respects the --color flag (auto|always|never) and auto-detects
// terminal capabilities.
package ui

import (
	"os"
	"sync"

	"github.com/muesli/termenv"
)

var (
	output    *termenv.Output
	initOnce  sync.Once
	colorMode string = "auto"
)

// Init configures color output based on the --color flag value.
// Valid values: "auto", "always", "never".
// Should be called once during CLI startup.
func Init(color string) {
	initOnce.Do(func() {
		colorMode = color
		output = createOutput(color)
	})
}

// createOutput creates a termenv.Output based on color mode.
func createOutput(color string) *termenv.Output {
	switch color {
	case "always":
		return termenv.NewOutput(os.Stdout, termenv.WithProfile(termenv.TrueColor))
	case "never":
		return termenv.NewOutput(os.Stdout, termenv.WithProfile(termenv.Ascii))
	default: // "auto"
		return termenv.NewOutput(os.Stdout)
	}
}

// getOutput returns the configured output, initializing with defaults if needed.
func getOutput() *termenv.Output {
	if output == nil {
		Init("auto")
	}
	return output
}

// Success returns text styled in green for success messages.
func Success(text string) string {
	return getOutput().String(text).Foreground(getOutput().Color("2")).String()
}

// Error returns text styled in red for error messages.
func Error(text string) string {
	return getOutput().String(text).Foreground(getOutput().Color("1")).String()
}

// Warning returns text styled in yellow for warning messages.
func Warning(text string) string {
	return getOutput().String(text).Foreground(getOutput().Color("3")).String()
}

// Info returns text styled in blue for informational messages.
func Info(text string) string {
	return getOutput().String(text).Foreground(getOutput().Color("4")).String()
}

// Bold returns text in bold.
func Bold(text string) string {
	return getOutput().String(text).Bold().String()
}

// Dim returns text with reduced intensity.
func Dim(text string) string {
	return getOutput().String(text).Faint().String()
}

// Cyan returns text styled in cyan.
func Cyan(text string) string {
	return getOutput().String(text).Foreground(getOutput().Color("6")).String()
}

// Magenta returns text styled in magenta.
func Magenta(text string) string {
	return getOutput().String(text).Foreground(getOutput().Color("5")).String()
}

// Underline returns underlined text.
func Underline(text string) string {
	return getOutput().String(text).Underline().String()
}

// Italic returns italicized text.
func Italic(text string) string {
	return getOutput().String(text).Italic().String()
}

// ColorMode returns the current color mode.
func ColorMode() string {
	return colorMode
}

// HasColors returns true if the terminal supports colors.
func HasColors() bool {
	o := getOutput()
	return o.Profile != termenv.Ascii
}

// Reset resets the output state. Useful for testing.
func Reset() {
	output = nil
	colorMode = "auto"
	initOnce = sync.Once{}
}
