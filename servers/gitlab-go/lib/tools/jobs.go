package tools

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

// JobsServiceInterface defines the interface for job-related GitLab operations.
// It provides methods for retrieving and managing jobs and their artifacts.
type JobsServiceInterface interface {
	// AddTo registers all job-related tools with the provided MCPServer.
	AddTo(srv *server.MCPServer)

	// ListPipelineJobs returns a tool for listing all jobs in a specific pipeline.
	ListPipelineJobs() server.ServerTool

	// ListDownstreamPipelines returns a tool for listing all downstream pipeline triggers in a specific pipeline.
	ListDownstreamPipelines() server.ServerTool

	// GetJob returns a tool for fetching a specific job by its ID.
	GetJob() server.ServerTool

	// DownloadJobArtifactsFile returns a tool for downloading a specific artifact file from a job.
	DownloadJobArtifactsFile() server.ServerTool

	// DownloadJobLog returns a tool for downloading the log file of a specific job.
	DownloadJobLog() server.ServerTool

	// RetryJob returns a tool for retrying a failed job.
	RetryJob() server.ServerTool

	// TriggerManualJob returns a tool for manually triggering a job.
	TriggerManualJob() server.ServerTool
}

// NewJobsTools creates a new instance of JobsServiceInterface with the provided GitLab client.
// It returns an implementation that can be used to interact with GitLab's jobs API.
func NewJobsTools(client *gitlab.Client) *JobsService {
	return &JobsService{client: client}
}

type JobsService struct {
	client *gitlab.Client
}

// AddTo registers all job-related tools with the provided MCPServer.
// It adds tools for listing, retrieving, and managing jobs and their artifacts.
func (j *JobsService) AddTo(srv *server.MCPServer) {
	srv.AddTools(
		j.ListPipelineJobs(),
		j.ListDownstreamPipelines(),
		j.GetJob(),
		j.DownloadJobArtifactsFile(),
		j.DownloadJobLog(),
		j.RetryJob(),
		j.TriggerManualJob(),
	)
}

type listPipelineJobsArgs struct {
	ProjectID      mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	PipelineID     int        `mcp_desc:"ID of the pipeline" mcp_required:"true"`
	Status         string     `mcp_desc:"Filter jobs by status (comma-separated): created, pending, running, failed, success, canceled, skipped, waiting_for_resource, manual"`
	IncludeRetried bool       `mcp_desc:"Include retried jobs in the response (defaults to false)"`
}

