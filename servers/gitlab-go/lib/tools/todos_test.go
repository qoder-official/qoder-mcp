package tools

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	glabtest "gitlab.com/gitlab-org/api/client-go/testing"
)

func TestListUserTodos(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		setupMock func(*glabtest.MockTodosServiceInterface)
		wantIDs   []int
	}{
		{
			name: "state filtering - pending",
			args: map[string]any{
				"state": "pending",
			},
			setupMock: func(mockTodos *glabtest.MockTodosServiceInterface) {
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListTodosOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Todo, *gitlab.Response, error) {
						if *opts.State != "pending" {
							t.Errorf("expected state 'pending', got %q", *opts.State)
						}

						return []*gitlab.Todo{{ID: 1}, {ID: 2}}, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						}, nil
					})
			},
			wantIDs: []int{1, 2},
		},
		{
			name: "action filtering - assigned",
			args: map[string]any{
				"action": "assigned",
			},
			setupMock: func(mockTodos *glabtest.MockTodosServiceInterface) {
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListTodosOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Todo, *gitlab.Response, error) {
						if opts.Action == nil || string(*opts.Action) != "assigned" {
							t.Errorf("expected action 'assigned', got %v", opts.Action)
						}

						return []*gitlab.Todo{{ID: 3}, {ID: 4}}, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						}, nil
					})
			},
			wantIDs: []int{3, 4},
		},
		{
			name: "author filtering",
			args: map[string]any{
				"author_id": 123,
			},
			setupMock: func(mockTodos *glabtest.MockTodosServiceInterface) {
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListTodosOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Todo, *gitlab.Response, error) {
						if opts.AuthorID == nil || *opts.AuthorID != 123 {
							t.Errorf("expected author_id 123, got %v", opts.AuthorID)
						}

						return []*gitlab.Todo{{ID: 5}}, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						}, nil
					})
			},
			wantIDs: []int{5},
		},
		{
			name: "project filtering",
			args: map[string]any{
				"project_id": 456,
			},
			setupMock: func(mockTodos *glabtest.MockTodosServiceInterface) {
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListTodosOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Todo, *gitlab.Response, error) {
						if opts.ProjectID == nil || *opts.ProjectID != 456 {
							t.Errorf("expected project_id 456, got %v", opts.ProjectID)
						}

						return []*gitlab.Todo{{ID: 6}}, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						}, nil
					})
			},
			wantIDs: []int{6},
		},
		{
			name: "group filtering",
			args: map[string]any{
				"group_id": 789,
			},
			setupMock: func(mockTodos *glabtest.MockTodosServiceInterface) {
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListTodosOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Todo, *gitlab.Response, error) {
						if opts.GroupID == nil || *opts.GroupID != 789 {
							t.Errorf("expected project_id 789, got %v", opts.GroupID)
						}

						return []*gitlab.Todo{{ID: 61}}, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						}, nil
					})
			},
			wantIDs: []int{61},
		},
		{
			name: "type filtering",
			args: map[string]any{
				"type": "MergeRequest",
			},
			setupMock: func(mockTodos *glabtest.MockTodosServiceInterface) {
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListTodosOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Todo, *gitlab.Response, error) {
						if opts.Type == nil || *opts.Type != "MergeRequest" {
							t.Errorf("expected type 'MergeRequest', got %v", opts.Type)
						}

						return []*gitlab.Todo{{ID: 7}}, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						}, nil
					})
			},
			wantIDs: []int{7},
		},
		{
			name: "combined filters",
			args: map[string]any{
				"state":      "done",
				"action":     "assigned",
				"project_id": 789,
			},
			setupMock: func(mockTodos *glabtest.MockTodosServiceInterface) {
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListTodosOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Todo, *gitlab.Response, error) {
						if *opts.State != "done" {
							t.Errorf("expected state 'done', got %q", *opts.State)
						}
						if opts.Action == nil || string(*opts.Action) != "assigned" {
							t.Errorf("expected action 'assigned', got %v", opts.Action)
						}
						if opts.ProjectID == nil || *opts.ProjectID != 789 {
							t.Errorf("expected project_id 789, got %v", opts.ProjectID)
						}

						return []*gitlab.Todo{{ID: 8}, {ID: 9}}, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
						}, nil
					})
			},
			wantIDs: []int{8, 9},
		},
		{
			name: "limit parameter",
			args: map[string]any{
				"state": "done",
				"limit": 2,
			},
			setupMock: func(mockTodos *glabtest.MockTodosServiceInterface) {
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					Return([]*gitlab.Todo{{ID: 10}, {ID: 11}, {ID: 12}}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
			},
			wantIDs: []int{10, 11}, // Expect only 2 due to limit
		},
		{
			name: "pagination",
			args: map[string]any{
				"state": "done",
			},
			setupMock: func(mockTodos *glabtest.MockTodosServiceInterface) {
				// First call returns results with NextPage set
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					Return([]*gitlab.Todo{{ID: 13}, {ID: 14}}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
						NextPage: 2,
					}, nil)

				// Second call returns results without NextPage (end of pagination)
				mockTodos.EXPECT().
					ListTodos(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListTodosOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Todo, *gitlab.Response, error) {
						if opts.Page != 2 {
							t.Errorf("expected Page 2, got %d", opts.Page)
						}

						return []*gitlab.Todo{{ID: 15}}, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: 0,
						}, nil
					})
			},
			wantIDs: []int{13, 14, 15},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)
			tt.setupMock(gitlabClient.MockTodos)

			todoTools := NewTodosTools(gitlabClient.Client)

			// Test direct function call
			var req mcp.CallToolRequest
			req.Params.Arguments = tt.args

			result, err := todoTools.listUserTodos(t.Context(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var todos []*gitlab.Todo
			if err := unmarshalResult(result, &todos); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			var gotIDs []int
			for _, todo := range todos {
				gotIDs = append(gotIDs, todo.ID)
			}

			if diff := cmp.Diff(tt.wantIDs, gotIDs); diff != "" {
				t.Errorf("IDs mismatch (-want/+got):\n%s", diff)
			}
		})
	}
}

