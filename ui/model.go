package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ulve/azuredevops-terminal-dashboard/api"
	"github.com/ulve/azuredevops-terminal-dashboard/models"
)

type view int

const (
	pipelineListView view = iota
	pipelineDetailView
)

// Model represents the main application model
type Model struct {
	client          *api.Client
	pipelines       []*models.Pipeline
	selectedIndex   int
	currentView     view
	selectedPipeline *models.Pipeline
	stages          []models.StageInfo
	logs            string
	viewport        viewport.Model
	width           int
	height          int
	loading         bool
	err             error
	autoRefresh     bool
}

// NewModel creates a new application model
func NewModel(client *api.Client) Model {
	return Model{
		client:      client,
		pipelines:   make([]*models.Pipeline, 0),
		currentView: pipelineListView,
		loading:     true,
		autoRefresh: true,
		viewport:    viewport.New(80, 20),
	}
}

type pipelinesLoadedMsg struct {
	pipelines []*models.Pipeline
	err       error
}

type pipelineDetailLoadedMsg struct {
	pipeline *models.Pipeline
	stages   []models.StageInfo
	logs     string
	err      error
}

type tickMsg time.Time

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadPipelines,
		tickCmd(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			if m.currentView == pipelineDetailView {
				m.currentView = pipelineListView
				m.selectedPipeline = nil
				m.stages = nil
				m.logs = ""
			}

		case "r":
			// Manual refresh
			if m.currentView == pipelineListView {
				m.loading = true
				return m, m.loadPipelines
			} else if m.currentView == pipelineDetailView && m.selectedPipeline != nil {
				m.loading = true
				return m, m.loadPipelineDetail(m.selectedPipeline.ID)
			}

		case "up", "k":
			if m.currentView == pipelineListView && m.selectedIndex > 0 {
				m.selectedIndex--
			} else if m.currentView == pipelineDetailView {
				m.viewport.LineUp(1)
			}

		case "down", "j":
			if m.currentView == pipelineListView && m.selectedIndex < len(m.pipelines)-1 {
				m.selectedIndex++
			} else if m.currentView == pipelineDetailView {
				m.viewport.LineDown(1)
			}

		case "enter":
			if m.currentView == pipelineListView && len(m.pipelines) > 0 {
				m.selectedPipeline = m.pipelines[m.selectedIndex]
				m.currentView = pipelineDetailView
				m.loading = true
				return m, m.loadPipelineDetail(m.selectedPipeline.ID)
			}

		case "pgup":
			if m.currentView == pipelineDetailView {
				m.viewport.ViewUp()
			}

		case "pgdown":
			if m.currentView == pipelineDetailView {
				m.viewport.ViewDown()
			}
		}

	case pipelinesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.pipelines = msg.pipelines
			if m.selectedIndex >= len(m.pipelines) {
				m.selectedIndex = 0
			}
		}

	case pipelineDetailLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.selectedPipeline = msg.pipeline
			m.stages = msg.stages
			m.logs = msg.logs
			m.viewport.SetContent(m.renderDetailContent())
		}

	case tickMsg:
		// Auto-refresh every 10 seconds
		if m.autoRefresh {
			if m.currentView == pipelineListView {
				cmds = append(cmds, m.loadPipelines)
			} else if m.currentView == pipelineDetailView && m.selectedPipeline != nil {
				cmds = append(cmds, m.loadPipelineDetail(m.selectedPipeline.ID))
			}
		}
		cmds = append(cmds, tickCmd())
	}

	return m, tea.Batch(cmds...)
}

// View renders the application
func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress q to quit", m.err))
	}

	var content string

	switch m.currentView {
	case pipelineListView:
		content = m.renderPipelineList()
	case pipelineDetailView:
		content = m.renderPipelineDetail()
	}

	help := m.renderHelp()

	return lipgloss.JoinVertical(lipgloss.Left, content, help)
}

