// internal/outfmt/format_test.go
package outfmt

import (
	"bytes"
	"testing"
)

func TestFormatJSON(t *testing.T) {
	data := map[string]string{"id": "123", "url": "https://dub.sh/test"}
	buf := new(bytes.Buffer)

	err := FormatJSON(buf, data, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte(`"id"`)) {
		t.Errorf("expected JSON output, got: %s", output)
	}
}

func TestFormatJSON_WithQuery(t *testing.T) {
	data := map[string]string{"id": "123", "url": "https://dub.sh/test"}
	buf := new(bytes.Buffer)

	err := FormatJSON(buf, data, ".id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "\"123\"\n" {
		t.Errorf("expected '\"123\"\\n', got: %q", output)
	}
}
