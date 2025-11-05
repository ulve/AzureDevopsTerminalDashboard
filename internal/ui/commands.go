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

		return DiffLoadedMsg{diff: diff}
	}
}
