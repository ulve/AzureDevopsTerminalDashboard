package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ulve/azuredevops-terminal-dashboard/api"
	"github.com/ulve/azuredevops-terminal-dashboard/config"
	"github.com/ulve/azuredevops-terminal-dashboard/ui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nPlease set the required environment variables:")
		fmt.Fprintln(os.Stderr, "  export AZDO_ORG=your-organization")
		fmt.Fprintln(os.Stderr, "  export AZDO_PROJECT=your-project")
		fmt.Fprintln(os.Stderr, "  export AZDO_PAT=your-personal-access-token")
		fmt.Fprintln(os.Stderr, "\nOr create a config file at ~/.config/azdo-tui/config.yaml:")
		fmt.Fprintln(os.Stderr, "  organization: your-organization")
		fmt.Fprintln(os.Stderr, "  project: your-project")
		fmt.Fprintln(os.Stderr, "  pat: your-personal-access-token")
		os.Exit(1)
	}

	// Create API client
	client, err := api.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Azure DevOps client: %v\n", err)
		os.Exit(1)
	}

	// Create and run the TUI application
	model := ui.NewModel(client)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