func (m Model) renderPipelineList() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Azure DevOps - Pipeline Dashboard")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.loading && len(m.pipelines) == 0 {
		b.WriteString("Loading pipelines...\n")
		return b.String()
	}

	if len(m.pipelines) == 0 {
		b.WriteString("No pipelines found.\n")
		return b.String()
	}

	// Pipeline list
	for i, pipeline := range m.pipelines {
		var line string

		status := GetStatusStyle(string(pipeline.Status)).Render(fmt.Sprintf("%-12s", pipeline.Status))
		definition := pipeline.Definition
		if len(definition) > 30 {
			definition = definition[:27] + "..."
		}

		branch := pipeline.SourceBranch
		if len(branch) > 25 {
			parts := strings.Split(branch, "/")
			branch = parts[len(parts)-1]
			if len(branch) > 25 {
				branch = branch[:22] + "..."
			}
		}

		duration := pipeline.Duration()

		line = fmt.Sprintf("%s  %-32s  %-27s  %s",
			status,
			definition,
			branch,
			duration,
		)

		if i == m.selectedIndex {
			line = selectedListItemStyle.Render("▶ " + line)
		} else {
			line = listItemStyle.Render("  " + line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderPipelineDetail() string {
	if m.selectedPipeline == nil {
		return "No pipeline selected"
	}

	var b strings.Builder

	// Title with back navigation hint
	title := titleStyle.Render(fmt.Sprintf("Pipeline: %s - %s", m.selectedPipeline.Definition, m.selectedPipeline.Number))
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Press ESC to go back"))
	b.WriteString("\n\n")

	if m.loading && len(m.stages) == 0 {
		b.WriteString("Loading pipeline details...\n")
		return b.String()
	}

	// Use viewport for scrollable content
	b.WriteString(m.viewport.View())

	return b.String()
}

func (m Model) renderDetailContent() string {
	var b strings.Builder

	// Pipeline info
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Status: "))
	b.WriteString(GetStatusStyle(string(m.selectedPipeline.Status)).Render(string(m.selectedPipeline.Status)))
	b.WriteString("\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Branch: "))
	b.WriteString(m.selectedPipeline.SourceBranch)
	b.WriteString("\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Requested By: "))
	b.WriteString(m.selectedPipeline.RequestedBy)
	b.WriteString("\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Duration: "))
	b.WriteString(m.selectedPipeline.Duration())
	b.WriteString("\n\n")

	// Stages and jobs
	if len(m.stages) > 0 {
		b.WriteString(titleStyle.Render("Pipeline Progress"))
		b.WriteString("\n\n")

		for _, stage := range m.stages {
			stageStatus := GetStatusStyle(stage.Result)
			if stage.Result == "None" || stage.Result == "" {
				stageStatus = GetStatusStyle(stage.State)
			}

			b.WriteString(stageStatus.Render(fmt.Sprintf("▼ Stage: %s", stage.Name)))
			b.WriteString(fmt.Sprintf(" [%s]", stage.State))
			if stage.Result != "None" && stage.Result != "" {
				b.WriteString(fmt.Sprintf(" - %s", stage.Result))
			}
			b.WriteString("\n")

			for _, job := range stage.Jobs {
				jobStatus := GetStatusStyle(job.Result)
				if job.Result == "None" || job.Result == "" {
					jobStatus = GetStatusStyle(job.State)
				}

				b.WriteString("  ")
				b.WriteString(jobStatus.Render(fmt.Sprintf("  • %s", job.Name)))
				b.WriteString(fmt.Sprintf(" [%s]", job.State))
				if job.Result != "None" && job.Result != "" {
					b.WriteString(fmt.Sprintf(" - %s", job.Result))
				}
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	// Logs
	if m.logs != "" {
		b.WriteString("\n")
		b.WriteString(titleStyle.Render("Recent Logs"))
		b.WriteString("\n\n")

		// Show last 50 lines of logs
		lines := strings.Split(m.logs, "\n")
		startLine := 0
		if len(lines) > 50 {
			startLine = len(lines) - 50
		}

		for i := startLine; i < len(lines); i++ {
			line := lines[i]
			if len(line) > m.viewport.Width-4 {
				line = line[:m.viewport.Width-7] + "..."
			}
			b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(line))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) renderHelp() string {
	var helps []string

	if m.currentView == pipelineListView {
		helps = []string{
			"↑/k up",
			"↓/j down",
			"enter select",
			"r refresh",
			"q quit",
		}
	} else {
		helps = []string{
			"↑/↓ scroll",
			"pgup/pgdown page",
			"esc back",
			"r refresh",
			"q quit",
		}
	}

	return helpStyle.Render(strings.Join(helps, " • "))
}

func (m Model) loadPipelines() tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	builds, err := m.client.GetBuilds(ctx)
	if err != nil {
		return pipelinesLoadedMsg{err: err}
	}

	pipelines := make([]*models.Pipeline, 0, len(builds))
	for _, build := range builds {
		b := build // Create a copy to avoid pointer issues
		pipelines = append(pipelines, models.FromBuild(&b))
	}

	return pipelinesLoadedMsg{pipelines: pipelines}
}

func (m Model) loadPipelineDetail(buildID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get build details
		build, err := m.client.GetBuild(ctx, buildID)
		if err != nil {
			return pipelineDetailLoadedMsg{err: err}
		}

		pipeline := models.FromBuild(build)

		// Get timeline (stages/jobs)
		timeline, err := m.client.GetBuildTimeline(ctx, buildID)
		if err != nil {
			// Timeline might not be available yet, don't fail
			timeline = nil
		}

		stages := models.ParseTimeline(timeline)

		// Get logs
		logs := ""
		buildLogs, err := m.client.GetBuildLogs(ctx, buildID)
		if err == nil && len(buildLogs) > 0 {
			// Get the most recent log
			lastLog := buildLogs[len(buildLogs)-1]
			if lastLog.Id != nil {
				logContent, err := m.client.GetBuildLogContent(ctx, buildID, *lastLog.Id)
				if err == nil {
					logs = logContent
				}
			}
		}

		return pipelineDetailLoadedMsg{
			pipeline: pipeline,
			stages:   stages,
			logs:     logs,
		}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