// ListPipelineJobs returns a ServerTool for listing all jobs in a specific pipeline.
// The tool accepts a project ID, pipeline ID, and optional filter parameters.
func (j *JobsService) ListPipelineJobs() server.ServerTool {
	return server.ServerTool{
		Handler: j.listPipelineJobs,
		Tool: mcpargs.NewTool("list_pipeline_jobs", listPipelineJobsArgs{},
			mcp.WithDescription("Get a list of jobs for a pipeline"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (j *JobsService) listPipelineJobs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listPipelineJobsArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opt = &gitlab.ListJobsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: maxPerPage,
			},
			IncludeRetried: gitlab.Ptr(args.IncludeRetried),
		}
		jobs []*gitlab.Job
	)

	// Parse status parameter
	opt.Scope = parseBuildStates(args.Status)

	for {
		j, resp, err := j.client.Jobs.ListPipelineJobs(args.ProjectID.Value(), args.PipelineID, opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListPipelineJobs(%q, %d): %w", args.ProjectID.Value(), args.PipelineID, err)
		}

		jobs = append(jobs, j...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(jobs)
}

type listDownstreamPipelinesArgs struct {
	ProjectID  mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	PipelineID int        `mcp_desc:"ID of the pipeline" mcp_required:"true"`
	Status     string     `mcp_desc:"Filter downstream pipeline triggers by status (comma-separated): created, pending, running, failed, success, canceled, skipped, waiting_for_resource, manual"`
}

// ListDownstreamPipelines returns a ServerTool for listing all downstream pipeline triggers in a specific pipeline.
// The tool accepts a project ID, pipeline ID, and optional filter parameters.
func (j *JobsService) ListDownstreamPipelines() server.ServerTool {
	return server.ServerTool{
		Handler: j.listDownstreamPipelines,
		Tool: mcpargs.NewTool("list_downstream_pipelines", listDownstreamPipelinesArgs{},
			mcp.WithDescription("Get a list of downstream pipeline triggered by a pipeline. Downstream pipelines are represented by a 'trigger job'."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (j *JobsService) listDownstreamPipelines(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listDownstreamPipelinesArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	var (
		opt = &gitlab.ListJobsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: maxPerPage,
			},
		}
		bridges []*gitlab.Bridge
	)

	// Parse status parameter
	opt.Scope = parseBuildStates(args.Status)

	for {
		b, resp, err := j.client.Jobs.ListPipelineBridges(args.ProjectID.Value(), args.PipelineID, opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListPipelineBridges(%q, %d): %w", args.ProjectID.Value(), args.PipelineID, err)
		}

		bridges = append(bridges, b...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return newToolResultJSON(bridges)
}

type getJobArgs struct {
	ProjectID mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	JobID     int        `mcp_desc:"ID of the job" mcp_required:"true"`
}

// GetJob returns a ServerTool for fetching information about a specific job.
// The tool accepts a project ID and a job ID as parameters.
func (j *JobsService) GetJob() server.ServerTool {
	return server.ServerTool{
		Handler: j.getJob,
		Tool: mcpargs.NewTool("get_job", getJobArgs{},
			mcp.WithDescription("Get a single job of a project"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (j *JobsService) getJob(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args getJobArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	job, _, err := j.client.Jobs.GetJob(args.ProjectID.Value(), args.JobID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetJob(%q, %d): %w", args.ProjectID.Value(), args.JobID, err)
	}

	return newToolResultJSON(job)
}

type downloadJobArtifactsFileArgs struct {
	ProjectID    mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	JobID        int        `mcp_desc:"ID of the job" mcp_required:"true"`
	ArtifactPath string     `mcp_desc:"Path to a file inside the artifacts archive" mcp_required:"true"`
}

// DownloadJobArtifactsFile returns a ServerTool for downloading a specific artifact file from a job.
// The tool accepts a project ID, job ID, and artifact path as parameters.
func (j *JobsService) DownloadJobArtifactsFile() server.ServerTool {
	return server.ServerTool{
		Handler: j.downloadJobArtifactsFile,
		Tool: mcpargs.NewTool("download_job_artifacts_file", downloadJobArtifactsFileArgs{},
			mcp.WithDescription("Download a single artifact file from a job"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (j *JobsService) downloadJobArtifactsFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args downloadJobArtifactsFileArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	reader, _, err := j.client.Jobs.DownloadSingleArtifactsFile(args.ProjectID.Value(), args.JobID, args.ArtifactPath, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("DownloadSingleArtifactsFile(%q, %d, %q): %w", args.ProjectID.Value(), args.JobID, args.ArtifactPath, err)
	}

	// Read all the content from the reader
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading artifact content: %w", err)
	}

	return mcp.NewToolResultText(string(content)), nil
}

type downloadJobLogArgs struct {
	ProjectID mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	JobID     int        `mcp_desc:"ID of the job" mcp_required:"true"`
}

// DownloadJobLog returns a ServerTool for downloading the log file of a specific job.
// The tool accepts a project ID and a job ID as parameters.
func (j *JobsService) DownloadJobLog() server.ServerTool {
	return server.ServerTool{
		Handler: j.downloadJobLog,
		Tool: mcpargs.NewTool("download_job_log", downloadJobLogArgs{},
			mcp.WithDescription("Download a log file for a specific job"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (j *JobsService) downloadJobLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args downloadJobLogArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	reader, _, err := j.client.Jobs.GetTraceFile(args.ProjectID.Value(), args.JobID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("GetTraceFile(%q, %d): %w", args.ProjectID.Value(), args.JobID, err)
	}

	// Read all the content from the reader
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading log content: %w", err)
	}

	return mcp.NewToolResultText(string(content)), nil
}

type retryJobArgs struct {
	ProjectID mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	JobID     int        `mcp_desc:"ID of the job" mcp_required:"true"`
}

// RetryJob returns a ServerTool for retrying a failed job.
// The tool accepts a project ID and a job ID as parameters.
func (j *JobsService) RetryJob() server.ServerTool {
	return server.ServerTool{
		Handler: j.retryJob,
		Tool: mcpargs.NewTool("retry_job", retryJobArgs{},
			mcp.WithDescription("Retry a single job of a project"),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(false),
		),
	}
}

func (j *JobsService) retryJob(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args retryJobArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	job, _, err := j.client.Jobs.RetryJob(args.ProjectID.Value(), args.JobID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("RetryJob(%q, %d): %w", args.ProjectID.Value(), args.JobID, err)
	}

	return newToolResultJSON(job)
}

type triggerManualJobArgs struct {
	ProjectID mcpargs.ID `mcp_desc:"ID of the project either in owner/project format or the numeric project ID" mcp_required:"true"`
	JobID     int        `mcp_desc:"ID of the job" mcp_required:"true"`
}

// TriggerManualJob returns a ServerTool for manually triggering a job.
// The tool accepts a project ID, job ID, and optional job variables as parameters.
func (j *JobsService) TriggerManualJob() server.ServerTool {
	return server.ServerTool{
		Handler: j.triggerManualJob,
		Tool: mcpargs.NewTool("trigger_manual_job", triggerManualJobArgs{},
			mcp.WithDescription("Trigger a manual job for a project"),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(false),
		),
	}
}

func (j *JobsService) triggerManualJob(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args triggerManualJobArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	job, _, err := j.client.Jobs.PlayJob(args.ProjectID.Value(), args.JobID, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("PlayJob(%q, %d): %w", args.ProjectID.Value(), args.JobID, err)
	}

	return newToolResultJSON(job)
}

// buildStateValues is a mapping of build state strings to their corresponding gitlab.BuildStateValue constants.
var buildStateValues = map[string]gitlab.BuildStateValue{
	string(gitlab.Created):            gitlab.Created,
	string(gitlab.WaitingForResource): gitlab.WaitingForResource,
	string(gitlab.Preparing):          gitlab.Preparing,
	string(gitlab.Pending):            gitlab.Pending,
	string(gitlab.Running):            gitlab.Running,
	string(gitlab.Success):            gitlab.Success,
	string(gitlab.Failed):             gitlab.Failed,
	string(gitlab.Canceled):           gitlab.Canceled,
	string(gitlab.Skipped):            gitlab.Skipped,
	string(gitlab.Manual):             gitlab.Manual,
	string(gitlab.Scheduled):          gitlab.Scheduled,
}

// parseBuildStates converts a comma-separated string of job statuses into a slice of
// gitlab.BuildStateValue or returns nil if the input is empty or contains no valid statuses.
func parseBuildStates(statuses string) *[]gitlab.BuildStateValue {
	var (
		haveState   = make(map[gitlab.BuildStateValue]bool)
		buildStates []gitlab.BuildStateValue
	)

	for _, statusStr := range strings.Split(statuses, ",") {
		statusStr = strings.TrimSpace(statusStr)
		if state, ok := buildStateValues[statusStr]; ok && !haveState[state] {
			buildStates = append(buildStates, state)
			haveState[state] = true
		}
	}

	if len(buildStates) == 0 {
		return nil
	}

	return gitlab.Ptr(buildStates)
}
