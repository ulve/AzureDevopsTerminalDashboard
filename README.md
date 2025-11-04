# Azure DevOps Terminal Dashboard

A Terminal User Interface (TUI) for Azure DevOps built with Go.

## Overview

Azure DevOps Terminal Dashboard provides a fast, keyboard-driven interface to interact with Azure DevOps directly from your terminal. Manage work items, view pipelines, browse repositories, and track pull requests without leaving the command line.

## Features

- **Work Items Management**: View, search, and update work items
- **Pipeline Monitoring**: Track build and release pipelines in real-time
- **Repository Browser**: Navigate repositories and view file contents
- **Pull Request Tracking**: Review and manage pull requests
- **Sprint Dashboard**: Visualize sprint progress and team velocity
- **Keyboard-Driven**: Fast navigation with intuitive keyboard shortcuts

## Installation

### Prerequisites

- Go 1.21 or higher
- Azure DevOps account and Personal Access Token (PAT)

### From Source

```bash
git clone https://github.com/ulve/AzureDevopsTerminalDashboard.git
cd AzureDevopsTerminalDashboard
go build -o azdo-tui
```

### Using Go Install

```bash
go install github.com/ulve/AzureDevopsTerminalDashboard@latest
```

## Configuration

Create a configuration file at `~/.config/azdo-tui/config.yaml`:

```yaml
organization: your-org-name
project: your-project-name
pat: your-personal-access-token
```

Alternatively, set environment variables:

```bash
export AZDO_ORG="your-org-name"
export AZDO_PROJECT="your-project-name"
export AZDO_PAT="your-personal-access-token"
```

### Creating a Personal Access Token

1. Go to Azure DevOps → User Settings → Personal Access Tokens
2. Click "New Token"
3. Select appropriate scopes (Work Items: Read & Write, Code: Read, Build: Read)
4. Copy the generated token

## Usage

```bash
# Launch the TUI
azdo-tui

# Specify organization and project
azdo-tui --org myorg --project myproject

# Show version
azdo-tui --version
```

## Keyboard Shortcuts

### Pipeline List View
| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `Enter` | View pipeline details |
| `r` | Refresh pipeline list |
| `q` | Quit |

### Pipeline Detail View
| Key | Action |
|-----|--------|
| `↑/↓` | Scroll up/down |
| `PgUp/PgDn` | Page up/down |
| `Esc` | Back to pipeline list |
| `r` | Refresh pipeline details |
| `q` | Quit |

## Pipeline Progress & Logs

When you select a pipeline (press Enter), you'll see:

- **Pipeline Status**: Current status (InProgress, Succeeded, Failed, etc.)
- **Branch & User Info**: Source branch and who requested the build
- **Duration**: How long the pipeline has been running
- **Stages & Jobs**: Real-time progress of each stage and job
- **Recent Logs**: Last 50 lines of pipeline logs

The view auto-refreshes every 10 seconds to show live progress updates.

## Development

### Building

```bash
go build -o azdo-tui
```

### Testing

```bash
go test ./...
```

### Dependencies

This project uses:
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [azure-devops-go-api](https://github.com/microsoft/azure-devops-go-api) - Azure DevOps API client

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Roadmap

- [ ] Work item creation and editing
- [ ] Pipeline triggering and cancellation
- [ ] Branch management
- [ ] Custom queries and filters
- [ ] Multi-project support
- [ ] Dashboard customization
- [ ] Export capabilities

## Support

For issues, questions, or suggestions, please open an issue on GitHub.
