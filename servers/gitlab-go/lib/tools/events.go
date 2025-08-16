package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

type EventsServiceInterface interface {
	// AddTo registers all issue-related tools with the provided MCPServer.
	AddTo(srv *server.MCPServer)

	ListUserEvents() server.ServerTool
}

func NewEventsTools(client *gitlab.Client) *EventsService {
	return &EventsService{client: client}
}

type EventsService struct {
	client *gitlab.Client
}

// AddTo registers all issue-related tools with the provided MCPServer.
// It adds tools for listing, retrieving, and managing issues and their related merge requests.
func (e *EventsService) AddTo(srv *server.MCPServer) {
	srv.AddTools(
		e.ListUserEvents(),
	)
}

type listEventArgs struct {
	Username   string `mcp_desc:"The username for which to load events, defaults to the current user"`
	Before     string `mcp_desc:"Load all events with a creation date before this date (format: YYYY-MM-DD). When both Before and After limits are missing, only 100 events are returned."`
	After      string `mcp_desc:"Load all events with a creation date after this date(format: YYYY-MM-DD). When both Before and After limits are missing, only 100 events are returned."`
	TargetType string `mcp_desc:"Filter events for a certain target type. If omitted, all target types are returned" mcp_enum:"epic,issue,merge_request,milestone,note,project,snippet,user"`
	ActionType string `mcp_desc:"Filter events for a certain action type. If omitted, all action types are returned" mcp_enum:"approved,closed,commented,created,destroyed,expired,joined,left,merged,pushed,reopened,updated"`
}

func (e *EventsService) ListUserEvents() server.ServerTool {
	return server.ServerTool{
		Handler: e.listUserEvents,
		Tool: mcpargs.NewTool("list_user_events", listEventArgs{},
			mcp.WithDescription("Use this tool to review event activity of a user."+
				"Events can include a wide range of actions including things like joining projects, "+
				"commenting on issues, pushing changes to merge requests. The events are returned from most recent to oldest."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

const (
	sortOrderDescending = "desc"
)

func (e *EventsService) listUserEvents(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listEventArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, fmt.Errorf("mcpargs.Unmarshal: %w", err)
	}

	if args.Username == "" {
		u, _, err := e.client.Users.CurrentUser(gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("CurrentUser: %w", err)
		}

		args.Username = u.Username
	}

	opts, err := listContributionEventsOptions(args)
	if err != nil {
		return nil, err
	}

	var events []*gitlab.ContributionEvent

	for {
		eventPage, response, err := e.client.Users.ListUserContributionEvents(args.Username, opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListUserContributionEvents: %w", err)
		}

		events = append(events, eventPage...)

		// Limit to one page if no boundaries were set. Loading all events will not
		// fit in a context anyway.
		if args.Before == "" || args.After == "" {
			break
		}

		if response.NextPage == 0 {
			break
		}

		opts.Page = response.NextPage
		opts.PerPage = maxPerPage
	}

	return newToolResultJSON(events)
}

func listContributionEventsOptions(args listEventArgs) (*gitlab.ListContributionEventsOptions, error) {
	opts := gitlab.ListContributionEventsOptions{
		Sort: gitlab.Ptr(sortOrderDescending),
		ListOptions: gitlab.ListOptions{
			PerPage: maxPerPage,
		},
	}

	if args.TargetType != "" {
		opts.TargetType = gitlab.Ptr(gitlab.EventTargetTypeValue(args.TargetType))
	}

	if args.ActionType != "" {
		opts.Action = gitlab.Ptr(gitlab.EventTypeValue(args.ActionType))
	}

	if args.Before != "" {
		parsedDate, err := gitlab.ParseISOTime(args.Before)
		if err != nil {
			return nil, fmt.Errorf("gitlab.ParseISOTime: %w", err)
		}

		opts.Before = &parsedDate
	}

	if args.After != "" {
		parsedDate, err := gitlab.ParseISOTime(args.After)
		if err != nil {
			return nil, fmt.Errorf("gitlab.ParseISOTime: %w", err)
		}

		opts.After = &parsedDate
	}

	return &opts, nil
}
