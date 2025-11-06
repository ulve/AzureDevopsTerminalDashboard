package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ulve/azuredevops-terminal-dashboard/internal/azuredevops"
	"github.com/ulve/azuredevops-terminal-dashboard/internal/config"
)

// View represents different views in the application
type View int

const (
	ViewDashboard View = iota
	ViewPRDetails
	ViewPRFiles
	ViewFileDiff
	ViewBuildLogs
)

// Model represents the application state
type Model struct {
	config          *config.Config
	client          *azuredevops.Client
	view            View
	pullRequests    []azuredevops.PullRequest
	builds          []azuredevops.Build
	prList          list.Model
	buildList       list.Model
	fileList        list.Model
	diffViewport    viewport.Model
	logsViewport    viewport.Model
	prDetailsViewport viewport.Model
	selectedPR      *azuredevops.PullRequest
	selectedBuild   *azuredevops.Build
	selectedBuildProject string
	prFiles         []string
	currentDiff     string
	currentFilePath string
	buildLogs       string
	loading         bool
	loadingLogs     bool
	err             error
	lastUpdate      time.Time
	autoRefresh     bool
	refreshInterval time.Duration
	width           int
	height          int
	activeTab       int // 0 = PRs, 1 = Builds
}

// TickMsg represents a timer tick for auto-refresh
type TickMsg time.Time

// DataLoadedMsg represents loaded data
type DataLoadedMsg struct {
	pullRequests []azuredevops.PullRequest
	builds       []azuredevops.Build
	err          error
}

// FilesLoadedMsg represents loaded PR files
type FilesLoadedMsg struct {
	files []string
	err   error
}

// DiffLoadedMsg represents loaded file diff
type DiffLoadedMsg struct {
	diff     string
	filePath string
	err      error
}

// LogsLoadedMsg represents loaded build logs
type LogsLoadedMsg struct {
	logs string
	err  error
}

