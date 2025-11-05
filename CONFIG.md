# Configuration Guide

This guide provides detailed information about configuring Azure DevOps Terminal Dashboard (ADTD).

## Quick Start

1. Copy the example config file:
   ```bash
   cp .adtd.json.example .adtd.json
   ```

2. Edit `.adtd.json` with your organization and project details

3. Set your Azure DevOps Personal Access Token:
   ```bash
   export AZURE_DEVOPS_PAT="your-personal-access-token"
   ```

4. Run the dashboard:
   ```bash
   ./adtd
   ```

## Configuration File Locations

ADTD searches for configuration files in the following order:

### 1. Current Working Directory (Default)
The most common location. Place `.adtd.json` in the directory where you run the command:

```bash
/path/to/your/project/.adtd.json
```

**Usage:**
```bash
cd /path/to/your/project
./adtd
```

This is ideal for project-specific monitoring configurations.

### 2. Custom Path (Command-Line Argument)
Specify any config file location as an argument:

```bash
./adtd /path/to/custom/config.json
./adtd ~/.config/adtd/work-config.json
./adtd ~/projects/myproject/azdo-config.json
```

This is useful for:
- Multiple configuration profiles (work, personal, different teams)
- Shared configs in a central location
- Project-specific configs with different names

### 3. Recommended Locations

#### Per-Project Configuration
For project-specific monitoring:
```
/path/to/project/.adtd.json
```

Add `.adtd.json` to `.gitignore` to keep credentials and project structure private.

#### User-Wide Configuration
For personal default configuration:
```
~/.config/adtd/config.json
~/.adtd.json
```

Then run from anywhere:
```bash
adtd ~/.config/adtd/config.json
```

#### Team Shared Configuration (Template)
Store a template (without sensitive data) in your repository:
```
/path/to/project/.adtd.json.template
```

Team members copy and customize:
```bash
cp .adtd.json.template .adtd.json
# Edit .adtd.json with personal organization/project details
```

## Configuration File Format

ADTD uses JSON configuration files with the following structure:

### Basic Configuration

```json
{
  "organization": "your-organization-name",
  "pullRequests": [
    {
      "project": "ProjectName",
      "repository": "RepositoryName"
    }
  ],
  "pipelines": [
    {
      "project": "ProjectName",
      "pipeline": "PipelineName"
    }
  ],
  "refreshInterval": 30
}
```

### Configuration Fields

#### `organization` (required, string)
Your Azure DevOps organization name.

**Finding it:** From your Azure DevOps URL: `https://dev.azure.com/{organization}`

**Example:**
```json
{
  "organization": "contoso"
}
```

#### `pullRequests` (optional, array)
Array of repositories to monitor for pull requests. At least one of `pullRequests` or `pipelines` must be configured.

**Fields:**
- `project` (required, string): The Azure DevOps project name
- `repository` (required, string): The repository name within the project

**Finding values:**
- Project: From your project URL: `https://dev.azure.com/{org}/{project}`
- Repository: Navigate to Repos → Files, repository name is in the breadcrumb/dropdown

**Example:**
```json
{
  "pullRequests": [
    {
      "project": "MyTeam",
      "repository": "web-frontend"
    },
    {
      "project": "MyTeam",
      "repository": "api-backend"
    },
    {
      "project": "Infrastructure",
      "repository": "terraform-config"
    }
  ]
}
```

#### `pipelines` (optional, array)
Array of pipelines to monitor for builds. At least one of `pullRequests` or `pipelines` must be configured.

**Fields:**
- `project` (required, string): The Azure DevOps project name
- `pipeline` (required, string): The pipeline name (not ID) as displayed in Azure DevOps

**Finding values:**
- Project: From your project URL: `https://dev.azure.com/{org}/{project}`
- Pipeline: Navigate to Pipelines, use the exact name shown in the list (not the ID)

**Example:**
```json
{
  "pipelines": [
    {
      "project": "MyTeam",
      "pipeline": "web-frontend-ci"
    },
    {
      "project": "MyTeam",
      "pipeline": "api-backend-build"
    },
    {
      "project": "Infrastructure",
      "pipeline": "deploy-production"
    }
  ]
}
```

#### `refreshInterval` (optional, integer)
Auto-refresh interval in seconds. Default: 30 seconds.

**Example:**
```json
{
  "refreshInterval": 60
}
```

Set to higher values (60-120) for large teams or to reduce API calls.

## Complete Configuration Examples

### Example 1: Frontend Development Team

```json
{
  "organization": "mycompany",
  "pullRequests": [
    {
      "project": "WebApps",
      "repository": "react-dashboard"
    },
    {
      "project": "WebApps",
      "repository": "vue-admin-panel"
    }
  ],
  "pipelines": [
    {
      "project": "WebApps",
      "pipeline": "react-dashboard-ci"
    },
    {
      "project": "WebApps",
      "pipeline": "vue-admin-panel-ci"
    }
  ],
  "refreshInterval": 45
}
```

### Example 2: DevOps/SRE Team

```json
{
  "organization": "mycompany",
  "pipelines": [
    {
      "project": "Infrastructure",
      "pipeline": "terraform-apply-prod"
    },
    {
      "project": "Infrastructure",
      "pipeline": "kubernetes-deploy-prod"
    },
    {
      "project": "Infrastructure",
      "pipeline": "terraform-apply-staging"
    },
    {
      "project": "Monitoring",
      "pipeline": "prometheus-deploy"
    }
  ],
  "refreshInterval": 30
}
```

### Example 3: Full-Stack Team (PRs Only)

