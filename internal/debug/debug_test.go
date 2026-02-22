package debug

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name        string
		debug       bool
		wantEnabled bool
	}{
		{"debug enabled", true, true},
		{"debug disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetForTesting()
			Init(tt.debug)
			if got := Enabled(); got != tt.wantEnabled {
				t.Errorf("Enabled() = %v, want %v", got, tt.wantEnabled)
			}
		})
	}
}

func TestLog(t *testing.T) {
	// Capture stderr
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	// Enable debug
	enabled.Store(true)

	Log("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Log() output should contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Log() output should contain 'key=value', got: %s", output)
	}

	// Disable debug
	buf.Reset()
	enabled.Store(false)
	Log("should not appear")
	if buf.Len() > 0 {
		t.Errorf("Log() should not output when disabled, got: %s", buf.String())
	}
}

func TestRequest(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	enabled.Store(true)
	Request("GET", "https://api.dub.co/links")

	output := buf.String()
	if !strings.Contains(output, "api request") {
		t.Errorf("Request() output should contain 'api request', got: %s", output)
	}
	if !strings.Contains(output, "method=GET") {
		t.Errorf("Request() output should contain 'method=GET', got: %s", output)
	}
	if !strings.Contains(output, "url=https://api.dub.co/links") {
		t.Errorf("Request() output should contain URL, got: %s", output)
	}
}

func TestResponse(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	enabled.Store(true)
	Response(200, 150*time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "api response") {
		t.Errorf("Response() output should contain 'api response', got: %s", output)
	}
	if !strings.Contains(output, "status=200") {
		t.Errorf("Response() output should contain 'status=200', got: %s", output)
	}
	if !strings.Contains(output, "duration=") {
		t.Errorf("Response() output should contain duration, got: %s", output)
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	enabled.Store(true)
	Info("info message", "foo", "bar")

	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("Info() output should contain 'info message', got: %s", output)
	}
	if !strings.Contains(output, "foo=bar") {
		t.Errorf("Info() output should contain 'foo=bar', got: %s", output)
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	enabled.Store(true)
	Error("error message", "err", "something went wrong")

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("Error() output should contain 'error message', got: %s", output)
	}
	if !strings.Contains(output, "something went wrong") {
		t.Errorf("Error() output should contain error details, got: %s", output)
	}
}

func TestLogDisabledNoOutput(t *testing.T) {
	// Restore stderr after test
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()

	// Create a pipe to capture stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	resetForTesting()
	Init(false)

	Log("should not appear")
	Request("GET", "https://example.com")
	Response(200, time.Second)
	Info("should not appear")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	if strings.Contains(buf.String(), "should not appear") {
		t.Error("disabled debug should not output debug/info messages")
	}
}