// NewModel creates a new application model
func NewModel(cfg *config.Config, client *azuredevops.Client) Model {
	// Create PR list
	prDelegate := list.NewDefaultDelegate()
	prList := list.New([]list.Item{}, prDelegate, 0, 0)
	prList.Title = "Pull Requests"
	prList.SetShowStatusBar(false)
	prList.SetFilteringEnabled(false)

	// Create build list
	buildDelegate := list.NewDefaultDelegate()
	buildList := list.New([]list.Item{}, buildDelegate, 0, 0)
	buildList.Title = "Pipeline Builds"
	buildList.SetShowStatusBar(false)
	buildList.SetFilteringEnabled(false)

	// Create file list
	fileDelegate := list.NewDefaultDelegate()
	fileList := list.New([]list.Item{}, fileDelegate, 0, 0)
	fileList.Title = "Changed Files"
	fileList.SetShowStatusBar(false)
	fileList.SetFilteringEnabled(false)

	// Create diff viewport
	diffViewport := viewport.New(0, 0)

	// Create logs viewport
	logsViewport := viewport.New(0, 0)

	// Create PR details viewport
	prDetailsViewport := viewport.New(0, 0)

	return Model{
		config:          cfg,
		client:          client,
		view:            ViewDashboard,
		prList:          prList,
		buildList:       buildList,
		fileList:        fileList,
		diffViewport:    diffViewport,
		logsViewport:    logsViewport,
		prDetailsViewport: prDetailsViewport,
		loading:         true,
		autoRefresh:     true,
		refreshInterval: time.Duration(cfg.RefreshInterval) * time.Second,
		activeTab:       0,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadData(),
		m.tickCmd(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "r":
			// Manual refresh
			m.loading = true
			return m, m.loadData()

		case "tab":
			// Switch between tabs in dashboard view
			if m.view == ViewDashboard {
				m.activeTab = (m.activeTab + 1) % 2
			}

		case "enter":
			return m.handleEnter()

		case "h", "left":
			// Go back to previous view
			switch m.view {
			case ViewPRDetails:
				m.view = ViewDashboard
				m.err = nil // Clear errors when going back
			case ViewPRFiles:
				m.view = ViewPRDetails
				m.err = nil // Clear errors when going back
			case ViewFileDiff:
				m.view = ViewPRFiles
				m.err = nil // Clear errors when going back
			case ViewBuildLogs:
				m.view = ViewDashboard
				m.err = nil // Clear errors when going back
			}

		case "g":
			// Open build in browser when in build logs view
			if m.view == ViewBuildLogs && m.selectedBuild != nil {
				return m, m.openBuildURL()
			}
			// Open PR in browser when in PR details view
			if m.view == ViewPRDetails && m.selectedPR != nil {
				return m, m.openPRURL()
			}

		case "c":
			// Clone PR repository when in PR details view
			if m.view == ViewPRDetails && m.selectedPR != nil {
				return m, m.clonePRRepo()
			}
		}

	case TickMsg:
		if m.autoRefresh && time.Since(m.lastUpdate) >= m.refreshInterval {
			cmds = append(cmds, m.loadData())
		}
		cmds = append(cmds, m.tickCmd())

	case DataLoadedMsg:
		m.loading = false
		m.lastUpdate = time.Now()
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.pullRequests = msg.pullRequests
			m.builds = msg.builds
			m.updateLists()
		}

	case FilesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.view = ViewPRFiles
		} else {
			m.err = nil // Clear any previous errors
			m.prFiles = msg.files
			m.updateFileList()
			m.view = ViewPRFiles
		}

	case DiffLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			// Stay in ViewPRFiles so user can see error and try another file
		} else {
			m.err = nil // Clear any previous errors
			m.currentDiff = msg.diff
			m.currentFilePath = msg.filePath
			// Format diff with colors and syntax highlighting
			coloredDiff := m.formatDiff(msg.diff, msg.filePath)
			m.diffViewport.SetContent(coloredDiff)
			m.view = ViewFileDiff
		}

	case LogsLoadedMsg:
		m.loadingLogs = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.buildLogs = msg.logs
			m.logsViewport.SetContent(msg.logs)
			m.view = ViewBuildLogs
		}
	}

	// Update active component based on view
	var cmd tea.Cmd
	switch m.view {
	case ViewDashboard:
		if m.activeTab == 0 {
			m.prList, cmd = m.prList.Update(msg)
		} else {
			m.buildList, cmd = m.buildList.Update(msg)
		}
	case ViewPRDetails:
		m.prDetailsViewport, cmd = m.prDetailsViewport.Update(msg)
	case ViewPRFiles:
		m.fileList, cmd = m.fileList.Update(msg)
	case ViewFileDiff:
		m.diffViewport, cmd = m.diffViewport.Update(msg)
	case ViewBuildLogs:
		m.logsViewport, cmd = m.logsViewport.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if m.loading && len(m.pullRequests) == 0 && len(m.builds) == 0 {
		return "\n  Loading data...\n"
	}

	switch m.view {
	case ViewDashboard:
		return m.renderDashboard()
	case ViewPRDetails:
		return m.renderPRDetails()
	case ViewPRFiles:
		return m.renderPRFiles()
	case ViewFileDiff:
		return m.renderFileDiff()
	case ViewBuildLogs:
		return m.renderBuildLogs()
	}

	return ""
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	activeTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

// renderDashboard renders the main dashboard view
func (m Model) renderDashboard() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("Azure DevOps Dashboard"))
	s.WriteString("\n")

	// Tabs
	prTab := tabStyle.Render(fmt.Sprintf("Pull Requests (%d)", len(m.pullRequests)))
	buildTab := tabStyle.Render(fmt.Sprintf("Builds (%d)", len(m.builds)))

	if m.activeTab == 0 {
		prTab = activeTabStyle.Render(fmt.Sprintf("Pull Requests (%d)", len(m.pullRequests)))
	} else {
		buildTab = activeTabStyle.Render(fmt.Sprintf("Builds (%d)", len(m.builds)))
	}

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, prTab, "  ", buildTab))
	s.WriteString("\n\n")

	// Content
	if m.activeTab == 0 {
		s.WriteString(m.prList.View())
	} else {
		s.WriteString(m.buildList.View())
	}

	// Status bar
	var statusText string
	if m.activeTab == 0 {
		statusText = fmt.Sprintf("Last update: %s | Auto-refresh: %v | Press 'r' to refresh, 'tab' to switch, 'enter' to view PR details, 'q' to quit",
			m.lastUpdate.Format("15:04:05"), m.autoRefresh)
	} else {
		statusText = fmt.Sprintf("Last update: %s | Auto-refresh: %v | Press 'r' to refresh, 'tab' to switch, 'enter' to view build logs, 'q' to quit",
			m.lastUpdate.Format("15:04:05"), m.autoRefresh)
	}
	s.WriteString("\n")
	s.WriteString(statusStyle.Render(statusText))

	if m.err != nil {
		s.WriteString("\n")
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return s.String()
}

