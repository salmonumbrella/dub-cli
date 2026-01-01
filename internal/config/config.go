// internal/config/config.go
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// ErrNoDefaultWorkspace is returned when no default workspace is configured
var ErrNoDefaultWorkspace = errors.New("no default workspace configured")

// Config represents the CLI configuration stored on disk
type Config struct {
	DefaultWorkspace string `json:"default_workspace,omitempty"`
}

// configPath returns the path to the config file (~/.config/dub-cli/config.json)
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", AppName, "config.json"), nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// GetDefaultWorkspace returns the configured default workspace
// Returns ErrNoDefaultWorkspace if no default is set
func GetDefaultWorkspace() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	if cfg.DefaultWorkspace == "" {
		return "", ErrNoDefaultWorkspace
	}
	return cfg.DefaultWorkspace, nil
}

// SetDefaultWorkspace sets the default workspace
func SetDefaultWorkspace(name string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.DefaultWorkspace = name
	return cfg.Save()
}

// ClearDefaultWorkspace removes the default workspace setting
func ClearDefaultWorkspace() error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.DefaultWorkspace = ""
	return cfg.Save()
}
