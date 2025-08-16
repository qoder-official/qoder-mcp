package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

// RepositoryServiceInterface defines the interface for repository file-related GitLab operations.
// It provides methods for accessing and manipulating repository files and contents.
type RepositoryServiceInterface interface {
	// AddTo registers all repository file-related tools with the provided MCPServer.
	AddTo(srv *server.MCPServer)

	// ListRepositoryDirectory returns a tool for listing repository files and directories.
	ListRepositoryDirectory() server.ServerTool

	// GetRepositoryFileContents returns a tool for retrieving contents of a file either by blob SHA or by file path.
	GetRepositoryFileContents() server.ServerTool
}

// NewRepositoryTools creates a new instance of RepositoryServiceInterface with the provided GitLab client.
// It returns an implementation that can be used to interact with GitLab's repository files API.
func NewRepositoryTools(client *gitlab.Client) *RepositoryService {
	return &RepositoryService{client: client}
}

type RepositoryService struct {
	client *gitlab.Client
}

// AddTo registers all repository file-related tools with the provided MCPServer.
// It adds tools for listing repository tree and retrieving file contents.
func (r *RepositoryService) AddTo(srv *server.MCPServer) {
	srv.AddTools(
		r.ListRepositoryDirectory(),
		r.GetRepositoryFileContents(),
	)
}

type listRepositoryDirectoryArgs struct {
	ProjectID mcpargs.ID `mcp_desc:"Specifies the project to list files from, using either owner/project format or numeric project ID" mcp_required:"true"`
	Path      string     `mcp_desc:"The path inside the repository to list files from; defaults to the repository root if not provided"`
	Ref       string     `mcp_desc:"Specifies which branch or tag to list files from; defaults to the project's default branch if not provided"`
	Recursive bool       `mcp_desc:"When set to true, lists files in subdirectories recursively instead of just the 'path' level"`
}

// ListRepositoryDirectory returns a ServerTool for listing repository files and directories.
// The tool accepts project ID and various optional parameters for customizing the tree listing.
func (r *RepositoryService) ListRepositoryDirectory() server.ServerTool {
	return server.ServerTool{
		Handler: r.listRepositoryDirectory,
		Tool: mcpargs.NewTool("list_repository_directory", listRepositoryDirectoryArgs{},
			mcp.WithDescription("Get a list of repository files and directories in a project. The returned JSON uses Git terminology, e.g. calling files 'blob' and directories 'tree'."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (r *RepositoryService) listRepositoryDirectory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listRepositoryDirectoryArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	// Prepare options for the API call
	opt := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: maxPerPage,
		},
	}

	// Set optional parameters if provided
	if args.Path != "" {
		opt.Path = gitlab.Ptr(args.Path)
	}

	if args.Ref != "" {
		opt.Ref = gitlab.Ptr(args.Ref)
	}

	opt.Recursive = gitlab.Ptr(args.Recursive)

	// Make the API call to get the repository tree and handle pagination internally
	var allTree []*gitlab.TreeNode

	for {
		tree, resp, err := r.client.Repositories.ListTree(args.ProjectID.Value(), opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListTree(%q): %w", args.ProjectID.Value(), err)
		}

		allTree = append(allTree, tree...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(allTree)
}

type getRepositoryFileContentsArgs struct {
	ProjectID mcpargs.ID `mcp_desc:"Specifies the project to get file contents from, using either owner/project format or numeric project ID" mcp_required:"true"`
	SHA       string     `mcp_desc:"Specifies the blob SHA to get contents from. Provide either 'sha' or 'file_path', but not both"`
	FilePath  string     `mcp_desc:"Specifies the path to the file in the repository. Provide either 'sha' or 'file_path', but not both"`
	Ref       string     `mcp_desc:"Specifies which branch or tag to use when accessing a file by path; defaults to the project's default branch if not provided"`
}

// GetRepositoryFileContents returns a ServerTool for retrieving the contents of a file.
// The tool accepts either a blob SHA or a file path with optional reference (branch/tag).
func (r *RepositoryService) GetRepositoryFileContents() server.ServerTool {
	return server.ServerTool{
		Handler: r.getRepositoryFileContents,
		Tool: mcpargs.NewTool("get_repository_file_contents", getRepositoryFileContentsArgs{},
			mcp.WithDescription("Get the contents of a single file from the repository."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (r *RepositoryService) getRepositoryFileContents(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getRepositoryFileContentsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	if haveSHA, havePath := args.SHA != "", args.FilePath != ""; haveSHA == havePath {
		return mcp.NewToolResultError("exactly one of 'sha' or 'file_path' must be provided"), nil
	}

	// Choose the appropriate API endpoint based on the provided parameters
	if args.SHA != "" {
		// Use RawBlobContent when SHA is provided
		contents, _, err := r.client.Repositories.RawBlobContent(args.ProjectID.Value(), args.SHA, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("RawBlobContent(%q, %q): %w", args.ProjectID.Value(), args.SHA, err)
		}

		return mcp.NewToolResultText(string(contents)), nil
	}

	// Use GetRawFile when FilePath is provided
	opt := gitlab.GetRawFileOptions{}
	if args.Ref != "" {
		opt.Ref = gitlab.Ptr(args.Ref)
	}

	contents, _, err := r.client.RepositoryFiles.GetRawFile(args.ProjectID.Value(), args.FilePath, &opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetRawFile(%q, %q, ref=%q): %w", args.ProjectID.Value(), args.FilePath, args.Ref, err)
	}

	return mcp.NewToolResultText(string(contents)), nil
}