// renderPRFiles renders the PR files view
func (m Model) renderPRFiles() string {
	var s strings.Builder

	if m.selectedPR != nil {
		s.WriteString(titleStyle.Render(fmt.Sprintf("PR #%d: %s", m.selectedPR.ID, m.selectedPR.Title)))
		s.WriteString("\n\n")
	}

	s.WriteString(m.fileList.View())
	s.WriteString("\n")
	s.WriteString(statusStyle.Render("Press 'enter' to view diff, 'h' or left arrow to go back, 'q' to quit"))

	if m.err != nil {
		s.WriteString("\n")
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return s.String()
}

// renderFileDiff renders the file diff view
func (m Model) renderFileDiff() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("File Diff"))
	s.WriteString("\n\n")
	s.WriteString(m.diffViewport.View())
	s.WriteString("\n")
	s.WriteString(statusStyle.Render("Press 'h' or left arrow to go back, 'q' to quit"))

	return s.String()
}

// renderPRDetails renders the pull request details view
func (m Model) renderPRDetails() string {
	var s strings.Builder

	if m.selectedPR == nil {
		return "No PR selected"
	}

	pr := m.selectedPR

	// Title with PR ID
	draftIndicator := ""
	if pr.IsDraft {
		draftIndicator = " [DRAFT]"
	}
	s.WriteString(titleStyle.Render(fmt.Sprintf("PR #%d: %s%s", pr.ID, pr.Title, draftIndicator)))
	s.WriteString("\n\n")

	// PR Details in a formatted viewport
	var details strings.Builder

	// Basic info section
	details.WriteString(lipgloss.NewStyle().Bold(true).Render("Status: "))
	statusColor := "green"
	switch pr.Status {
	case "completed":
		statusColor = "green"
	case "active":
		statusColor = "blue"
	case "abandoned":
		statusColor = "red"
	}
	details.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true).Render(pr.Status))
	details.WriteString("\n\n")

	details.WriteString(lipgloss.NewStyle().Bold(true).Render("Created by: "))
	details.WriteString(pr.CreatedBy.DisplayName)
	details.WriteString("\n")

	details.WriteString(lipgloss.NewStyle().Bold(true).Render("Created: "))
	details.WriteString(pr.CreationDate.Format("2006-01-02 15:04:05"))
	details.WriteString("\n\n")

	// Branch information
	sourceBranch := strings.TrimPrefix(pr.SourceRefName, "refs/heads/")
	targetBranch := strings.TrimPrefix(pr.TargetRefName, "refs/heads/")
	details.WriteString(lipgloss.NewStyle().Bold(true).Render("Branches: "))
	details.WriteString(fmt.Sprintf("%s → %s\n\n", sourceBranch, targetBranch))

	// Repository information
	details.WriteString(lipgloss.NewStyle().Bold(true).Render("Project: "))
	details.WriteString(pr.Repository.Project.Name)
	details.WriteString("\n")

	details.WriteString(lipgloss.NewStyle().Bold(true).Render("Repository: "))
	details.WriteString(pr.Repository.Name)
	details.WriteString("\n\n")

	// Description
	details.WriteString(lipgloss.NewStyle().Bold(true).Render("Description:"))
	details.WriteString("\n")
	if pr.Description != "" {
		details.WriteString(pr.Description)
	} else {
		details.WriteString(lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("241")).Render("(No description provided)"))
	}
	details.WriteString("\n")

	m.prDetailsViewport.SetContent(details.String())
	s.WriteString(m.prDetailsViewport.View())
	s.WriteString("\n")
	s.WriteString(statusStyle.Render("Press 'enter' to view files, 'g' to open in browser, 'c' to clone repo, 'h' or left arrow to go back, 'q' to quit"))

	return s.String()
}

