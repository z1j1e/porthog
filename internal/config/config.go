package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all porthog configuration.
type Config struct {
	DefaultFormat    string   `yaml:"default_output_format"`
	ColorTheme       string   `yaml:"color_theme"`
	CriticalDenylist []string `yaml:"critical_process_denylist"`
	DefaultColumns   []string `yaml:"default_columns"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		DefaultFormat: "auto",
		ColorTheme:    "auto",
		CriticalDenylist: []string{
			"systemd", "launchd", "init", "csrss.exe", "smss.exe", "wininit.exe",
		},
		DefaultColumns: []string{"proto", "local_addr", "pid", "process", "user", "state"},
	}
}

// Load reads config from file, applies env overrides, and validates.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path := configPath()
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("invalid config %s: %w", path, err)
		}
	}

	applyEnv(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func configPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "porthog", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "porthog", "config.yaml")
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("PORTHOG_FORMAT"); v != "" {
		cfg.DefaultFormat = v
	}
	if v := os.Getenv("PORTHOG_COLOR"); v != "" {
		cfg.ColorTheme = v
	}
	if v := os.Getenv("PORTHOG_DENYLIST"); v != "" {
		cfg.CriticalDenylist = strings.Split(v, ",")
	}
}

// Validate checks config values are valid.
func (c *Config) Validate() error {
	validFormats := map[string]bool{"auto": true, "table": true, "json": true, "plain": true}
	if !validFormats[c.DefaultFormat] {
		return fmt.Errorf("invalid default_output_format: %q (must be auto|table|json|plain)", c.DefaultFormat)
	}
	validThemes := map[string]bool{"auto": true, "always": true, "never": true}
	if !validThemes[c.ColorTheme] {
		return fmt.Errorf("invalid color_theme: %q (must be auto|always|never)", c.ColorTheme)
	}
	return nil
}