func TestCompleteTodoItem(t *testing.T) {
	mockClient := glabtest.NewTestClient(t)

	// Expect MarkTodoAsDone to be called with the todo ID
	mockClient.MockTodos.EXPECT().
		MarkTodoAsDone(gomock.Eq(1), gomock.Any()).
		Return(&gitlab.Response{
			Response: &http.Response{
				StatusCode: http.StatusOK,
			},
		}, nil)

	todoTools := NewTodosTools(mockClient.Client)

	srv, err := mcptest.NewServer(t, todoTools.CompleteTodoItem())
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	client := srv.Client()

	var req mcp.CallToolRequest
	req.Params.Name = "complete_todo_item"
	req.Params.Arguments = map[string]any{
		"id": 1,
	}

	result, err := client.CallTool(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}

	got, err := resultToString(result)
	if err != nil {
		t.Fatal(err)
	}

	if want := "success"; got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestCompleteAllTodoItems(t *testing.T) {
	mockClient := glabtest.NewTestClient(t)

	// Expect MarkAllTodosAsDone to be called
	mockClient.MockTodos.EXPECT().
		MarkAllTodosAsDone(gomock.Any()).
		Return(&gitlab.Response{
			Response: &http.Response{
				StatusCode: http.StatusNoContent,
			},
		}, nil)

	todoTools := NewTodosTools(mockClient.Client)

	srv, err := mcptest.NewServer(t, todoTools.CompleteAllTodoItems())
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	client := srv.Client()

	var req mcp.CallToolRequest
	req.Params.Name = "complete_all_todo_items"
	req.Params.Arguments = map[string]any{}

	result, err := client.CallTool(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}

	got, err := resultToString(result)
	if err != nil {
		t.Fatal(err)
	}

	if want := "success"; got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
