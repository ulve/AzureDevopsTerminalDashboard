package api

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/ulve/azuredevops-terminal-dashboard/config"
)

// Client wraps the Azure DevOps API client
type Client struct {
	connection  *azuredevops.Connection
	buildClient build.Client
	config      *config.Config
}

// NewClient creates a new Azure DevOps API client
func NewClient(cfg *config.Config) (*Client, error) {
	organizationUrl := fmt.Sprintf("https://dev.azure.com/%s", cfg.Organization)

	connection := azuredevops.NewPatConnection(organizationUrl, cfg.PAT)

	buildClient, err := build.NewClient(context.Background(), connection)
	if err != nil {
		return nil, fmt.Errorf("failed to create build client: %w", err)
	}

	return &Client{
		connection:  connection,
		buildClient: buildClient,
		config:      cfg,
	}, nil
}

// GetBuilds retrieves recent builds/pipelines
func (c *Client) GetBuilds(ctx context.Context) ([]build.Build, error) {
	top := 50
	args := build.GetBuildsArgs{
		Project: &c.config.Project,
		Top:     &top,
	}

	buildsResp, err := c.buildClient.GetBuilds(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to get builds: %w", err)
	}

	if buildsResp == nil {
		return []build.Build{}, nil
	}

	return buildsResp.Value, nil
}

// GetBuild retrieves a specific build by ID
func (c *Client) GetBuild(ctx context.Context, buildID int) (*build.Build, error) {
	args := build.GetBuildArgs{
		Project: &c.config.Project,
		BuildId: &buildID,
	}

	buildResult, err := c.buildClient.GetBuild(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to get build: %w", err)
	}

	return buildResult, nil
}

// GetBuildTimeline retrieves the timeline (stages/jobs) for a build
func (c *Client) GetBuildTimeline(ctx context.Context, buildID int) (*build.Timeline, error) {
	args := build.GetBuildTimelineArgs{
		Project: &c.config.Project,
		BuildId: &buildID,
	}

	timeline, err := c.buildClient.GetBuildTimeline(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to get build timeline: %w", err)
	}

	return timeline, nil
}

// GetBuildLogs retrieves logs for a build
func (c *Client) GetBuildLogs(ctx context.Context, buildID int) ([]build.BuildLog, error) {
	args := build.GetBuildLogsArgs{
		Project: &c.config.Project,
		BuildId: &buildID,
	}

	logs, err := c.buildClient.GetBuildLogs(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to get build logs: %w", err)
	}

	if logs == nil {
		return []build.BuildLog{}, nil
	}

	return *logs, nil
}

// GetBuildLogContent retrieves the content of a specific log
func (c *Client) GetBuildLogContent(ctx context.Context, buildID int, logID int) (string, error) {
	args := build.GetBuildLogArgs{
		Project: &c.config.Project,
		BuildId: &buildID,
		LogId:   &logID,
	}

	logReader, err := c.buildClient.GetBuildLog(ctx, args)
	if err != nil {
		return "", fmt.Errorf("failed to get build log content: %w", err)
	}
	defer logReader.Close()

	// Read the log content
	var buf strings.Builder
	_, err = io.Copy(&buf, logReader)
	if err != nil {
		return "", fmt.Errorf("failed to read log content: %w", err)
	}

	return buf.String(), nil
}
