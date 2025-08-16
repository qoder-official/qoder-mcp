package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

// SnippetsServiceInterface defines the interface for snippet-related GitLab operations.
// It provides methods for retrieving, listing, creating, updating, and deleting snippets.
type SnippetsServiceInterface interface {
	// AddTo registers all snippet-related tools with the provided MCPServer.
	AddTo(srv *server.MCPServer)

	// ListUserSnippets returns a tool for listing snippets owned by the current user.
	ListUserSnippets() server.ServerTool

	// ListAllSnippets returns a tool for listing all snippets the user has access to.
	ListAllSnippets() server.ServerTool

	// GetSnippet returns a tool for fetching a specific snippet by its ID.
	GetSnippet() server.ServerTool

	// GetSnippetContent returns a tool for fetching the content of a specific snippet.
	GetSnippetContent() server.ServerTool

	// CreateSnippet returns a tool for creating a new snippet.
	CreateSnippet() server.ServerTool

	// UpdateSnippet returns a tool for updating an existing snippet.
	UpdateSnippet() server.ServerTool

	// DeleteSnippet returns a tool for deleting a snippet.
	DeleteSnippet() server.ServerTool
}

const (
	snippetVisibilityPrivate  = "private"
	snippetVisibilityInternal = "internal"
	snippetVisibilityPublic   = "public"
)

// NewSnippetsTools creates a new instance of SnippetsServiceInterface with the provided GitLab client
// and current user. It returns an implementation that can be used to interact with GitLab's snippet API.
func NewSnippetsTools(client *gitlab.Client, currentUser string) *SnippetsService {
	return &SnippetsService{
		client:      client,
		currentUser: currentUser,
	}
}

type SnippetsService struct {
	client      *gitlab.Client
	currentUser string
}

// AddTo registers all snippet-related tools with the provided MCPServer.
// It adds tools for listing, retrieving, creating, updating, and deleting snippets.
func (s *SnippetsService) AddTo(srv *server.MCPServer) {
	srv.AddTools(
		s.ListUserSnippets(),
		s.ListAllSnippets(),
		s.GetSnippet(),
		s.GetSnippetContent(),
		s.CreateSnippet(),
		s.UpdateSnippet(),
		s.DeleteSnippet(),
	)
}

