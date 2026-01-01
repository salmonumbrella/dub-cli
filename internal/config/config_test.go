// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_SaveLoad(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Create and save config
	cfg := &Config{DefaultWorkspace: "my-workspace"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(tmpDir, ".config", AppName, "config.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("config file not created at %s", expectedPath)
	}

	// Load config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.DefaultWorkspace != "my-workspace" {
		t.Errorf("expected DefaultWorkspace %q, got %q", "my-workspace", loaded.DefaultWorkspace)
	}
}

func TestConfig_LoadNonExistent(t *testing.T) {
	// Create a temp directory with no config
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Loading should return empty config, not error
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DefaultWorkspace != "" {
		t.Errorf("expected empty DefaultWorkspace, got %q", cfg.DefaultWorkspace)
	}
}

func TestGetDefaultWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// No config - should return ErrNoDefaultWorkspace
	_, err := GetDefaultWorkspace()
	if err != ErrNoDefaultWorkspace {
		t.Errorf("expected ErrNoDefaultWorkspace, got %v", err)
	}

	// Set default workspace
	if err := SetDefaultWorkspace("production"); err != nil {
		t.Fatalf("failed to set default workspace: %v", err)
	}

	// Now should return the workspace
	ws, err := GetDefaultWorkspace()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws != "production" {
		t.Errorf("expected %q, got %q", "production", ws)
	}
}

func TestSetDefaultWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Set workspace
	if err := SetDefaultWorkspace("staging"); err != nil {
		t.Fatalf("failed to set default workspace: %v", err)
	}

	// Verify it's set
	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.DefaultWorkspace != "staging" {
		t.Errorf("expected %q, got %q", "staging", cfg.DefaultWorkspace)
	}

	// Change to different workspace
	if err := SetDefaultWorkspace("production"); err != nil {
		t.Fatalf("failed to set default workspace: %v", err)
	}

	// Verify change
	cfg, err = Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.DefaultWorkspace != "production" {
		t.Errorf("expected %q, got %q", "production", cfg.DefaultWorkspace)
	}
}

func TestClearDefaultWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Set then clear
	if err := SetDefaultWorkspace("production"); err != nil {
		t.Fatalf("failed to set default workspace: %v", err)
	}

	if err := ClearDefaultWorkspace(); err != nil {
		t.Fatalf("failed to clear default workspace: %v", err)
	}

	// Should now return ErrNoDefaultWorkspace
	_, err := GetDefaultWorkspace()
	if err != ErrNoDefaultWorkspace {
		t.Errorf("expected ErrNoDefaultWorkspace, got %v", err)
	}
}

func TestConfig_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	cfg := &Config{DefaultWorkspace: "test"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Check file permissions (should be 0600 for security)
	path := filepath.Join(tmpDir, ".config", AppName, "config.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("expected file permissions 0600, got %04o", perm)
	}
}
