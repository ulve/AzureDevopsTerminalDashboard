package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Organization string `yaml:"organization"`
	Project      string `yaml:"project"`
	PAT          string `yaml:"pat"`
}

// Load loads configuration from file, env vars, or returns defaults
func Load() (*Config, error) {
	cfg := &Config{}

	// Try to load from config file
	configPath := os.Getenv("HOME") + "/.config/azdo-tui/config.yaml"
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables if set
	if org := os.Getenv("AZDO_ORG"); org != "" {
		cfg.Organization = org
	}
	if project := os.Getenv("AZDO_PROJECT"); project != "" {
		cfg.Project = project
	}
	if pat := os.Getenv("AZDO_PAT"); pat != "" {
		cfg.PAT = pat
	}

	// Validate required fields
	if cfg.Organization == "" {
		return nil, fmt.Errorf("organization is required (set AZDO_ORG or add to config file)")
	}
	if cfg.Project == "" {
		return nil, fmt.Errorf("project is required (set AZDO_PROJECT or add to config file)")
	}
	if cfg.PAT == "" {
		return nil, fmt.Errorf("PAT is required (set AZDO_PAT or add to config file)")
	}

	return cfg, nil
}
