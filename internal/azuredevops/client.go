package azuredevops

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	apiVersion = "7.1"
	baseURL    = "https://dev.azure.com"
)

// Client represents an Azure DevOps API client
type Client struct {
	organization string
	pat          string
	httpClient   *http.Client
}

// NewClient creates a new Azure DevOps client
func NewClient(organization, pat string) *Client {
	return &Client{
		organization: organization,
		pat:          pat,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an authenticated HTTP request
func (c *Client) doRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic authentication with PAT
	req.SetBasicAuth("", c.pat)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// PullRequest represents a pull request
type PullRequest struct {
	ID           int       `json:"pullRequestId"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	CreatedBy    User      `json:"createdBy"`
	CreationDate time.Time `json:"creationDate"`
	Repository   Repository `json:"repository"`
	SourceRefName string   `json:"sourceRefName"`
	TargetRefName string   `json:"targetRefName"`
	IsDraft      bool      `json:"isDraft"`
}

// User represents a user
type User struct {
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
}

// Repository represents a repository
type Repository struct {
	Name    string  `json:"name"`
	Project Project `json:"project"`
}

// Project represents a project
type Project struct {
	Name string `json:"name"`
}

// PullRequestsResponse represents the API response for pull requests
type PullRequestsResponse struct {
	Value []PullRequest `json:"value"`
	Count int          `json:"count"`
}

// GetPullRequests fetches active pull requests for a repository
func (c *Client) GetPullRequests(project, repository string) ([]PullRequest, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/pullrequests?searchCriteria.status=active&api-version=%s",
		baseURL, c.organization, project, repository, apiVersion)

	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var response PullRequestsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse pull requests response: %w", err)
	}

	return response.Value, nil
}

// Build represents a pipeline build/run
type Build struct {
	ID            int       `json:"id"`
	BuildNumber   string    `json:"buildNumber"`
	Status        string    `json:"status"`
	Result        string    `json:"result"`
	QueueTime     time.Time `json:"queueTime"`
	StartTime     time.Time `json:"startTime"`
	FinishTime    time.Time `json:"finishTime"`
	SourceBranch  string    `json:"sourceBranch"`
	Definition    Definition `json:"definition"`
	RequestedFor  User      `json:"requestedFor"`
}

// Definition represents a pipeline definition
type Definition struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// BuildsResponse represents the API response for builds
type BuildsResponse struct {
	Value []Build `json:"value"`
	Count int     `json:"count"`
}

// GetBuilds fetches recent builds for a pipeline
// Either pipelineName or definitionID can be provided. If definitionID is provided (> 0), it will be used directly.
func (c *Client) GetBuilds(project, pipelineName string, definitionID int) ([]Build, error) {
	var definition Definition
	var err error

	if definitionID > 0 {
		// Use the provided definition ID directly
		definition, err = c.getDefinitionByID(project, definitionID)
		if err != nil {
			return nil, err
		}
	} else {
		// Search for pipeline by name
		definition, err = c.getPipelineDefinition(project, pipelineName)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds?definitions=%d&statusFilter=all&queryOrder=queueTimeDescending&$top=10&api-version=%s",
		baseURL, c.organization, project, definition.ID, apiVersion)

	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var response BuildsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse builds response: %w", err)
	}

	// Ensure all builds have the definition name populated
	// The API response may not include the full definition details
	for i := range response.Value {
		if response.Value[i].Definition.Name == "" {
			response.Value[i].Definition.Name = definition.Name
		}
		if response.Value[i].Definition.ID == 0 {
			response.Value[i].Definition.ID = definition.ID
		}
	}

	return response.Value, nil
}

// DefinitionsResponse represents the API response for pipeline definitions
type DefinitionsResponse struct {
	Value []Definition `json:"value"`
	Count int         `json:"count"`
}

// getDefinitionByID gets the pipeline definition by its ID
func (c *Client) getDefinitionByID(project string, definitionID int) (Definition, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/build/definitions/%d?api-version=%s",
		baseURL, c.organization, project, definitionID, apiVersion)

	body, err := c.doRequest(url)
	if err != nil {
		return Definition{}, fmt.Errorf("failed to get definition %d: %w", definitionID, err)
	}

	var definition Definition
	if err := json.Unmarshal(body, &definition); err != nil {
		return Definition{}, fmt.Errorf("failed to parse definition response: %w", err)
	}

	return definition, nil
}

// getPipelineDefinition gets the pipeline definition (ID and Name) by name
func (c *Client) getPipelineDefinition(project, pipelineName string) (Definition, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/build/definitions?name=%s&api-version=%s",
		baseURL, c.organization, project, pipelineName, apiVersion)

	body, err := c.doRequest(url)
	if err != nil {
		return Definition{}, err
	}

	var response DefinitionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return Definition{}, fmt.Errorf("failed to parse definitions response: %w", err)
	}

	if len(response.Value) == 0 {
		return Definition{}, fmt.Errorf("pipeline '%s' not found in project '%s'", pipelineName, project)
	}

	return response.Value[0], nil
}

// PRIteration represents a pull request iteration
type PRIteration struct {
	ID int `json:"id"`
}

// PRIterationsResponse represents the API response for PR iterations
type PRIterationsResponse struct {
	Value []PRIteration `json:"value"`
}

// PRChange represents a file change in a pull request
type PRChange struct {
	ChangeType string     `json:"changeType"`
	Item       PRItem     `json:"item"`
}

// PRItem represents a file item in a pull request
type PRItem struct {
	Path string `json:"path"`
}

// PRChangesResponse represents the API response for PR changes
type PRChangesResponse struct {
	Changes []PRChange `json:"changes"`
}

// GetPRFiles fetches the list of changed files in a pull request
func (c *Client) GetPRFiles(project, repository string, prID int) ([]string, error) {
	// Get the latest iteration
	iterURL := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/pullRequests/%d/iterations?api-version=%s",
		baseURL, c.organization, project, repository, prID, apiVersion)

	body, err := c.doRequest(iterURL)
	if err != nil {
		return nil, err
	}

	var iterResponse PRIterationsResponse
	if err := json.Unmarshal(body, &iterResponse); err != nil {
		return nil, fmt.Errorf("failed to parse iterations response: %w", err)
	}

	if len(iterResponse.Value) == 0 {
		return []string{}, nil
	}

	latestIter := iterResponse.Value[len(iterResponse.Value)-1]

	// Get changes for the latest iteration
	changesURL := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/pullRequests/%d/iterations/%d/changes?api-version=%s",
		baseURL, c.organization, project, repository, prID, latestIter.ID, apiVersion)

	body, err = c.doRequest(changesURL)
	if err != nil {
		return nil, err
	}

	var changesResponse PRChangesResponse
	if err := json.Unmarshal(body, &changesResponse); err != nil {
		return nil, fmt.Errorf("failed to parse changes response: %w", err)
	}

	files := make([]string, 0, len(changesResponse.Changes))
	for _, change := range changesResponse.Changes {
		files = append(files, change.Item.Path)
	}

	return files, nil
}

// PRDiff represents a diff for a file
type PRDiff struct {
	Path string
	Diff string
}

// GetPRFileDiff fetches the diff for a specific file in a pull request
func (c *Client) GetPRFileDiff(project, repository string, prID int, filePath string) (string, error) {
	// Get the PR details to get source and target commits
	prURL := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/pullRequests/%d?api-version=%s",
		baseURL, c.organization, project, repository, prID, apiVersion)

	body, err := c.doRequest(prURL)
	if err != nil {
		return "", err
	}

	var pr struct {
		LastMergeSourceCommit struct {
			CommitID string `json:"commitId"`
		} `json:"lastMergeSourceCommit"`
		LastMergeTargetCommit struct {
			CommitID string `json:"commitId"`
		} `json:"lastMergeTargetCommit"`
	}

	if err := json.Unmarshal(body, &pr); err != nil {
		return "", fmt.Errorf("failed to parse PR details: %w", err)
	}

	// Get the file content from target commit (base)
	targetURL := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/items?path=%s&versionDescriptor.versionType=commit&versionDescriptor.version=%s&api-version=%s",
		baseURL, c.organization, project, repository, filePath, pr.LastMergeTargetCommit.CommitID, apiVersion)

	targetContent, err := c.doRequest(targetURL)
	targetText := ""
	isNewFile := false
	if err != nil {
		isNewFile = true
		targetText = ""
	} else {
		targetText = string(targetContent)
	}

	// Get the file content from source commit (new)
	sourceURL := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/items?path=%s&versionDescriptor.versionType=commit&versionDescriptor.version=%s&api-version=%s",
		baseURL, c.organization, project, repository, filePath, pr.LastMergeSourceCommit.CommitID, apiVersion)

	sourceContent, err := c.doRequest(sourceURL)
	sourceText := ""
	isDeletedFile := false
	if err != nil {
		isDeletedFile = true
		sourceText = ""
	} else {
		sourceText = string(sourceContent)
	}

	// Generate unified diff
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(targetText, sourceText, false)

	// Convert to unified diff format
	var result strings.Builder
	result.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))

	if isNewFile {
		result.WriteString("new file\n")
		result.WriteString(fmt.Sprintf("--- /dev/null\n"))
		result.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))
		// For new files, show all content as additions
		lines := strings.Split(sourceText, "\n")
		if len(lines) > 0 {
			result.WriteString("@@ -0,0 +1," + fmt.Sprintf("%d", len(lines)) + " @@\n")
			for _, line := range lines {
				result.WriteString("+" + line + "\n")
			}
		}
	} else if isDeletedFile {
		result.WriteString("deleted file\n")
		result.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
		result.WriteString(fmt.Sprintf("+++ /dev/null\n"))
		// For deleted files, show all content as deletions
		lines := strings.Split(targetText, "\n")
		if len(lines) > 0 {
			result.WriteString("@@ -1," + fmt.Sprintf("%d", len(lines)) + " +0,0 @@\n")
			for _, line := range lines {
				result.WriteString("-" + line + "\n")
			}
		}
	} else {
		result.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
		result.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

		// Generate line-by-line diff manually
		targetLines := strings.Split(targetText, "\n")
		sourceLines := strings.Split(sourceText, "\n")

		// Simple line-by-line comparison
		result.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(targetLines), len(sourceLines)))

		// Show the diff using the dmp library's results
		for _, diff := range diffs {
			lines := strings.Split(diff.Text, "\n")
			for i, line := range lines {
				// Skip empty last line from split
				if i == len(lines)-1 && line == "" {
					continue
				}

				switch diff.Type {
				case diffmatchpatch.DiffInsert:
					result.WriteString("+" + line + "\n")
				case diffmatchpatch.DiffDelete:
					result.WriteString("-" + line + "\n")
				case diffmatchpatch.DiffEqual:
					result.WriteString(" " + line + "\n")
				}
			}
		}
	}

	return result.String(), nil
}

// BuildLog represents a build log
type BuildLog struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// BuildLogsResponse represents the API response for build logs
type BuildLogsResponse struct {
	Value []BuildLog `json:"value"`
	Count int        `json:"count"`
}

// GetBuildLogs fetches the list of logs for a build
func (c *Client) GetBuildLogs(project string, buildID int) ([]BuildLog, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d/logs?api-version=%s",
		baseURL, c.organization, project, buildID, apiVersion)

	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var response BuildLogsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse build logs response: %w", err)
	}

	return response.Value, nil
}

// GetBuildLogContent fetches the content of a specific build log
func (c *Client) GetBuildLogContent(project string, buildID int, logID int) (string, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d/logs/%d?api-version=%s",
		baseURL, c.organization, project, buildID, logID, apiVersion)

	body, err := c.doRequest(url)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
