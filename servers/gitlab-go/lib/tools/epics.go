package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

// EpicServiceInterface defines the interface for epic-related GitLab operations.
// It provides methods for retrieving and listing epics and their associated issues.
type EpicServiceInterface interface {
	// AddTo registers all epic-related tools with the provided MCPServer.
	AddTo(srv *server.MCPServer)

	// ListGroupEpics returns a tool for listing all epics in a specific group.
	ListGroupEpics() server.ServerTool

	// GetEpic returns a tool for fetching a specific epic by its ID.
	GetEpic() server.ServerTool

	// GetEpicLinks returns a tool for fetching all child epics of a specific epic.
	GetEpicLinks() server.ServerTool

	// ListEpicIssues returns a tool for listing all issues assigned to a specific epic.
	ListEpicIssues() server.ServerTool
}

// NewEpicTools creates a new instance of EpicServiceInterface with the provided GitLab client.
// It returns an implementation that can be used to interact with GitLab's epic API.
func NewEpicTools(client *gitlab.Client) *EpicService {
	return &EpicService{client: client}
}

type EpicService struct {
	client *gitlab.Client
}

// AddTo registers all epic-related tools with the provided MCPServer.
// It adds tools for listing, retrieving, and managing epics and their associated issues.
func (e *EpicService) AddTo(srv *server.MCPServer) {
	srv.AddTools(
		e.ListGroupEpics(),
		e.GetEpic(),
		e.GetEpicLinks(),
		e.ListEpicIssues(),
	)
}

type listGroupEpicsArgs struct {
	GroupID mcpargs.ID `mcp_desc:"ID of the group either in owner/namespace format or the numeric group ID" mcp_required:"true"`
	State   string     `mcp_desc:"Return all epics or just those that are opened or closed" mcp_enum:"all,opened,closed"`
}

// ListGroupEpics returns a ServerTool for listing all epics in a specific group.
// The tool accepts a group ID and optional state filter parameters.
func (e *EpicService) ListGroupEpics() server.ServerTool {
	return server.ServerTool{
		Handler: e.listGroupEpics,
		Tool: mcpargs.NewTool("list_group_epics", listGroupEpicsArgs{},
			mcp.WithDescription("Get all epics for a specific group"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (e *EpicService) listGroupEpics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listGroupEpicsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opt = &gitlab.ListGroupEpicsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: maxPerPage,
			},
		}
		epics []*gitlab.Epic
	)

	// Parse optional parameters
	if args.State != "" {
		opt.State = gitlab.Ptr(args.State)
	}

	for {
		e, resp, err := e.client.Epics.ListGroupEpics(args.GroupID.Value(), opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListGroupEpics(%q): %w", args.GroupID.Value(), err)
		}

		epics = append(epics, e...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(epics)
}

type getEpicArgs struct {
	GroupID mcpargs.ID `mcp_desc:"ID of the group either in owner/namespace format or the numeric group ID" mcp_required:"true"`
	EpicIID int        `mcp_desc:"Internal ID of the epic to fetch" mcp_required:"true"`
}

// GetEpic returns a ServerTool for fetching information about a specific epic.
// The tool accepts a group ID and an epic internal ID as parameters.
func (e *EpicService) GetEpic() server.ServerTool {
	return server.ServerTool{
		Handler: e.getEpic,
		Tool: mcpargs.NewTool("get_epic", getEpicArgs{},
			mcp.WithDescription("Fetches information about an epic by ID"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (e *EpicService) getEpic(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getEpicArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	epic, _, err := e.client.Epics.GetEpic(args.GroupID.Value(), args.EpicIID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetEpic(%q, %d): %w", args.GroupID.Value(), args.EpicIID, err)
	}

	return newToolResultJSON(epic)
}

type getEpicLinksArgs struct {
	GroupID mcpargs.ID `mcp_desc:"ID of the group either in owner/namespace format or the numeric group ID" mcp_required:"true"`
	EpicIID int        `mcp_desc:"Internal ID of the epic to fetch child epics for" mcp_required:"true"`
}

// GetEpicLinks returns a ServerTool for retrieving all child epics of a specific epic.
// The tool accepts a group ID and an epic internal ID as parameters.
func (e *EpicService) GetEpicLinks() server.ServerTool {
	return server.ServerTool{
		Handler: e.getEpicLinks,
		Tool: mcpargs.NewTool("get_epic_links", getEpicLinksArgs{},
			mcp.WithDescription("GetEpicLinks gets all child epics of an epic."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (e *EpicService) getEpicLinks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getEpicLinksArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	epics, _, err := e.client.Epics.GetEpicLinks(args.GroupID.Value(), args.EpicIID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetEpicLinks(%q, %d): %w", args.GroupID.Value(), args.EpicIID, err)
	}

	return newToolResultJSON(epics)
}

type listEpicIssuesArgs struct {
	GroupID mcpargs.ID `mcp_desc:"ID of the group either in owner/namespace format or the numeric group ID" mcp_required:"true"`
	EpicIID int        `mcp_desc:"Internal ID of the epic to fetch issues for" mcp_required:"true"`
}

// ListEpicIssues returns a ServerTool for listing all issues assigned to a specific epic.
// The tool accepts a group ID and an epic internal ID as parameters.
func (e *EpicService) ListEpicIssues() server.ServerTool {
	return server.ServerTool{
		Handler: e.listEpicIssues,
		Tool: mcpargs.NewTool("list_epic_issues", listEpicIssuesArgs{},
			mcp.WithDescription("Returns a list of issues assigned to the provided epic"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (e *EpicService) listEpicIssues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listEpicIssuesArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opts = gitlab.ListOptions{
			PerPage: maxPerPage,
		}
		issues []*gitlab.Issue
	)

	// Fetch issues from GitLab
	for {
		is, resp, err := e.client.EpicIssues.ListEpicIssues(args.GroupID.Value(), args.EpicIID, &opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListEpicIssues(%q, %d): %w", args.GroupID.Value(), args.EpicIID, err)
		}

		issues = append(issues, is...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return newToolResultJSON(issues)
}
