package ui

import (
	"fmt"
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
	ViewPRFiles
	ViewFileDiff
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
	selectedPR      *azuredevops.PullRequest
	prFiles         []string
	currentDiff     string
	loading         bool
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
	diff string
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

	return Model{
		config:          cfg,
		client:          client,
		view:            ViewDashboard,
		prList:          prList,
		buildList:       buildList,
		fileList:        fileList,
		diffViewport:    diffViewport,
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

		case "esc":
			// Go back to previous view
			switch m.view {
			case ViewPRFiles:
				m.view = ViewDashboard
			case ViewFileDiff:
				m.view = ViewPRFiles
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
		} else {
			m.prFiles = msg.files
			m.updateFileList()
			m.view = ViewPRFiles
		}

	case DiffLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.currentDiff = msg.diff
			m.diffViewport.SetContent(msg.diff)
			m.view = ViewFileDiff
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
	case ViewPRFiles:
		m.fileList, cmd = m.fileList.Update(msg)
	case ViewFileDiff:
		m.diffViewport, cmd = m.diffViewport.Update(msg)
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
	case ViewPRFiles:
		return m.renderPRFiles()
	case ViewFileDiff:
		return m.renderFileDiff()
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
	status := fmt.Sprintf("Last update: %s | Auto-refresh: %v | Press 'r' to refresh, 'tab' to switch, 'q' to quit",
		m.lastUpdate.Format("15:04:05"), m.autoRefresh)
	s.WriteString("\n")
	s.WriteString(statusStyle.Render(status))

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
	s.WriteString(statusStyle.Render("Press 'enter' to view diff, 'esc' to go back, 'q' to quit"))

	return s.String()
}

// renderFileDiff renders the file diff view
func (m Model) renderFileDiff() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("File Diff"))
	s.WriteString("\n\n")
	s.WriteString(m.diffViewport.View())
	s.WriteString("\n")
	s.WriteString(statusStyle.Render("Press 'esc' to go back, 'q' to quit"))

	return s.String()
}

// handleEnter handles the enter key press
func (m Model) handleEnter() (Model, tea.Cmd) {
	switch m.view {
	case ViewDashboard:
		if m.activeTab == 0 && len(m.pullRequests) > 0 {
			// Load PR files
			idx := m.prList.Index()
			if idx >= 0 && idx < len(m.pullRequests) {
				m.selectedPR = &m.pullRequests[idx]
				return m, m.loadPRFiles(m.selectedPR)
			}
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
}

// tickCmd returns a command that sends a tick message
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
