package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// PullRequestConfig represents a single pull request source
type PullRequestConfig struct {
	Project    string `json:"project"`
	Repository string `json:"repository"`
}

// PipelineConfig represents a single pipeline source
type PipelineConfig struct {
	Project      string `json:"project"`
	Pipeline     string `json:"pipeline"`     // Pipeline name (optional if DefinitionID is provided)
	DefinitionID int    `json:"definitionId"` // Pipeline definition ID (optional if Pipeline is provided)
}

// Config represents the application configuration
type Config struct {
	Organization    string              `json:"organization"`
	PullRequests    []PullRequestConfig `json:"pullRequests"`
	Pipelines       []PipelineConfig    `json:"pipelines"`
	RefreshInterval int                 `json:"refreshInterval"` // in seconds
}

// Load loads the configuration from a file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set default refresh interval if not specified
	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = 30
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Organization == "" {
		return fmt.Errorf("organization is required")
	}

	if len(c.PullRequests) == 0 && len(c.Pipelines) == 0 {
		return fmt.Errorf("at least one pull request or pipeline must be configured")
	}

	for i, pr := range c.PullRequests {
		if pr.Project == "" {
			return fmt.Errorf("pull request %d: project is required", i)
		}
		if pr.Repository == "" {
			return fmt.Errorf("pull request %d: repository is required", i)
		}
	}

	for i, p := range c.Pipelines {
		if p.Project == "" {
			return fmt.Errorf("pipeline %d: project is required", i)
		}
		if p.Pipeline == "" && p.DefinitionID == 0 {
			return fmt.Errorf("pipeline %d: either pipeline name or definitionId is required", i)
		}
	}

	return nil
}
