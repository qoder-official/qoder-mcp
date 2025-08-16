package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/gliter"
	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

const defaultTodosLimit = 100

// TodosServiceInterface defines the interface for todo-related GitLab operations.
type TodosServiceInterface interface {
	// AddTo registers all todo-related tools with the provided MCPServer.
	AddTo(srv *server.MCPServer)

	// ListUserTodos returns a tool for listing all todos for the current user.
	ListUserTodos() server.ServerTool
}

// NewTodosTools creates a new instance of TodosServiceInterface with the provided GitLab client.
func NewTodosTools(client *gitlab.Client) *TodosService {
	return &TodosService{client: client}
}

type TodosService struct {
	client *gitlab.Client
}

// AddTo registers all todo-related tools with the provided MCPServer.
func (t *TodosService) AddTo(srv *server.MCPServer) {
	srv.AddTools(
		t.ListUserTodos(),
		t.CompleteTodoItem(),
		t.CompleteAllTodoItems(),
	)
}

type listUserTodosArgs struct {
	Action    string `mcp_desc:"Filter by the action that caused the todo item" mcp_enum:"assigned,mentioned,build_failed,marked,approval_required,unmergeable,directly_addressed,merge_train_removed,member_access_requested"`
	AuthorID  int    `mcp_desc:"Filter by the ID of the author who created the todo item"`
	ProjectID int    `mcp_desc:"Filter by the ID of the project the todo item belongs to"`
	GroupID   int    `mcp_desc:"Filter by the ID of the group the todo item belongs to"`
	State     string `mcp_desc:"Filter by the state of the todo item, defaults to 'pending'" mcp_enum:"pending,done"`
	Type      string `mcp_desc:"Filter by the type of resource the todo item is associated with" mcp_enum:"Issue,MergeRequest,Commit,Epic,DesignManagement::Design,AlertManagement::Alert,Project,Namespace,Vulnerability,WikiPage::Meta"`
	Limit     int    `mcp_desc:"Maximum number of todos to return. If not set or zero, defaults to 100."`
}

// ListUserTodos returns a ServerTool for listing all todos for the current user.
func (t *TodosService) ListUserTodos() server.ServerTool {
	return server.ServerTool{
		Handler: t.listUserTodos,
		Tool: mcpargs.NewTool("list_user_todos", listUserTodosArgs{},
			mcp.WithDescription("Get all todos for the current user, with optional filtering."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
	}
}

func (t *TodosService) listUserTodos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args listUserTodosArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	if args.Limit <= 0 {
		args.Limit = defaultTodosLimit
	}

	opt := gitlab.ListTodosOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: min(args.Limit, maxPerPage),
		},
	}

	// Parse optional parameters
	opt.Action = todoActions[args.Action]
	opt.State = todoStates[args.State]

	if args.AuthorID != 0 {
		opt.AuthorID = gitlab.Ptr(args.AuthorID)
	}

	if args.ProjectID != 0 {
		opt.ProjectID = gitlab.Ptr(args.ProjectID)
	}

	if args.GroupID != 0 {
		opt.GroupID = gitlab.Ptr(args.GroupID)
	}

	if args.Type != "" {
		opt.Type = gitlab.Ptr(args.Type)
	}

	nextPage := func(opts *gitlab.ListTodosOptions, page int) {
		opts.ListOptions.Page = page
	}

	iter := gliter.All(ctx, t.client.Todos.ListTodos, opt, nextPage)
	iter = gliter.Limited(iter, args.Limit)

	var todos []*gitlab.Todo

	for todo, err := range iter {
		if err != nil {
			return nil, fmt.Errorf("ListTodos(): %w", err)
		}

		todos = append(todos, todo)
	}

	return newToolResultJSON(todos)
}

var todoStates = map[string]*string{
	"pending": gitlab.Ptr("pending"),
	"done":    gitlab.Ptr("done"),
	"":        gitlab.Ptr("pending"), // Default value
}

// todoActions is a map of all possible todo actions.
var todoActions = map[string]*gitlab.TodoAction{
	string(gitlab.TodoAssigned):          gitlab.Ptr(gitlab.TodoAssigned),
	string(gitlab.TodoMentioned):         gitlab.Ptr(gitlab.TodoMentioned),
	string(gitlab.TodoBuildFailed):       gitlab.Ptr(gitlab.TodoBuildFailed),
	string(gitlab.TodoMarked):            gitlab.Ptr(gitlab.TodoMarked),
	string(gitlab.TodoApprovalRequired):  gitlab.Ptr(gitlab.TodoApprovalRequired),
	string(gitlab.TodoDirectlyAddressed): gitlab.Ptr(gitlab.TodoDirectlyAddressed),
}

type completeTodoItemArgs struct {
	ID int `mcp_desc:"The ID of the todo item to mark as done" mcp_required:"true"`
}

// CompleteTodoItem returns a ServerTool for marking a single todo item as done.
// The tool accepts the ID of the todo item to be marked as done.
func (t *TodosService) CompleteTodoItem() server.ServerTool {
	return server.ServerTool{
		Handler: t.completeTodoItem,
		Tool: mcpargs.NewTool("complete_todo_item", completeTodoItemArgs{},
			mcp.WithDescription("Marks a single pending todo item as done"),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithIdempotentHintAnnotation(true),
		),
	}
}

func (t *TodosService) completeTodoItem(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args completeTodoItemArgs
	if err := mcpargs.Unmarshal(request.GetArguments(), &args); err != nil {
		return nil, err //nolint:wrapcheck
	}

	_, err := t.client.Todos.MarkTodoAsDone(args.ID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("MarkTodoAsDone(%d): %w", args.ID, err)
	}

	return mcp.NewToolResultText("success"), nil
}

// CompleteAllTodoItems returns a ServerTool for marking all pending todo items as done.
// This tool takes no parameters as it affects all pending todo items.
func (t *TodosService) CompleteAllTodoItems() server.ServerTool {
	return server.ServerTool{
		Handler: t.completeAllTodoItems,
		Tool: mcp.NewTool("complete_all_todo_items",
			mcp.WithDescription("Marks all pending todo items for the current user as done. Only perform this action when explicitly requested by the user."),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithIdempotentHintAnnotation(false),
		),
	}
}

func (t *TodosService) completeAllTodoItems(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	_, err := t.client.Todos.MarkAllTodosAsDone(gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("MarkAllTodosAsDone: %w", err)
	}

	return mcp.NewToolResultText("success"), nil
}