// renderBuildLogs renders the build logs view
func (m Model) renderBuildLogs() string {
	var s strings.Builder

	if m.selectedBuild != nil {
		s.WriteString(titleStyle.Render(fmt.Sprintf("Build #%s Logs", m.selectedBuild.BuildNumber)))
		s.WriteString("\n\n")
	}

	if m.loadingLogs {
		s.WriteString("\n  Loading logs...\n")
	} else {
		s.WriteString(m.logsViewport.View())
	}
	s.WriteString("\n")
	s.WriteString(statusStyle.Render("Press 'g' to open in browser, 'h' or left arrow to go back, 'q' to quit"))

	return s.String()
}

// handleEnter handles the enter key press
func (m Model) handleEnter() (Model, tea.Cmd) {
	switch m.view {
	case ViewDashboard:
		if m.activeTab == 0 && len(m.pullRequests) > 0 {
			// Show PR details
			idx := m.prList.Index()
			if idx >= 0 && idx < len(m.pullRequests) {
				m.selectedPR = &m.pullRequests[idx]
				m.view = ViewPRDetails
				return m, nil
			}
		} else if m.activeTab == 1 && len(m.builds) > 0 {
			// Load build logs
			idx := m.buildList.Index()
			if idx >= 0 && idx < len(m.builds) {
				m.selectedBuild = &m.builds[idx]
				m.loadingLogs = true
				return m, m.loadBuildLogs(m.selectedBuild)
			}
		}

	case ViewPRDetails:
		// Navigate to PR files from details view
		if m.selectedPR != nil {
			return m, m.loadPRFiles(m.selectedPR)
		}

	case ViewPRFiles:
		if len(m.prFiles) > 0 {
			// Load file diff
			idx := m.fileList.Index()
			if idx >= 0 && idx < len(m.prFiles) && m.selectedPR != nil {
				filePath := m.prFiles[idx]
				return m, m.loadFileDiff(m.selectedPR, filePath)
			}
		}
	}

	return m, nil
}

// updateSizes updates the sizes of UI components
func (m *Model) updateSizes() {
	listHeight := m.height - 10
	if listHeight < 10 {
		listHeight = 10
	}

	m.prList.SetSize(m.width-4, listHeight)
	m.buildList.SetSize(m.width-4, listHeight)
	m.fileList.SetSize(m.width-4, listHeight)
	m.diffViewport.Width = m.width - 4
	m.diffViewport.Height = m.height - 6
	m.logsViewport.Width = m.width - 4
	m.logsViewport.Height = m.height - 6
	m.prDetailsViewport.Width = m.width - 4
	m.prDetailsViewport.Height = m.height - 8
}

// tickCmd returns a command that sends a tick message
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// openBuildURL opens the build in the default browser
func (m Model) openBuildURL() tea.Cmd {
	return func() tea.Msg {
		if m.selectedBuild == nil {
			return nil
		}

		// Try to determine the project by checking which pipeline this build belongs to
		project := ""
		if len(m.config.Pipelines) > 0 {
			// If we only have one project, use it
			if len(m.config.Pipelines) == 1 {
				project = m.config.Pipelines[0].Project
			} else {
				// Try to match by definition ID
				for _, p := range m.config.Pipelines {
					if p.DefinitionID == m.selectedBuild.Definition.ID {
						project = p.Project
						break
					}
				}
				// If no match found, use the first project
				if project == "" {
					project = m.config.Pipelines[0].Project
				}
			}
		}

		// Construct the Azure DevOps build URL
		url := fmt.Sprintf("https://dev.azure.com/%s/%s/_build/results?buildId=%d",
			m.config.Organization, project, m.selectedBuild.ID)

		// Open URL in default browser based on OS
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			return nil
		}

		_ = cmd.Start()
		return nil
	}
}

