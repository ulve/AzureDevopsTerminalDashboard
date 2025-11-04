package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ulve/azuredevops-terminal-dashboard/internal/azuredevops"
	"github.com/ulve/azuredevops-terminal-dashboard/internal/config"
	"github.com/ulve/azuredevops-terminal-dashboard/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	configPath := ".adtd.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Get PAT from environment
	pat := os.Getenv("AZURE_DEVOPS_PAT")
	if pat == "" {
		return fmt.Errorf("AZURE_DEVOPS_PAT environment variable is not set")
	}

	// Create Azure DevOps client
	client := azuredevops.NewClient(cfg.Organization, pat)

	// Create and run the UI
	model := ui.NewModel(cfg, client)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
