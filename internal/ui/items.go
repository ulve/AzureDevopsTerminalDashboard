package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/ulve/azuredevops-terminal-dashboard/internal/azuredevops"
)

// prItem wraps a PullRequest for use in a list
type prItem struct {
	pr azuredevops.PullRequest
}

func (i prItem) FilterValue() string {
	return i.pr.Title
}

func (i prItem) Title() string {
	draftIndicator := ""
	if i.pr.IsDraft {
		draftIndicator = "[DRAFT] "
	}
	return fmt.Sprintf("%sPR #%d: %s", draftIndicator, i.pr.ID, i.pr.Title)
}

func (i prItem) Description() string {
	branch := strings.TrimPrefix(i.pr.SourceRefName, "refs/heads/")
	targetBranch := strings.TrimPrefix(i.pr.TargetRefName, "refs/heads/")
	return fmt.Sprintf("%s/%s | %s → %s | by %s",
		i.pr.Repository.Project.Name,
		i.pr.Repository.Name,
		branch,
		targetBranch,
		i.pr.CreatedBy.DisplayName)
}

// buildItem wraps a Build for use in a list
type buildItem struct {
	build azuredevops.Build
}

func (i buildItem) FilterValue() string {
	return i.build.BuildNumber
}

func (i buildItem) Title() string {
	status := i.build.Status
	if i.build.Result != "" {
		status = i.build.Result
	}

	statusIcon := getStatusIcon(status)

	// Show the actual build name from DevOps (which includes PR description, etc.)
	return fmt.Sprintf("%s %s", statusIcon, i.build.BuildNumber)
}

func (i buildItem) Description() string {
	branch := strings.TrimPrefix(i.build.SourceBranch, "refs/heads/")

	// Use Result if available (succeeded, failed), otherwise use Status (inProgress, etc.)
	status := i.build.Status
	if i.build.Result != "" {
		status = i.build.Result
	}

	// Format time with date (yyyy-mm-dd) before timestamp
	timeStr := ""
	if !i.build.StartTime.IsZero() {
		timeStr = i.build.StartTime.Format("2006-01-02 15:04:05")
	} else if !i.build.QueueTime.IsZero() {
		timeStr = "Queued at " + i.build.QueueTime.Format("2006-01-02 15:04:05")
	}

	return fmt.Sprintf("Status: %s | Branch: %s | %s | by %s",
		getColoredStatus(status),
		branch,
		timeStr,
		i.build.RequestedFor.DisplayName)
}

// fileItem wraps a file path for use in a list
type fileItem struct {
	path string
}

func (i fileItem) FilterValue() string {
	return i.path
}

func (i fileItem) Title() string {
	return i.path
}

func (i fileItem) Description() string {
	return ""
}

// getStatusIcon returns a colored icon for a build status
func getStatusIcon(status string) string {
	var style lipgloss.Style
	var icon string

	switch strings.ToLower(status) {
	case "succeeded":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
		icon = "✓"
	case "failed":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
		icon = "✗"
	case "inprogress":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
		icon = "●"
	case "notstarted":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		icon = "●"
	case "canceled", "cancelled":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		icon = "○"
	case "partiallysucceeded":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
		icon = "◐"
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		icon = "◯"
	}

	return style.Render(icon)
}

// getColoredStatus returns a colored status string
func getColoredStatus(status string) string {
	var style lipgloss.Style

	switch strings.ToLower(status) {
	case "succeeded":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Green
	case "failed":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red
	case "inprogress":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true) // Yellow
	case "notstarted":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	case "canceled", "cancelled":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	case "partiallysucceeded":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true) // Yellow
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	}

	return style.Render(status)
}

// updateLists updates the list items with current data
func (m *Model) updateLists() {
	// Update PR list
	prItems := make([]list.Item, len(m.pullRequests))
	for i, pr := range m.pullRequests {
		prItems[i] = prItem{pr: pr}
	}
	m.prList.SetItems(prItems)

	// Update build list
	buildItems := make([]list.Item, len(m.builds))
	for i, build := range m.builds {
		buildItems[i] = buildItem{build: build}
	}
	m.buildList.SetItems(buildItems)
}

// updateFileList updates the file list with current PR files
func (m *Model) updateFileList() {
	fileItems := make([]list.Item, len(m.prFiles))
	for i, file := range m.prFiles {
		fileItems[i] = fileItem{path: file}
	}
	m.fileList.SetItems(fileItems)
}