```json
{
  "organization": "startup",
  "pullRequests": [
    {
      "project": "ProductA",
      "repository": "web-app"
    },
    {
      "project": "ProductA",
      "repository": "mobile-app"
    },
    {
      "project": "ProductA",
      "repository": "api-gateway"
    },
    {
      "project": "ProductA",
      "repository": "auth-service"
    }
  ],
  "refreshInterval": 30
}
```

### Example 4: Multi-Project Monitoring

```json
{
  "organization": "enterprise",
  "pullRequests": [
    {
      "project": "CustomerPortal",
      "repository": "frontend"
    },
    {
      "project": "CustomerPortal",
      "repository": "backend"
    },
    {
      "project": "InternalTools",
      "repository": "admin-dashboard"
    }
  ],
  "pipelines": [
    {
      "project": "CustomerPortal",
      "pipeline": "frontend-cd-prod"
    },
    {
      "project": "CustomerPortal",
      "pipeline": "backend-cd-prod"
    },
    {
      "project": "InternalTools",
      "pipeline": "admin-dashboard-ci"
    },
    {
      "project": "Platform",
      "pipeline": "database-migrations"
    }
  ],
  "refreshInterval": 60
}
```

## Environment Variables

### AZURE_DEVOPS_PAT (required)

Your Azure DevOps Personal Access Token for authentication.

**Setting it:**
```bash
# Temporarily (current shell session only)
export AZURE_DEVOPS_PAT="your-personal-access-token"

# Permanently (add to ~/.bashrc, ~/.zshrc, or ~/.profile)
echo 'export AZURE_DEVOPS_PAT="your-personal-access-token"' >> ~/.bashrc
source ~/.bashrc
```

**Creating a PAT:**
1. Go to Azure DevOps → User Settings → Personal Access Tokens
2. Click "New Token"
3. Give it a descriptive name (e.g., "ADTD Terminal Dashboard")
4. Select expiration date
5. **Required Scopes:**
   - **Code (Read)** - For pull requests and file changes
   - **Build (Read)** - For pipeline builds and runs
6. Click "Create" and copy the token immediately

**Security Note:** Store your PAT securely. Consider using a password manager or secret management tool. Never commit PATs to version control.

## Validation

ADTD validates your configuration on startup. Common validation errors:

### "organization is required"
The `organization` field is missing or empty.

**Fix:**
```json
{
  "organization": "your-org-name"
}
```

### "at least one pull request or pipeline must be configured"
Both `pullRequests` and `pipelines` arrays are empty or missing.

**Fix:** Add at least one pull request or pipeline:
```json
{
  "organization": "your-org",
  "pullRequests": [
    {
      "project": "Project1",
      "repository": "Repo1"
    }
  ]
}
```

### "pull request X: project is required"
A pull request entry is missing the `project` field.

**Fix:** Add the project name:
```json
{
  "pullRequests": [
    {
      "project": "MyProject",
      "repository": "MyRepo"
    }
  ]
}
```

### "pipeline X: pipeline is required"
A pipeline entry is missing the `pipeline` field.

**Fix:** Add the pipeline name:
```json
{
  "pipelines": [
    {
      "project": "MyProject",
      "pipeline": "MyPipeline"
    }
  ]
}
```

### "AZURE_DEVOPS_PAT environment variable is not set"
The required environment variable is not set.

**Fix:**
```bash
export AZURE_DEVOPS_PAT="your-token"
```

## Tips and Best Practices

### 1. Use Multiple Config Files
Create different configs for different contexts:

```bash
# Work projects
./adtd ~/.config/adtd/work.json

# Personal projects
./adtd ~/.config/adtd/personal.json

# Specific team
./adtd ~/.config/adtd/platform-team.json
```

### 2. Project-Specific Configs
Keep project-specific configs in project directories:

```bash
cd ~/projects/web-app
./adtd .adtd.json  # Monitors just this project
```

### 3. Add to .gitignore
If your config contains project structure info you want to keep private:

```bash
echo ".adtd.json" >> .gitignore
```

### 4. Use Templates for Teams
Create a template for your team:

```json
// .adtd.json.template
{
  "organization": "mycompany",
  "pullRequests": [
    {
      "project": "OurProject",
      "repository": "main-repo"
    }
  ],
  "pipelines": [
    {
      "project": "OurProject",
      "pipeline": "ci-cd-pipeline"
    }
  ],
  "refreshInterval": 30
}
```

Team members copy and customize as needed.

### 5. Adjust Refresh Interval
- **Faster (15-30s):** Small teams, active development, critical monitoring
- **Moderate (30-60s):** Normal development, balanced API usage
- **Slower (60-120s):** Large teams, background monitoring, API rate limiting

### 6. Monitor Only What You Need
Don't add all repositories and pipelines. Focus on:
- Projects you actively contribute to
- Critical pipelines you need to monitor
- Repositories with frequent PR activity

This reduces API calls and improves performance.

## Troubleshooting

### "failed to read config file"
- Check the file path is correct
- Verify the file exists
- Ensure you have read permissions

### "failed to parse config file"
- Validate JSON syntax (use `jsonlint` or an online validator)
- Check for missing commas, brackets, or quotes
- Ensure proper JSON escaping

### "401 Unauthorized" errors
- Verify your PAT is correct
- Check PAT hasn't expired
- Ensure PAT has required scopes (Code: Read, Build: Read)

### "404 Not Found" errors
- Double-check organization, project, repository, and pipeline names
- Ensure exact name matches (case-sensitive)
- Verify you have access to the specified resources

### Performance issues
- Reduce number of monitored repositories/pipelines
- Increase `refreshInterval`
- Check network connectivity

## See Also

- [README.md](README.md) - General usage and features
- [.adtd.json.example](.adtd.json.example) - Example configuration file
