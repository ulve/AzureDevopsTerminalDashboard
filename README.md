# Azure DevOps Terminal Dashboard (ADTD)

A Terminal User Interface (TUI) for monitoring Azure DevOps pull requests and pipelines, built with Go and Bubbletea.

## Overview

Azure DevOps Terminal Dashboard (ADTD) provides a fast, keyboard-driven interface to monitor your Azure DevOps pull requests and pipeline builds directly from your terminal. Stay on top of your team's work without leaving the command line.

## Features

- **Pull Request Tracking**: View active pull requests from multiple repositories
- **Pipeline Monitoring**: Track recent builds and their status in real-time
- **File Diff Viewer**: Examine changed files in pull requests and view diffs
- **Auto-Refresh**: Automatic updates at configurable intervals
- **Keyboard-Driven**: Fast navigation with intuitive keyboard shortcuts
- **Multi-Project Support**: Monitor multiple projects and repositories simultaneously

## Installation

### Prerequisites

- Go 1.21 or higher
- Azure DevOps account and Personal Access Token (PAT)

### From Source

```bash
git clone https://github.com/ulve/AzureDevopsTerminalDashboard.git
cd AzureDevopsTerminalDashboard
go mod download
go build -o adtd ./cmd
```

### Using Go Install

```bash
go install github.com/ulve/AzureDevopsTerminalDashboard/cmd@latest
```

## Configuration

### 1. Create a Personal Access Token (PAT)

1. Go to Azure DevOps → User Settings → Personal Access Tokens
2. Click "New Token"
3. Give it a name and select expiration
4. **Required Scopes**:
   - **Code (Read)** - Required to read pull requests and file changes
   - **Build (Read)** - Required to read pipeline builds and runs
5. Click "Create" and copy the generated token

**Important**: Store your PAT securely. You'll need to set it as an environment variable.

### 2. Set Environment Variable

Set your PAT as an environment variable:

```bash
export AZURE_DEVOPS_PAT="your-personal-access-token"
```

Add this to your `~/.bashrc`, `~/.zshrc`, or equivalent to persist it across sessions.

### 3. Create Configuration File

Create a `.adtd.json` file in your project directory or home directory:

```json
{
  "organization": "your-organization-name",
  "pullRequests": [
    {
      "project": "ProjectName",
      "repository": "RepositoryName"
    },
    {
      "project": "AnotherProject",
      "repository": "AnotherRepo"
    }
  ],
  "pipelines": [
    {
      "project": "ProjectName",
      "pipeline": "PipelineName"
    },
    {
      "project": "AnotherProject",
      "pipeline": "CI-Pipeline"
    }
  ],
  "refreshInterval": 30
}
```

**Configuration Fields**:
- `organization` (required): Your Azure DevOps organization name (from the URL: `dev.azure.com/{organization}`)
- `pullRequests` (optional): Array of repositories to monitor for pull requests
  - `project`: The project name in Azure DevOps
  - `repository`: The repository name within the project
- `pipelines` (optional): Array of pipelines to monitor for builds
  - `project`: The project name in Azure DevOps
  - `pipeline`: The pipeline name (not ID) as shown in Azure DevOps
- `refreshInterval` (optional): Auto-refresh interval in seconds (default: 30)

**Finding Your Configuration Values**:
- **Organization**: From your Azure DevOps URL: `https://dev.azure.com/{organization}`
- **Project**: The project name from your project URL: `https://dev.azure.com/{org}/{project}`
- **Repository**: Navigate to Repos → Files, and the repository name is in the breadcrumb
- **Pipeline**: Navigate to Pipelines, and use the exact name shown in the list

See `.adtd.json.example` for a complete example.

## Usage

### Basic Usage

```bash
# Run with default config file (.adtd.json in current directory)
./adtd

# Specify a custom config file
./adtd /path/to/config.json

# Make sure AZURE_DEVOPS_PAT is set
export AZURE_DEVOPS_PAT="your-pat-token"
./adtd
```

### Dashboard Views

The dashboard has three main views:

1. **Dashboard View**: Shows pull requests and pipeline builds
   - Toggle between PRs and Builds using `Tab`
   - Navigate items with arrow keys
   - Press `Enter` on a PR to view changed files

2. **PR Files View**: Shows files changed in a selected pull request
   - Navigate files with arrow keys
   - Press `Enter` to view the diff for a file
   - Press `Esc` to return to dashboard

3. **File Diff View**: Shows the diff for a selected file
   - Scroll with arrow keys or Page Up/Down
   - Press `Esc` to return to files list

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `q` or `Ctrl+C` | Quit the application |
| `Tab` | Switch between Pull Requests and Builds tabs (Dashboard view) |
| `↑/↓` | Navigate up/down in lists |
| `Enter` | Select item / View details |
| `Esc` | Go back to previous view |
| `r` | Manually refresh data |
| `PgUp/PgDn` | Scroll in diff view |

## Development

### Building

```bash
# Build the application
go build -o adtd ./cmd

# Run directly
go run ./cmd
```

### Testing

```bash
go test ./...
```

### Project Structure

```
.
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── azuredevops/         # Azure DevOps API client
│   │   └── client.go        # API implementation
│   ├── config/              # Configuration management
│   │   └── config.go        # Config loading and validation
│   └── ui/                  # Bubbletea UI
│       ├── model.go         # Main UI model and state
│       ├── commands.go      # Async commands for data loading
│       └── items.go         # List items and rendering
├── .adtd.json.example       # Example configuration file
├── go.mod                   # Go module definition
└── README.md
```

### Dependencies

This project uses:
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI components

The Azure DevOps API client is implemented directly without external dependencies.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Roadmap

- [x] Pull request monitoring
- [x] Pipeline build monitoring
- [x] File diff viewer
- [x] Auto-refresh functionality
- [x] Multi-project support
- [ ] Better diff rendering (syntax highlighting)
- [ ] Pipeline triggering
- [ ] PR commenting
- [ ] Work item integration
- [ ] Custom themes
- [ ] Dashboard layout customization
- [ ] Export capabilities

## Support

For issues, questions, or suggestions, please open an issue on GitHub.
