package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#0078D4") // Azure blue
	successColor   = lipgloss.Color("#10B981")
	errorColor     = lipgloss.Color("#EF4444")
	warningColor   = lipgloss.Color("#F59E0B")
	infoColor      = lipgloss.Color("#3B82F6")
	mutedColor     = lipgloss.Color("#6B7280")
	highlightColor = lipgloss.Color("#8B5CF6")

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Bold(true)

	// Status styles
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	inProgressStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			Bold(true)

	// List styles
	listItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedListItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(primaryColor)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Progress bar styles
	progressBarStyle = lipgloss.NewStyle().
				Foreground(successColor)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(mutedColor)
)

// GetStatusStyle returns the appropriate style for a status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "succeeded", "Succeeded", "completed":
		return successStyle
	case "failed", "Failed":
		return errorStyle
	case "inProgress", "InProgress", "running":
		return inProgressStyle
	case "canceled", "Cancelled":
		return warningStyle
	default:
		return lipgloss.NewStyle()
	}
}

// RenderProgressBar renders a simple text-based progress bar
func RenderProgressBar(completed, total int, width int) string {
	if total == 0 {
		return progressEmptyStyle.Render("[" + lipgloss.PlaceHorizontal(width-2, lipgloss.Left, "N/A") + "]")
	}

	percentage := float64(completed) / float64(total)
	filledWidth := int(float64(width-2) * percentage)

	filled := ""
	for i := 0; i < filledWidth; i++ {
		filled += "█"
	}

	empty := ""
	for i := 0; i < (width-2)-filledWidth; i++ {
		empty += "░"
	}

	return "[" +
		progressBarStyle.Render(filled) +
		progressEmptyStyle.Render(empty) +
		"]"
}
