// Package tools provides a set of GitLab API tools for interacting with GitLab resources
// such as epics, issues, and merge requests. This package wraps the GitLab client-go library
// and exposes its functionality through a structured MCP (Machine Conversation Protocol) interface.
//
// The tools package enables applications to query and manipulate GitLab resources
// through a unified API. It handles pagination automatically and provides
// convenient access to GitLab's epics, issues, and merge requests endpoints.
// All results are returned as JSON for easy integration with other systems.
package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const maxPerPage = 100

// ErrArgumentType is an error indicating that the argument provided is of an invalid type.
var ErrArgumentType = errors.New("invalid argument type")

// Tools is the main entry point for GitLab API tools, providing access to
// GitLab epics, issues, and merge requests functionality.
type Tools struct {
	// Discussions provides access to GitLab discussion-related operations.
	Discussions DiscussionServiceInterface

	// Epics provides access to GitLab epic-related operations.
	Epics EpicServiceInterface

	// Events provides tools to look up events
	Events EventsServiceInterface

	// Issues provides access to GitLab issue-related operations.
	Issues IssuesServiceInterface

	// Jobs provides access to pipeline job-related operations.
	Jobs JobsServiceInterface

	// MergeRequests provides access to GitLab merge request-related operations.
	MergeRequests MergeRequestsServiceInterface

	// Repositories provides access to GitLab repository-related operations.
	Repositories RepositoryServiceInterface

	// Snippets provides access to GitLab snippet-related operations.
	Snippets SnippetsServiceInterface

	// Todos provides access to GitLab todo-related operations.
	Todos TodosServiceInterface

	// Users provides tools for looking up user IDs, a user's activity and status, etc.
	Users UsersServiceInterface
}

// New creates a new instance of Tools with the provided GitLab client and current user.
// It initializes all the service interfaces for epics, issues, and merge requests.
func New(client *gitlab.Client, currentUser string) *Tools {
	return &Tools{
		Discussions:   NewDiscussionTools(client),
		Epics:         NewEpicTools(client),
		Events:        NewEventsTools(client),
		Issues:        NewIssuesTools(client, currentUser),
		Jobs:          NewJobsTools(client),
		MergeRequests: NewMergeRequestsTools(client, currentUser),
		Repositories:  NewRepositoryTools(client),
		Snippets:      NewSnippetsTools(client, currentUser),
		Todos:         NewTodosTools(client),
		Users:         NewUsersTools(client, currentUser),
	}
}

// AddTo registers all GitLab tools with the provided MCPServer.
// It calls AddTo on all service interfaces to register their respective tools.
func (s *Tools) AddTo(srv *server.MCPServer) {
	s.Discussions.AddTo(srv)
	s.Epics.AddTo(srv)
	s.Events.AddTo(srv)
	s.Issues.AddTo(srv)
	s.Jobs.AddTo(srv)
	s.MergeRequests.AddTo(srv)
	s.Repositories.AddTo(srv)
	s.Snippets.AddTo(srv)
	s.Todos.AddTo(srv)
	s.Users.AddTo(srv)
}

// newToolResultJSON encodes the provided value as JSON and returns it as a tool result.
// It handles the JSON encoding and error handling, providing a consistent way to return
// JSON responses from tool handlers.
func newToolResultJSON(v any) (*mcp.CallToolResult, error) {
	// If v is a slice and v is nil, we return "[]" (indicating an empty slice) rather than "null".
	if value := reflect.ValueOf(v); value.Kind() == reflect.Slice && value.IsNil() {
		return mcp.NewToolResultText("[]"), nil
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(v); err != nil {
		return nil, fmt.Errorf("encoding to JSON: %w", err)
	}

	return mcp.NewToolResultText(b.String()), nil
}

// newListOptions initializes a ListOptions struct populated with default values
// and values from common arguments in the provided MCP CallToolRequest.
func newListOptions(req mcp.CallToolRequest) gitlab.ListOptions {
	opts := gitlab.ListOptions{
		PerPage: maxPerPage,
	}

	opts.OrderBy = req.GetString("order_by", "")
	opts.Sort = req.GetString("sort_order", "")

	return opts
}