// ListUserSnippets returns a ServerTool for listing snippets owned by the current user.
func (s *SnippetsService) ListUserSnippets() server.ServerTool {
	return server.ServerTool{
		Handler: s.listUserSnippets,
		Tool: mcp.NewTool("list_user_snippets",
			mcp.WithDescription("List snippets owned by the current user"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (s *SnippetsService) listUserSnippets(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var (
		opt = &gitlab.ListSnippetsOptions{
			PerPage: maxPerPage,
		}
		snippets []*gitlab.Snippet
	)

	for {
		snips, resp, err := s.client.Snippets.ListSnippets(opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListSnippets: %w", err)
		}

		snippets = append(snippets, snips...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(snippets)
}

type listAllSnippetsArgs struct {
	IncludePrivate bool `mcp_desc:"Include private snippets in the results. If false (the default), only public snippets are returned"`
}

// ListAllSnippets returns a ServerTool for listing all snippets the user has access to.
// The include_private parameter determines whether to include private snippets or only public ones.
func (s *SnippetsService) ListAllSnippets() server.ServerTool {
	return server.ServerTool{
		Handler: s.listAllSnippets,
		Tool: mcpargs.NewTool("list_all_snippets", listAllSnippetsArgs{},
			mcp.WithDescription("List all snippets the user has access to"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (s *SnippetsService) listAllSnippets(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listAllSnippetsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	if args.IncludePrivate {
		return s.listAllSnippetsIncludingPrivate(ctx)
	}

	return s.listAllPublicSnippets(ctx)
}

// listAllSnippetsIncludingPrivate lists all snippets, private and public.
func (s *SnippetsService) listAllSnippetsIncludingPrivate(ctx context.Context) (*mcp.CallToolResult, error) {
	opt := &gitlab.ListAllSnippetsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: maxPerPage,
		},
	}

	var snippets []*gitlab.Snippet

	for {
		snips, resp, err := s.client.Snippets.ListAllSnippets(opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListAllSnippets: %w", err)
		}

		snippets = append(snippets, snips...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(snippets)
}

// listAllPublicSnippets lists all public snippets.
func (s *SnippetsService) listAllPublicSnippets(ctx context.Context) (*mcp.CallToolResult, error) {
	// List only public snippets
	opt := &gitlab.ExploreSnippetsOptions{
		PerPage: maxPerPage,
	}

	var snippets []*gitlab.Snippet

	for {
		snips, resp, err := s.client.Snippets.ExploreSnippets(opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ExploreSnippets: %w", err)
		}

		snippets = append(snippets, snips...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(snippets)
}

type getSnippetArgs struct {
	SnippetID int `mcp_desc:"The ID of the snippet" mcp_required:"true"`
}

// GetSnippet returns a ServerTool for fetching information about a specific snippet.
// The tool accepts a snippet ID as a parameter.
func (s *SnippetsService) GetSnippet() server.ServerTool {
	return server.ServerTool{
		Handler: s.getSnippet,
		Tool: mcpargs.NewTool("get_snippet", getSnippetArgs{},
			mcp.WithDescription("Returns the metadata of a snippet, such as title and description. File content is not returned."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (s *SnippetsService) getSnippet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getSnippetArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	snippet, _, err := s.client.Snippets.GetSnippet(args.SnippetID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetSnippet(%d): %w", args.SnippetID, err)
	}

	return newToolResultJSON(snippet)
}

type getSnippetContentArgs struct {
	SnippetID int `mcp_desc:"The ID of the snippet" mcp_required:"true"`
}

// GetSnippetContent returns a ServerTool for fetching the content of a specific snippet.
// The tool accepts a snippet ID as a parameter.
func (s *SnippetsService) GetSnippetContent() server.ServerTool {
	return server.ServerTool{
		Handler: s.getSnippetContent,
		Tool: mcpargs.NewTool("get_snippet_content", getSnippetContentArgs{},
			mcp.WithDescription("Get the raw content of a snippet"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (s *SnippetsService) getSnippetContent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getSnippetContentArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	content, _, err := s.client.Snippets.SnippetContent(args.SnippetID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("SnippetContent(%d): %w", args.SnippetID, err)
	}

	return newToolResultJSON(map[string]string{"content": string(content)})
}

type createSnippetArgs struct {
	Title       string `mcp_desc:"The title of the snippet" mcp_required:"true"`
	FileName    string `mcp_desc:"The name of the snippet file" mcp_required:"true"`
	Content     string `mcp_desc:"The content of the snippet" mcp_required:"true"`
	Visibility  string `mcp_desc:"The visibility level of the snippet. Default to private." mcp_enum:"private,internal,public"`
	Description string `mcp_desc:"The description of the snippet"`
}

// CreateSnippet returns a ServerTool for creating a new snippet.
// The tool accepts title, file_name, content, visibility, and description parameters.
func (s *SnippetsService) CreateSnippet() server.ServerTool {
	return server.ServerTool{
		Handler: s.createSnippet,
		Tool: mcpargs.NewTool("create_snippet", createSnippetArgs{},
			mcp.WithDescription("Creates a new snippet with a single file. "+
				"Creating snippets with multiple files is a multi-step process: "+
				"create a snippet with one file using 'create_snippet', "+
				"then add additional files using 'update_snippet' "+
				"with a different file name and 'file_action' set to 'create'. "+
				"When creating a snippet, include its 'web_url' in your response."),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(false),
		),
	}
}

func (s *SnippetsService) createSnippet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args createSnippetArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	// Set default visibility if not provided
	visibility := args.Visibility
	if visibility == "" {
		visibility = snippetVisibilityPrivate
	}

	var visibilityLevel gitlab.VisibilityValue

	switch visibility {
	case snippetVisibilityPrivate:
		visibilityLevel = gitlab.PrivateVisibility
	case snippetVisibilityInternal:
		visibilityLevel = gitlab.InternalVisibility
	case snippetVisibilityPublic:
		visibilityLevel = gitlab.PublicVisibility
	default:
		return nil, fmt.Errorf("%w: invalid visibility level: %s", ErrArgumentType, visibility)
	}

	opt := &gitlab.CreateSnippetOptions{
		Title:      gitlab.Ptr(args.Title),
		Visibility: &visibilityLevel,
		Files: gitlab.Ptr([]*gitlab.CreateSnippetFileOptions{
			{
				FilePath: gitlab.Ptr(args.FileName),
				Content:  gitlab.Ptr(args.Content),
			},
		}),
	}

	// Add description if provided
	if args.Description != "" {
		opt.Description = gitlab.Ptr(args.Description)
	}

	snippet, _, err := s.client.Snippets.CreateSnippet(opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("CreateSnippet: %w", err)
	}

	return newToolResultJSON(snippet)
}

type updateSnippetArgs struct {
	SnippetID   int    `mcp_desc:"The ID of the snippet to update" mcp_required:"true"`
	Title       string `mcp_desc:"The title of the snippet"`
	Description string `mcp_desc:"The description of the snippet"`
	Visibility  string `mcp_desc:"The visibility level of the snippet" mcp_enum:"private,internal,public"`
	FileAction  string `mcp_desc:"The action to perform on the file. Default is 'update'" mcp_enum:"create,update,delete"`
	FileName    string `mcp_desc:"The name of the snippet file" mcp_required:"true"`
	Content     string `mcp_desc:"The content of the snippet" mcp_required:"true"`
}

// UpdateSnippet returns a ServerTool for updating an existing snippet.
// The tool accepts snippet_id and optional title, file_name, content, visibility, and description parameters.
func (s *SnippetsService) UpdateSnippet() server.ServerTool {
	return server.ServerTool{
		Handler: s.updateSnippet,
		Tool: mcpargs.NewTool("update_snippet", updateSnippetArgs{},
			mcp.WithDescription("Update an existing snippet. This tool can create, update, or delete a file in the snippet, as well as update snippet metadata."),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
		),
	}
}

func (s *SnippetsService) updateSnippet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args updateSnippetArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	opt := &gitlab.UpdateSnippetOptions{}

	var fileActionValue gitlab.FileActionValue

	switch args.FileAction {
	case "create":
		fileActionValue = gitlab.FileCreate
	case "update", "":
		fileActionValue = gitlab.FileUpdate
	case "delete":
		fileActionValue = gitlab.FileDelete
	default:
		return nil, fmt.Errorf("%w: invalid file action: %q", ErrArgumentType, args.FileAction)
	}

	opt.Files = gitlab.Ptr([]*gitlab.UpdateSnippetFileOptions{{
		Action:   gitlab.Ptr(string(fileActionValue)),
		FilePath: gitlab.Ptr(args.FileName),
		Content:  gitlab.Ptr(args.Content),
	}})

	// Add optional parameters if provided
	if args.Title != "" {
		opt.Title = gitlab.Ptr(args.Title)
	}

	if args.Description != "" {
		opt.Description = gitlab.Ptr(args.Description)
	}

	if args.Visibility != "" {
		var visibilityLevel gitlab.VisibilityValue

		switch args.Visibility {
		case snippetVisibilityPrivate:
			visibilityLevel = gitlab.PrivateVisibility
		case snippetVisibilityInternal:
			visibilityLevel = gitlab.InternalVisibility
		case snippetVisibilityPublic:
			visibilityLevel = gitlab.PublicVisibility
		default:
			return nil, fmt.Errorf("%w: invalid visibility level: %q", ErrArgumentType, args.Visibility)
		}

		opt.Visibility = gitlab.Ptr(visibilityLevel)
	}

	snippet, _, err := s.client.Snippets.UpdateSnippet(args.SnippetID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("UpdateSnippet(%d): %w", args.SnippetID, err)
	}

	return newToolResultJSON(snippet)
}

type deleteSnippetArgs struct {
	SnippetID int `mcp_desc:"The ID of the snippet to delete" mcp_required:"true"`
}

// DeleteSnippet returns a ServerTool for deleting a snippet.
// The tool accepts a snippet ID as a parameter.
func (s *SnippetsService) DeleteSnippet() server.ServerTool {
	return server.ServerTool{
		Handler: s.deleteSnippet,
		Tool: mcpargs.NewTool("delete_snippet", deleteSnippetArgs{},
			mcp.WithDescription("Delete a snippet"),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithIdempotentHintAnnotation(true),
		),
	}
}

func (s *SnippetsService) deleteSnippet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args deleteSnippetArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	_, err := s.client.Snippets.DeleteSnippet(args.SnippetID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("DeleteSnippet(%d): %w", args.SnippetID, err)
	}

	return newToolResultJSON(map[string]string{"result": "success"})
}
