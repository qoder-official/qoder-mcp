# GitLab MCP

A Model Context Protocol (MCP) server for GitLab, enabling Claude to interact directly with your GitLab instance.

## :speech_balloon: Feedback

Please provide feedback in https://gitlab.com/gitlab-org/ux-research/-/issues/3495, or (if you don't have access) open a new issue at https://gitlab.com/fforster/gitlab-mcp/-/issues.

## Overview

GitLab MCP implements the [Model Context Protocol](https://github.com/anthropics/model-context-protocol) for GitLab, allowing AI assistants like Claude to access and manipulate GitLab resources. This integration enables Claude to work with:

- Discussions and Notes
- Epics
- Issues
- Jobs
- Merge Requests
- Repository Files and Directories
- Snippets
- User information

## Installation

### Option 1: Install from Homebrew

This is the simplest way to install *gitlab-mcp* when you're on Mac.

```sh
# Add the gitlab-mcp tap repository
brew tap fforster/gitlab-mcp https://gitlab.com/fforster/homebrew-gitlab-mcp.git

# Install gitlab-mcp
brew install gitlab-mcp
```

### Option 2: Build from source

This is recommended if you plan on contributing features or bug fixes to *gitlab-mcp*.

1. Clone this repository
2. Build the binary with `go build`, this creates `gitlab-mcp` in the current directory.
   Use `-o <path>` to chose a different filename (or directory).

### Option 3: Docker

A pre-built Docker image is available:

```bash
docker run -i -e GITLAB_TOKEN=your_gitlab_token registry.gitlab.com/fforster/gitlab-mcp:latest
```

## Personal Access Token

_gitlab-mcp_ requires a personal access token for authentication.
Generate one from your [GitLab User Settings](https://gitlab.com/-/user_settings/personal_access_tokens) with the `api` scope, or `read_api` for read-only access.

## Usage with Claude Desktop

To use the GitLab MCP server with Claude Desktop, add the following to your Claude Desktop configuration file:

### Using the binary

```json
{
  "mcpServers": {
    "GitLab": {
      "command": "/opt/homebrew/bin/gitlab-mcp",
      "env": {
        "GITLAB_TOKEN": "your_gitlab_token"
      }
    }
  }
}
```

### Using Docker

```json
{
  "mcpServers": {
    "GitLab": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "--pull=always",
        "-e", "GITLAB_TOKEN",
        "registry.gitlab.com/fforster/gitlab-mcp:latest"
      ],
      "env": {
        "GITLAB_TOKEN": "your_gitlab_token"
      }
    }
  }
}
```

**Explanation:**

* `-i` is required so Docker keeps `STDIN` open.
* `--rm` cleans up the container when *gitlab-mcp* exits.
* `--pull=always` ensures that Docker downloads the latest version of *gitlab-mcp* before running.
* `-e GITLAB_TOKEN` exports the `GITLAB_TOKEN` environment variable to *gitlab-mcp*.

### Config file location

The configuration file is typically located at:
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

### Environment Variables

You can provide the GitLab token in one of two ways:

1. **Directly in the configuration**, by adding an `env` section. This is what the above examples use.
2. **Via shell initialization files**, where your `.bashrc` or similar file sets the `GITLAB_TOKEN` environment variable:

   ```json
   {
     "mcpServers": {
       "GitLab": {
         "command": "/bin/bash",
         "args": [
           "-c",
           "/opt/homebrew/bin/gitlab-mcp"
         ]
       }
     }
   }
   ```

## Supported Tools

| Tool Name | Description |
|-----------|-------------|
| `complete_all_todo_items` | Marks all pending todo items for the current user as done |
| `complete_todo_item` | Marks a single pending todo item as done |
| `create_issue` | Creates a new GitLab issue |
| `create_snippet` | Create a new snippet |
| `delete_snippet` | Delete a snippet |
| `discussion_add_note` | Adds a new note (i.e. a reply) to an existing discussion thread |
| `discussion_delete_note` | Deletes a note from a discussion thread |
| `discussion_list` | Lists all discussions for a GitLab resource |
| `discussion_modify_note` | Modifies an existing note in a discussion thread |
| `discussion_new` | Creates a new discussion thread on a GitLab resource |
| `discussion_resolve` | Resolves or unresolves a discussion thread in a merge request |
| `download_job_artifacts_file` | Download a single artifact file from a job |
| `download_job_log` | Download a log file for a specific job |
| `edit_issue` | Updates an existing GitLab issue. You can modify the issue's title, description, and metadata, and close/reopen issues. |
| `get_epic_links` | GetEpicLinks gets all child epics of an epic |
| `get_epic` | Fetches information about an epic by ID |
| `get_issue` | Get a single project issue |
| `get_issues_closed_on_merge` | Get all the issues that would be closed by merging the provided merge request |
| `get_job` | Get a single job of a project |
| `get_merge_request_approvals` | Get approvals for a merge request |
| `get_merge_request_changes` | Get information about the merge request including its files and changes |
| `get_merge_request_commits` | Get all commits associated with a merge request |
| `get_merge_request_dependencies` | Get merge request dependencies |
| `get_merge_request_participants` | Get a list of merge request participants |
| `get_merge_request_reviewers` | Get a list of merge request reviewers |
| `get_merge_request` | Get a single merge request |
| `get_repository_file_contents` | Get the contents of a single file from the repository |
| `get_snippet_content` | Get the raw content of a snippet |
| `get_snippet` | Returns the metadata of a snippet, such as title and description. File content is not returned |
| `get_user_status` | Get a user's status |
| `get_user` | Get information about a specific user or the current user. In particular, this tool can be used to resolve a username to an ID. |
| `list_all_snippets` | List all snippets the user has access to |
| `list_draft_notes` | Returns a list of draft notes for the merge request |
| `list_downstream_pipelines` | Get a list of downstream pipeline triggers for a pipeline |
| `list_epic_issues` | Returns a list of issues assigned to the provided epic |
| `list_group_epics` | Get all epics for a specific group |
| `list_group_issues` | Get a list of a group's issues |
| `list_group_merge_requests` | Get all merge requests for this group |
| `list_merge_request_diffs` | Get a list of merge request diff versions |
| `list_merge_request_pipelines` | Get a list of merge request pipelines |
| `list_merge_requests_related_to_issue` | Get all merge requests that are related to the specified issue |
| `list_pipeline_jobs` | Get a list of jobs for a pipeline |
| `list_project_issues` | Get a list of a project's issues |
| `list_project_merge_requests` | Get all merge requests for this project |
| `list_repository_directory` | Get a list of repository files and directories in a project |
| `list_user_events`  | Get events for users between |
| `list_user_issues` | Lists all issues assigned to a user |
| `list_user_merge_requests` | Get all merge requests authored by or assigned to a user for review |
| `list_user_snippets` | List snippets owned by the current user |
| `list_user_todos` | Get all todos for the current user, with optional filtering by state (pending, done) and a limit on the number returned (default 100). |
| `retry_job` | Retry a single job of a project |
| `set_user_status` | Set the current user's status |
| `trigger_manual_job` | Trigger a manual job for a project |
| `update_snippet` | Update an existing snippet |

## Example Usage

Once configured, you can ask Claude to interact with your GitLab instance:

- "Prioritize the open issues assigned to me."
- "List the merge requests I need to review."
- "Add a comment to issue #42 suggesting a possible solution."
- "Summarize the discussion of issue #123 in the `my-project` project."
- "Create a new snippet with the code I just shared."
- "What epics are currently open in the `my-group` group?"
- "Please explain why the pipeline of merge request !456 failed."
- "How do I use the MCP server according to the `README.md` file in the `fforster/gitlab-mcp` GitLab project."
- "On GitLab, what have reprazent and ffoster been working on between 2025-04-28
  and 2025-05-02?"
- "List all of my pending todos."
- "List the latest 100 completed todos."

## License

This project is licensed under the MIT License - see the LICENSE file for details.
