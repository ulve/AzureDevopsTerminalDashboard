package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ulve/azuredevops-terminal-dashboard/internal/azuredevops"
)

// loadData loads pull requests and builds from Azure DevOps
func (m Model) loadData() tea.Cmd {
	return func() tea.Msg {
		var allPRs []azuredevops.PullRequest
		var allBuilds []azuredevops.Build
		var lastErr error

		// Load pull requests
		for _, prConfig := range m.config.PullRequests {
			prs, err := m.client.GetPullRequests(prConfig.Project, prConfig.Repository)
			if err != nil {
				lastErr = fmt.Errorf("failed to load PRs for %s/%s: %w", prConfig.Project, prConfig.Repository, err)
				continue
			}
			allPRs = append(allPRs, prs...)
		}

		// Load builds
		for _, pipelineConfig := range m.config.Pipelines {
			builds, err := m.client.GetBuilds(pipelineConfig.Project, pipelineConfig.Pipeline, pipelineConfig.DefinitionID)
			if err != nil {
				pipelineIdentifier := pipelineConfig.Pipeline
				if pipelineConfig.DefinitionID > 0 {
					pipelineIdentifier = fmt.Sprintf("ID:%d", pipelineConfig.DefinitionID)
				}
				lastErr = fmt.Errorf("failed to load builds for %s/%s: %w", pipelineConfig.Project, pipelineIdentifier, err)
				continue
			}
			allBuilds = append(allBuilds, builds...)
		}

		return DataLoadedMsg{
			pullRequests: allPRs,
			builds:       allBuilds,
			err:          lastErr,
		}
	}
}

// loadPRFiles loads the files changed in a pull request
func (m Model) loadPRFiles(pr *azuredevops.PullRequest) tea.Cmd {
	return func() tea.Msg {
		files, err := m.client.GetPRFiles(pr.Repository.Project.Name, pr.Repository.Name, pr.ID)
		if err != nil {
			return FilesLoadedMsg{err: fmt.Errorf("failed to load PR files: %w", err)}
		}

		return FilesLoadedMsg{files: files}
	}
}

// loadFileDiff loads the diff for a file in a pull request
func (m Model) loadFileDiff(pr *azuredevops.PullRequest, filePath string) tea.Cmd {
	return func() tea.Msg {
		diff, err := m.client.GetPRFileDiff(pr.Repository.Project.Name, pr.Repository.Name, pr.ID, filePath)
		if err != nil {
			return DiffLoadedMsg{err: fmt.Errorf("failed to load file diff: %w", err)}
		}

		return DiffLoadedMsg{diff: diff, filePath: filePath}
	}
}

// loadBuildLogs loads the logs for a build
func (m Model) loadBuildLogs(build *azuredevops.Build) tea.Cmd {
	return func() tea.Msg {
		// Get the project name from the build - we'll need to find it from config
		// For now, we'll try all projects in the config
		var logs string
		var lastErr error

		for _, pipelineConfig := range m.config.Pipelines {
			buildLogs, err := m.client.GetBuildLogs(pipelineConfig.Project, build.ID)
			if err != nil {
				lastErr = err
				continue
			}

			// Concatenate all log files
			for _, log := range buildLogs {
				content, err := m.client.GetBuildLogContent(pipelineConfig.Project, build.ID, log.ID)
				if err != nil {
					continue
				}
				logs += fmt.Sprintf("=== Log %d ===\n%s\n\n", log.ID, content)
			}

			// If we got logs, return them
			if logs != "" {
				return LogsLoadedMsg{logs: logs}
			}
		}

		if lastErr != nil {
			return LogsLoadedMsg{err: fmt.Errorf("failed to load build logs: %w", lastErr)}
		}

		return LogsLoadedMsg{logs: "No logs available for this build"}
	}
}
