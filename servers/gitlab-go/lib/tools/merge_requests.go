package tools

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/discussions"
	"gitlab.com/fforster/gitlab-mcp/lib/gliter"
	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

const defaultMergeRequestsLimit = 1000

// MergeRequestsServiceInterface defines the interface for merge request-related GitLab operations.
// It provides methods for retrieving and listing merge requests and their associated data such as
// approvals, commits, changes, participants, pipelines, and dependencies.
type MergeRequestsServiceInterface interface { //nolint:interfacebloat
	// AddTo registers all merge request-related tools with the provided MCPServer.
	AddTo(srv *server.MCPServer)

	// ListUserMergeRequests returns a tool for listing all merge requests authored by or assigned to a specific user for review.
	ListUserMergeRequests() server.ServerTool

	// ListProjectMergeRequests returns a tool for listing all merge requests within a specific project.
	ListProjectMergeRequests() server.ServerTool

	// ListGroupMergeRequests returns a tool for listing all merge requests within a specific group.
	ListGroupMergeRequests() server.ServerTool

	// GetMergeRequest returns a tool for fetching a specific merge request by its ID.
	GetMergeRequest() server.ServerTool

	// GetMergeRequestApprovals returns a tool for fetching approval information for a specific merge request.
	GetMergeRequestApprovals() server.ServerTool

	// GetMergeRequestCommits returns a tool for fetching all commits associated with a specific merge request.
	GetMergeRequestCommits() server.ServerTool

	// ListMergeRequestDiffs returns a tool for listing diffs of the files changed in a merge request.
	ListMergeRequestDiffs() server.ServerTool

	// GetMergeRequestParticipants returns a tool for fetching all participants of a specific merge request.
	GetMergeRequestParticipants() server.ServerTool

	// GetMergeRequestReviewers returns a tool for fetching all reviewers of a specific merge request.
	GetMergeRequestReviewers() server.ServerTool

	// ListMergeRequestPipelines returns a tool for listing all CI/CD pipelines for a specific merge request.
	ListMergeRequestPipelines() server.ServerTool

	// GetIssuesClosedOnMerge returns a tool for fetching all issues that would be closed by merging a specific merge request.
	GetIssuesClosedOnMerge() server.ServerTool

	// GetMergeRequestDependencies returns a tool for fetching all dependencies of a specific merge request.
	GetMergeRequestDependencies() server.ServerTool

	// EditMergeRequest returns a tool for updating an existing merge request.
	EditMergeRequest() server.ServerTool

	// ListDraftNotes returns a tool for fetching draft MR notes.
	ListDraftNotes() server.ServerTool
}

// NewMergeRequestsTools creates a new instance of MergeRequestsServiceInterface with the provided GitLab client
// and current user. It returns an implementation that can be used to interact with GitLab's merge request API.
func NewMergeRequestsTools(client *gitlab.Client, currentUser string) *MergeRequestsService {
	return &MergeRequestsService{
		client:      client,
		currentUser: currentUser,
	}
}

type MergeRequestsService struct {
	client      *gitlab.Client
	currentUser string
}

// AddTo registers all merge request-related tools with the provided MCPServer.
// It adds tools for listing, retrieving, and managing merge requests and their associated data.
func (m *MergeRequestsService) AddTo(srv *server.MCPServer) {
	srv.AddTools(
		m.ListUserMergeRequests(),
		m.ListProjectMergeRequests(),
		m.ListGroupMergeRequests(),
		m.GetMergeRequest(),
		m.GetMergeRequestApprovals(),
		m.GetMergeRequestCommits(),
		m.ListMergeRequestDiffs(),
		m.GetMergeRequestParticipants(),
		m.GetMergeRequestReviewers(),
		m.ListMergeRequestPipelines(),
		m.GetIssuesClosedOnMerge(),
		m.GetMergeRequestDependencies(),
		m.EditMergeRequest(),
		m.ListDraftNotes(),
	)
}

