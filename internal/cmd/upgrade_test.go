// internal/cmd/upgrade_test.go
package cmd

import (
	"bytes"
	"testing"
)

func TestUpgradeCmd_Exists(t *testing.T) {
	cmd := newUpgradeCmd()
	if cmd == nil {
		t.Fatal("expected upgrade command to exist")
	}
	if cmd.Use != "upgrade" {
		t.Errorf("expected Use to be 'upgrade', got %q", cmd.Use)
	}
}

func TestUpgradeCmd_HasCheckFlag(t *testing.T) {
	cmd := newUpgradeCmd()
	flag := cmd.Flags().Lookup("check")
	if flag == nil {
		t.Fatal("expected --check flag to exist")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected --check default to be 'false', got %q", flag.DefValue)
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.0.0", "v1.0.0"},
		{"1.0.0", "v1.0.0"},
		{"v0.1.0", "v0.1.0"},
		{"0.1.0", "v0.1.0"},
		{"dev", "dev"},
		{"", "dev"},
		{"v1.2.3-beta", "v1.2.3-beta"},
		{"1.2.3-beta", "v1.2.3-beta"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeVersion(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildAssetName(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "version with v prefix",
			version:  "v1.0.0",
			expected: "dub-cli_1.0.0_",
		},
		{
			name:     "version without v prefix",
			version:  "1.0.0",
			expected: "dub-cli_1.0.0_",
		},
		{
			name:     "prerelease version",
			version:  "v1.0.0-beta.1",
			expected: "dub-cli_1.0.0-beta.1_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAssetName(tt.version)
			// Just check the prefix since OS/arch varies
			if len(result) < len(tt.expected) {
				t.Errorf("buildAssetName(%q) = %q, expected to start with %q", tt.version, result, tt.expected)
			}
			if result[:len(tt.expected)] != tt.expected {
				t.Errorf("buildAssetName(%q) = %q, expected to start with %q", tt.version, result, tt.expected)
			}
		})
	}
}

func TestUpgradeCmd_DevVersion(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() { Version = originalVersion }()

	// Set to dev version
	Version = "dev"

	cmd := newUpgradeCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected some output")
	}
	expected := "Cannot upgrade development builds"
	if !bytes.Contains(out.Bytes(), []byte(expected)) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}