// openPRURL opens the pull request in the default browser
func (m Model) openPRURL() tea.Cmd {
	return func() tea.Msg {
		if m.selectedPR == nil {
			return nil
		}

		pr := m.selectedPR
		project := pr.Repository.Project.Name
		repository := pr.Repository.Name

		// Construct the Azure DevOps PR URL
		url := fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s/pullrequest/%d",
			m.config.Organization, project, repository, pr.ID)

		// Open URL in default browser based on OS
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			return nil
		}

		_ = cmd.Start()
		return nil
	}
}

// clonePRRepo clones the PR repository and checks out the source branch
func (m Model) clonePRRepo() tea.Cmd {
	return func() tea.Msg {
		if m.selectedPR == nil {
			return nil
		}

		pr := m.selectedPR
		repository := pr.Repository.Name
		sourceBranch := strings.TrimPrefix(pr.SourceRefName, "refs/heads/")

		// Construct the clone URL
		cloneURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s",
			m.config.Organization, pr.Repository.Project.Name, repository)

		// Clone to current directory with repository name
		cloneCmd := exec.Command("git", "clone", cloneURL, repository)
		if err := cloneCmd.Run(); err != nil {
			return nil
		}

		// Checkout the PR source branch
		checkoutCmd := exec.Command("git", "-C", repository, "checkout", sourceBranch)
		_ = checkoutCmd.Run()

		return nil
	}
}

// formatDiff colorizes diff output with syntax highlighting
func (m Model) formatDiff(diff, filePath string) string {
	lines := strings.Split(diff, "\n")
	var result strings.Builder

	// Detect language from file path
	language := detectLanguage(filePath)

	// Define styles for different diff elements
	addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))        // bright green
	deletedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))     // bright red
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true) // bright blue
	hunkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))        // purple/magenta

	for _, line := range lines {
		if len(line) == 0 {
			result.WriteString("\n")
			continue
		}

		switch {
		case strings.HasPrefix(line, "+"):
			if !strings.HasPrefix(line, "+++") {
				// Apply syntax highlighting to added lines
				if language != "" && len(line) > 1 {
					// Get the code part (after the + prefix)
					codePart := line[1:]
					highlighted := highlightCode(codePart, language)
					// Style the + prefix and append highlighted code
					result.WriteString(addedStyle.Render("+") + highlighted + "\n")
				} else {
					result.WriteString(addedStyle.Render(line) + "\n")
				}
			} else {
				result.WriteString(headerStyle.Render(line) + "\n")
			}
		case strings.HasPrefix(line, "-"):
			if !strings.HasPrefix(line, "---") {
				// Apply syntax highlighting to deleted lines
				if language != "" && len(line) > 1 {
					// Get the code part (after the - prefix)
					codePart := line[1:]
					highlighted := highlightCode(codePart, language)
					// Style the - prefix and append highlighted code
					result.WriteString(deletedStyle.Render("-") + highlighted + "\n")
				} else {
					result.WriteString(deletedStyle.Render(line) + "\n")
				}
			} else {
				result.WriteString(headerStyle.Render(line) + "\n")
			}
		case strings.HasPrefix(line, "@@"):
			result.WriteString(hunkStyle.Render(line) + "\n")
		case strings.HasPrefix(line, "diff --git"):
			result.WriteString(headerStyle.Render(line) + "\n")
		case strings.HasPrefix(line, "new file") || strings.HasPrefix(line, "deleted file"):
			result.WriteString(hunkStyle.Render(line) + "\n")
		default:
			// Apply syntax highlighting to context lines (unchanged code)
			if language != "" && len(line) > 0 && line[0] == ' ' {
				// Get the code part (after the space prefix)
				codePart := line[1:]
				highlighted := highlightCode(codePart, language)
				result.WriteString(" " + highlighted + "\n")
			} else {
				result.WriteString(line + "\n")
			}
		}
	}

	return result.String()
}