// detailedMergeRequest combines gitlab.MergeRequest with its discussions and diffs.
type detailedMergeRequest struct {
	MergeRequest *gitlab.MergeRequest       `json:"merge_request"`
	Discussions  []*gitlab.Discussion       `json:"discussions,omitempty"`
	Diffs        []*gitlab.MergeRequestDiff `json:"diffs,omitempty"`
}

type listUserMergeRequestsArgs struct {
	Username string `mcp_desc:"Filter merge requests by username. If left blank, returns merge requests for the authenticated user"`
	State    string `mcp_desc:"Return all merge requests or just those that are opened, closed, or merged" mcp_enum:"all,opened,closed,merged"`
	Role     string `mcp_desc:"Specify whether to list merge requests where the user is the author or reviewer" mcp_enum:"author,reviewer"`
	Limit    int    `mcp_desc:"The maximum number of merge requests to return. Defaults to 1000."`
}

// ListUserMergeRequests returns a ServerTool for listing all merge requests authored by or assigned to a specific user for review.
// The tool accepts optional username and state filter parameters, as well as a role parameter to determine
// whether to list merge requests where the user is the author or a reviewer.
func (m *MergeRequestsService) ListUserMergeRequests() server.ServerTool {
	return server.ServerTool{
		Handler: m.listUserMergeRequests,
		Tool: mcpargs.NewTool("list_user_merge_requests", listUserMergeRequestsArgs{},
			mcp.WithDescription("Get all merge requests authored by or assigned to a user for review"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) listUserMergeRequests(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listUserMergeRequestsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opts = gitlab.ListMergeRequestsOptions{
			Scope: gitlab.Ptr("all"),
			ListOptions: gitlab.ListOptions{
				PerPage: maxPerPage,
			},
		}
		mergeRequests []*gitlab.BasicMergeRequest
	)

	// Parse optional parameters
	username := m.currentUser
	if args.Username != "" {
		username = args.Username
	}

	switch args.Role {
	case "author", "":
		opts.AuthorUsername = gitlab.Ptr(username)
	case "reviewer":
		opts.ReviewerUsername = gitlab.Ptr(username)
	default:
		return nil, fmt.Errorf("%w: invalid role: %q, must be one of 'author' or 'reviewer'", ErrArgumentType, args.Role)
	}

	if args.State != "" {
		opts.State = gitlab.Ptr(args.State)
	}

	if args.Limit <= 0 {
		args.Limit = defaultMergeRequestsLimit
	}

	nextPage := func(opts *gitlab.ListMergeRequestsOptions, page int) {
		opts.ListOptions.Page = page
	}

	iter := gliter.All(ctx, m.client.MergeRequests.ListMergeRequests, opts, nextPage)
	iter = gliter.Limited(iter, args.Limit)

	for mr, err := range iter {
		if err != nil {
			return nil, fmt.Errorf("ListMergeRequests(): %w", err)
		}

		mergeRequests = append(mergeRequests, mr)
	}

	return newToolResultJSON(mergeRequests)
}

type listProjectMergeRequestsArgs struct {
	ProjectID mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	State     string     `mcp_desc:"Return all merge requests or just those that are opened, closed, or merged" mcp_enum:"all,opened,closed,merged"`
	Limit     int        `mcp_desc:"The maximum number of merge requests to return. Defaults to 1000."`
}

// ListProjectMergeRequests returns a ServerTool for listing all merge requests within a specific project.
// The tool accepts a project ID and optional state filter parameters.
func (m *MergeRequestsService) ListProjectMergeRequests() server.ServerTool {
	return server.ServerTool{
		Handler: m.listProjectMergeRequests,
		Tool: mcpargs.NewTool("list_project_merge_requests", listProjectMergeRequestsArgs{},
			mcp.WithDescription("Get all merge requests for this project"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) listProjectMergeRequests(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listProjectMergeRequestsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opts = gitlab.ListProjectMergeRequestsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: maxPerPage,
			},
		}
		mergeRequests []*gitlab.BasicMergeRequest
	)

	// Parse optional parameters
	if args.State != "" {
		opts.State = gitlab.Ptr(args.State)
	}

	if args.Limit <= 0 {
		args.Limit = defaultMergeRequestsLimit
	}

	nextPage := func(opts *gitlab.ListProjectMergeRequestsOptions, page int) {
		opts.Page = page
	}

	iter := gliter.AllWithID(ctx, args.ProjectID.Value(), m.client.MergeRequests.ListProjectMergeRequests, opts, nextPage)
	iter = gliter.Limited(iter, args.Limit)

	for mr, err := range iter {
		if err != nil {
			return nil, fmt.Errorf("ListProjectMergeRequests(%v): %w", args.ProjectID.Value(), err)
		}

		mergeRequests = append(mergeRequests, mr)
	}

	return newToolResultJSON(mergeRequests)
}

type listGroupMergeRequestsArgs struct {
	GroupID mcpargs.ID `mcp_desc:"ID of the group either in owner/namespace format or the numeric group ID" mcp_required:"true"`
	State   string     `mcp_desc:"Return all merge requests or just those that are opened, closed, or merged" mcp_enum:"all,opened,closed,merged"`
	Limit   int        `mcp_desc:"The maximum number of merge requests to return. Defaults to 1000."`
}

// ListGroupMergeRequests returns a ServerTool for listing all merge requests within a specific group.
// The tool accepts a group ID and optional state filter parameters.
func (m *MergeRequestsService) ListGroupMergeRequests() server.ServerTool {
	return server.ServerTool{
		Handler: m.listGroupMergeRequests,
		Tool: mcpargs.NewTool("list_group_merge_requests", listGroupMergeRequestsArgs{},
			mcp.WithDescription("Get all merge requests for this group"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) listGroupMergeRequests(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listGroupMergeRequestsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opts = gitlab.ListGroupMergeRequestsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: maxPerPage,
			},
		}
		mergeRequests []*gitlab.BasicMergeRequest
	)

	// Parse optional parameters
	if args.State != "" {
		opts.State = gitlab.Ptr(args.State)
	}

	if args.Limit <= 0 {
		args.Limit = defaultMergeRequestsLimit
	}

	nextPage := func(opts *gitlab.ListGroupMergeRequestsOptions, page int) {
		opts.Page = page
	}

	iter := gliter.AllWithID(ctx, args.GroupID.Value(), m.client.MergeRequests.ListGroupMergeRequests, opts, nextPage)
	iter = gliter.Limited(iter, args.Limit)

	for mr, err := range iter {
		if err != nil {
			return nil, fmt.Errorf("ListGroupMergeRequests(%v): %w", args.GroupID.Value(), err)
		}

		mergeRequests = append(mergeRequests, mr)
	}

	return newToolResultJSON(mergeRequests)
}

type getMergeRequestArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// GetMergeRequest returns a ServerTool for fetching a specific merge request by its ID, including discussions and diffs.
// If fetching any part fails, an error is returned. Data is fetched concurrently.
func (m *MergeRequestsService) GetMergeRequest() server.ServerTool {
	return server.ServerTool{
		Handler: m.getMergeRequest,
		Tool: mcpargs.NewTool("get_merge_request", getMergeRequestArgs{},
			mcp.WithDescription("Get a single merge request, including its discussions and code changes (diffs)."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) getMergeRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getMergeRequestArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		ret   detailedMergeRequest
		wg    sync.WaitGroup
		errCh = make(chan error)
	)

	// For MR, discussions, diffs
	wg.Add(3)

	// Goroutine to fetch Merge Request details
	go func() {
		defer wg.Done()

		var (
			opts = gitlab.GetMergeRequestsOptions{}
			err  error
		)

		ret.MergeRequest, _, err = m.client.MergeRequests.GetMergeRequest(args.ProjectID.Value(), args.MergeRequestIID, &opts, gitlab.WithContext(ctx))
		if err != nil {
			errCh <- fmt.Errorf("getting merge request IID %d (project %v): %w", args.MergeRequestIID, args.ProjectID.Value(), err)
		}
	}()

	// Goroutine to fetch Discussions
	go func() {
		defer wg.Done()

		discMgr, err := discussions.NewMergeRequestDiscussion(m.client, args.ProjectID, args.MergeRequestIID)
		if err != nil {
			errCh <- fmt.Errorf("creating discussion manager for MR IID %d (project %v): %w", args.MergeRequestIID, args.ProjectID.Value(), err)
			return
		}

		// Assuming non-confidential discussions by default.
		// A 'confidential' argument could be added to 'getMergeRequestArgs' if control over this is needed.
		const confidential = false

		ret.Discussions, err = discMgr.List(ctx, confidential)
		if err != nil {
			errCh <- fmt.Errorf("listing discussions for MR IID %d (project %v): %w", args.MergeRequestIID, args.ProjectID.Value(), err)
		}
	}()

	// Goroutine to fetch Diffs
	go func() {
		defer wg.Done()

		var err error

		ret.Diffs, err = m.listAllMergeRequestDiffs(ctx, args.ProjectID.Value(), args.MergeRequestIID)
		if err != nil {
			errCh <- fmt.Errorf("listing diffs for MR IID %d (project %v): %w", args.MergeRequestIID, args.ProjectID.Value(), err)
		}
	}()

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errs error
	for err := range errCh {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return nil, errs
	}

	return newToolResultJSON(ret)
}

type getMergeRequestApprovalsArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// GetMergeRequestApprovals returns a ServerTool for fetching approval information for a specific merge request.
// The tool accepts a project ID and a merge request internal ID as parameters.
func (m *MergeRequestsService) GetMergeRequestApprovals() server.ServerTool {
	return server.ServerTool{
		Handler: m.getMergeRequestApprovals,
		Tool: mcpargs.NewTool("get_merge_request_approvals", getMergeRequestApprovalsArgs{},
			mcp.WithDescription("Get approvals for a merge request"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) getMergeRequestApprovals(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getMergeRequestApprovalsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	approvals, _, err := m.client.MergeRequests.GetMergeRequestApprovals(args.ProjectID.Value(), args.MergeRequestIID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetMergeRequestApprovals(%q, %d): %w", args.ProjectID.Value(), args.MergeRequestIID, err)
	}

	return newToolResultJSON(approvals)
}

type getMergeRequestCommitsArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// GetMergeRequestCommits returns a ServerTool for fetching all commits associated with a specific merge request.
// The tool accepts a project ID and a merge request internal ID as parameters.
func (m *MergeRequestsService) GetMergeRequestCommits() server.ServerTool {
	return server.ServerTool{
		Handler: m.getMergeRequestCommits,
		Tool: mcpargs.NewTool("get_merge_request_commits", getMergeRequestCommitsArgs{},
			mcp.WithDescription("Get all commits associated with a merge request"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) getMergeRequestCommits(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getMergeRequestCommitsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opt = gitlab.GetMergeRequestCommitsOptions{
			PerPage: maxPerPage,
		}
		commits []*gitlab.Commit
	)

	for {
		c, resp, err := m.client.MergeRequests.GetMergeRequestCommits(args.ProjectID.Value(), args.MergeRequestIID, &opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("GetMergeRequestCommits(%q, %d): %w", args.ProjectID.Value(), args.MergeRequestIID, err)
		}

		commits = append(commits, c...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(commits)
}

type listMergeRequestDiffsArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// ListMergeRequestDiffs returns a ServerTool for listing all versions of diffs for a specific merge request.
func (m *MergeRequestsService) ListMergeRequestDiffs() server.ServerTool {
	return server.ServerTool{
		Handler: m.listMergeRequestDiffs,
		Tool: mcpargs.NewTool("list_merge_request_diffs", listMergeRequestDiffsArgs{},
			mcp.WithDescription("Returns the changes made to files by the merge request. Diffs are presented in the unified diff format."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

// listMergeRequestDiffs is the handler for the ListMergeRequestDiffs tool.
func (m *MergeRequestsService) listMergeRequestDiffs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listMergeRequestDiffsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	diffs, err := m.listAllMergeRequestDiffs(ctx, args.ProjectID.Value(), args.MergeRequestIID)
	if err != nil {
		return nil, err
	}

	return newToolResultJSON(diffs)
}

// listAllMergeRequestDiffs fetches all diffs for a given merge request, handling pagination.
func (m *MergeRequestsService) listAllMergeRequestDiffs(ctx context.Context, projectID any, mrIID int) ([]*gitlab.MergeRequestDiff, error) {
	var allDiffs []*gitlab.MergeRequestDiff

	opts := gitlab.ListMergeRequestDiffsOptions{
		ListOptions: gitlab.ListOptions{PerPage: maxPerPage},
		Unidiff:     gitlab.Ptr(true),
	}

	for {
		diffs, resp, err := m.client.MergeRequests.ListMergeRequestDiffs(projectID, mrIID, &opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("listing merge request diffs for MR IID %d (project %v): %w", mrIID, projectID, err)
		}

		allDiffs = append(allDiffs, diffs...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return allDiffs, nil
}

type getMergeRequestParticipantsArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// GetMergeRequestParticipants returns a ServerTool for fetching all participants of a specific merge request.
// The tool accepts a project ID and a merge request internal ID as parameters.
func (m *MergeRequestsService) GetMergeRequestParticipants() server.ServerTool {
	return server.ServerTool{
		Handler: m.getMergeRequestParticipants,
		Tool: mcpargs.NewTool("get_merge_request_participants", getMergeRequestParticipantsArgs{},
			mcp.WithDescription("Get a list of merge request participants"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) getMergeRequestParticipants(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getMergeRequestParticipantsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	participants, _, err := m.client.MergeRequests.GetMergeRequestParticipants(args.ProjectID.Value(), args.MergeRequestIID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetMergeRequestParticipants(%q, %d): %w", args.ProjectID.Value(), args.MergeRequestIID, err)
	}

	return newToolResultJSON(participants)
}

type getMergeRequestReviewersArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// GetMergeRequestReviewers returns a ServerTool for fetching all reviewers of a specific merge request.
// The tool accepts a project ID and a merge request internal ID as parameters.
func (m *MergeRequestsService) GetMergeRequestReviewers() server.ServerTool {
	return server.ServerTool{
		Handler: m.getMergeRequestReviewers,
		Tool: mcpargs.NewTool("get_merge_request_reviewers", getMergeRequestReviewersArgs{},
			mcp.WithDescription("Get a list of merge request reviewers"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) getMergeRequestReviewers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getMergeRequestReviewersArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	reviewers, _, err := m.client.MergeRequests.GetMergeRequestReviewers(args.ProjectID.Value(), args.MergeRequestIID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetMergeRequestReviewers(%q, %d): %w", args.ProjectID.Value(), args.MergeRequestIID, err)
	}

	return newToolResultJSON(reviewers)
}

type listMergeRequestPipelinesArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// ListMergeRequestPipelines returns a ServerTool for listing all CI/CD pipelines for a specific merge request.
// The tool accepts a project ID and a merge request internal ID as parameters.
func (m *MergeRequestsService) ListMergeRequestPipelines() server.ServerTool {
	return server.ServerTool{
		Handler: m.listMergeRequestPipelines,
		Tool: mcpargs.NewTool("list_merge_request_pipelines", listMergeRequestPipelinesArgs{},
			mcp.WithDescription("Get a list of merge request pipelines"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) listMergeRequestPipelines(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listMergeRequestPipelinesArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	pipelines, _, err := m.client.MergeRequests.ListMergeRequestPipelines(args.ProjectID.Value(), args.MergeRequestIID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("ListMergeRequestPipelines(%q, %d): %w", args.ProjectID.Value(), args.MergeRequestIID, err)
	}

	return newToolResultJSON(pipelines)
}

type getIssuesClosedOnMergeArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// GetIssuesClosedOnMerge returns a ServerTool for fetching all issues that would be closed by merging a specific merge request.
// The tool accepts a project ID and a merge request internal ID as parameters.
func (m *MergeRequestsService) GetIssuesClosedOnMerge() server.ServerTool {
	return server.ServerTool{
		Handler: m.getIssuesClosedOnMerge,
		Tool: mcpargs.NewTool("get_issues_closed_on_merge", getIssuesClosedOnMergeArgs{},
			mcp.WithDescription("Get all the issues that would be closed by merging the provided merge request"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) getIssuesClosedOnMerge(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getIssuesClosedOnMergeArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opt = gitlab.GetIssuesClosedOnMergeOptions{
			PerPage: maxPerPage,
		}
		issues []*gitlab.Issue
	)

	for {
		is, resp, err := m.client.MergeRequests.GetIssuesClosedOnMerge(args.ProjectID.Value(), args.MergeRequestIID, &opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("GetIssuesClosedOnMerge(%q, %d): %w", args.ProjectID.Value(), args.MergeRequestIID, err)
		}

		issues = append(issues, is...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(issues)
}

type getMergeRequestDependenciesArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// GetMergeRequestDependencies returns a ServerTool for fetching all dependencies of a specific merge request.
// The tool accepts a project ID and a merge request internal ID as parameters.
func (m *MergeRequestsService) GetMergeRequestDependencies() server.ServerTool {
	return server.ServerTool{
		Handler: m.getMergeRequestDependencies,
		Tool: mcpargs.NewTool("get_merge_request_dependencies", getMergeRequestDependenciesArgs{},
			mcp.WithDescription("Get merge request dependencies"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) getMergeRequestDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getMergeRequestDependenciesArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	dependencies, _, err := m.client.MergeRequests.GetMergeRequestDependencies(args.ProjectID.Value(), args.MergeRequestIID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetMergeRequestDependencies(%q, %d): %w", args.ProjectID.Value(), args.MergeRequestIID, err)
	}

	return newToolResultJSON(dependencies)
}

type editMergeRequestArgs struct {
	ProjectID          mcpargs.ID           `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID    int                  `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
	Title              string               `mcp_desc:"Changes the title of the merge request"`
	Description        string               `mcp_desc:"Changes the description of the merge request. The description uses GitLab Flavored Markdown"`
	TargetBranch       string               `mcp_desc:"Changes the target branch of the merge request"`
	AssigneeIDs        string               `mcp_desc:"Comma-separated list of user IDs to assign the merge request to. Pass '-' to clear all assignees. Omit the parameter to leave assignees unchanged"`
	ReviewerIDs        string               `mcp_desc:"Comma-separated list of user IDs to set as reviewers for the merge request. Pass '-' to clear all reviewers. Omit the parameter to leave reviewers unchanged"`
	MilestoneID        int                  `mcp_desc:"The ID of a milestone to assign the merge request to"`
	AddLabels          string               `mcp_desc:"Comma-separated label names to add to the merge request"`
	RemoveLabels       string               `mcp_desc:"Comma-separated label names to remove from the merge request"`
	StateEvent         string               `mcp_desc:"The state of the merge request. Use 'close' to close the merge request or 'reopen' to reopen a closed merge request. Omit to keep the merge request state unchanged" mcp_enum:"close,reopen"`
	RemoveSourceBranch mcpargs.OptionalBool `mcp_desc:"Flag indicating if the merge request should remove the source branch when merging"`
	Squash             mcpargs.OptionalBool `mcp_desc:"If true, squash all commits into a single commit on merge"`
	DiscussionLocked   mcpargs.OptionalBool `mcp_desc:"Flag indicating if the merge request's discussion is locked. If true, only project members can add or edit comments"`
	AllowCollaboration mcpargs.OptionalBool `mcp_desc:"Allow commits from members who can merge to the target branch"`
}

// EditMergeRequest returns a ServerTool for updating an existing merge request.
// The tool can modify various aspects of the merge request including title, description, assignees,
// reviewers, labels, milestone, target branch, and state.
func (m *MergeRequestsService) EditMergeRequest() server.ServerTool {
	return server.ServerTool{
		Handler: m.editMergeRequest,
		Tool: mcpargs.NewTool("edit_merge_request", editMergeRequestArgs{},
			mcp.WithDescription("Updates an existing GitLab merge request. You can modify the merge request's title and description, add or remove labels, assign or unassign users, change reviewers, change the milestone, close or reopen the merge request, and control discussion settings."),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
		),
	}
}

func (m *MergeRequestsService) editMergeRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args editMergeRequestArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	opts := gitlab.UpdateMergeRequestOptions{}

	// Set basic fields if provided
	if args.Title != "" {
		opts.Title = gitlab.Ptr(args.Title)
	}

	if args.Description != "" {
		opts.Description = gitlab.Ptr(args.Description)
	}

	if args.TargetBranch != "" {
		opts.TargetBranch = gitlab.Ptr(args.TargetBranch)
	}

	if args.MilestoneID != 0 {
		opts.MilestoneID = gitlab.Ptr(args.MilestoneID)
	}

	// Sets of user IDs
	opts.AssigneeIDs = parseUserIDs(args.AssigneeIDs)
	opts.ReviewerIDs = parseUserIDs(args.ReviewerIDs)

	// Sets of labels
	opts.AddLabels = newLabelOptions(args.AddLabels)
	opts.RemoveLabels = newLabelOptions(args.RemoveLabels)

	switch args.StateEvent {
	case "close", "reopen":
		opts.StateEvent = gitlab.Ptr(args.StateEvent)
	case "":
		// no-op
	default:
		return nil, fmt.Errorf("%w: invalid state_event: %q, must be one of 'close' or 'reopen'", ErrArgumentType, args.StateEvent)
	}

	// Handle boolean flags using the parseBoolString utility
	opts.RemoveSourceBranch = args.RemoveSourceBranch.Ptr()
	opts.Squash = args.Squash.Ptr()

	opts.DiscussionLocked = args.DiscussionLocked.Ptr()
	opts.AllowCollaboration = args.AllowCollaboration.Ptr()

	mergeRequest, _, err := m.client.MergeRequests.UpdateMergeRequest(args.ProjectID.Value(), args.MergeRequestIID, &opts, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("UpdateMergeRequest(%q, %d): %w", args.ProjectID.Value(), args.MergeRequestIID, err)
	}

	return newToolResultJSON(mergeRequest)
}

type listDraftNotesArgs struct {
	ProjectID       mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	MergeRequestIID int        `mcp_desc:"The internal ID of the merge request" mcp_required:"true"`
}

// ListDraftNotes returns a ServerTool for listing all draft notes for a specific merge request.
func (m *MergeRequestsService) ListDraftNotes() server.ServerTool {
	return server.ServerTool{
		Handler: m.listDraftNotes,
		Tool: mcpargs.NewTool("list_draft_notes", listDraftNotesArgs{},
			mcp.WithDescription("Returns a list of draft notes for the merge request. Draft notes are pending merge request review comments that have not yet been published."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

// listDraftNotes returns all draft notes for the merge request.
func (m *MergeRequestsService) listDraftNotes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listDraftNotesArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opt = &gitlab.ListDraftNotesOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: maxPerPage,
			},
		}
		allNotes []*gitlab.DraftNote
	)

	for {
		notes, resp, err := m.client.DraftNotes.ListDraftNotes(args.ProjectID.Value(), args.MergeRequestIID, opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListDraftNotes(%v, %d): %w", args.ProjectID.Value(), args.MergeRequestIID, err)
		}

		allNotes = append(allNotes, notes...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(allNotes)
}
