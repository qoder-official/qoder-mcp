package tools

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	glabtest "gitlab.com/gitlab-org/api/client-go/testing"
)

func TestGetMergeRequest(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		setupMock func(*glabtest.MockMergeRequestsServiceInterface, *glabtest.MockDiscussionsServiceInterface)
		want      *detailedMergeRequest
		wantErr   bool
	}{
		{
			name: "successful fetch of merge request with discussions and diffs",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 123,
			},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface, mockDisc *glabtest.MockDiscussionsServiceInterface) {
				// Mock GetMergeRequest
				mockMR.EXPECT().
					GetMergeRequest(gomock.Eq("test/project"), gomock.Eq(123), gomock.Any(), gomock.Any()).
					Return(&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:          12345,
							IID:         123,
							ProjectID:   54321,
							Title:       "Test merge request",
							Description: "Test description",
							State:       "opened",
						},
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)

				// Mock ListMergeRequestDiscussions
				mockDisc.EXPECT().
					ListMergeRequestDiscussions(gomock.Any(), gomock.Eq(123), gomock.Any(), gomock.Any()).
					Return([]*gitlab.Discussion{
						{
							ID: "discussion1",
							Notes: []*gitlab.Note{
								{
									ID:     1,
									Body:   "Test comment",
									Author: gitlab.NoteAuthor{ID: 1, Username: "user1"},
								},
							},
						},
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
						NextPage: 0,
					}, nil)

				// Mock ListMergeRequestDiffs
				mockMR.EXPECT().
					ListMergeRequestDiffs(gomock.Any(), gomock.Eq(123), gomock.Any(), gomock.Any()).
					Return([]*gitlab.MergeRequestDiff{
						{
							OldPath:     "file.go",
							NewPath:     "file.go",
							AMode:       "100644",
							BMode:       "100644",
							NewFile:     false,
							RenamedFile: false,
							DeletedFile: false,
							Diff:        "@@ -1,3 +1,4 @@\n Line 1\n+Line 2\n Line 3\n",
						},
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
						NextPage: 0,
					}, nil)
			},
			want: &detailedMergeRequest{
				MergeRequest: &gitlab.MergeRequest{
					BasicMergeRequest: gitlab.BasicMergeRequest{
						ID:          12345,
						IID:         123,
						ProjectID:   54321,
						Title:       "Test merge request",
						Description: "Test description",
						State:       "opened",
					},
				},
				Discussions: []*gitlab.Discussion{
					{
						ID: "discussion1",
						Notes: []*gitlab.Note{
							{
								ID:     1,
								Body:   "Test comment",
								Author: gitlab.NoteAuthor{ID: 1, Username: "user1"},
							},
						},
					},
				},
				Diffs: []*gitlab.MergeRequestDiff{
					{
						OldPath:     "file.go",
						NewPath:     "file.go",
						AMode:       "100644",
						BMode:       "100644",
						NewFile:     false,
						RenamedFile: false,
						DeletedFile: false,
						Diff:        "@@ -1,3 +1,4 @@\n Line 1\n+Line 2\n Line 3\n",
					},
				},
			},
		},
		{
			name: "error in fetching merge request",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 456,
			},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface, mockDisc *glabtest.MockDiscussionsServiceInterface) {
				// Mock GetMergeRequest with error
				mockMR.EXPECT().
					GetMergeRequest(gomock.Any(), gomock.Eq(456), gomock.Any(), gomock.Any()).
					Return(nil, nil, fmt.Errorf("merge request not found"))

				// These won't be called due to the error, but we need to set expectations
				// for all concurrent fetches to avoid test flakiness
				mockDisc.EXPECT().
					ListMergeRequestDiscussions(gomock.Any(), gomock.Eq(456), gomock.Any(), gomock.Any()).
					Return([]*gitlab.Discussion{}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil).AnyTimes()

				mockMR.EXPECT().
					ListMergeRequestDiffs(gomock.Any(), gomock.Eq(456), gomock.Any(), gomock.Any()).
					Return([]*gitlab.MergeRequestDiff{}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil).AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "error in fetching discussions",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 789,
			},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface, mockDisc *glabtest.MockDiscussionsServiceInterface) {
				// Mock GetMergeRequest
				mockMR.EXPECT().
					GetMergeRequest(gomock.Any(), gomock.Eq(789), gomock.Any(), gomock.Any()).
					Return(&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:    78901,
							IID:   789,
							Title: "Test merge request with discussion error",
						},
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)

				// Mock ListMergeRequestDiscussions with error
				mockDisc.EXPECT().
					ListMergeRequestDiscussions(gomock.Any(), gomock.Eq(789), gomock.Any(), gomock.Any()).
					Return(nil, nil, fmt.Errorf("discussions not available"))

				// Mock ListMergeRequestDiffs
				mockMR.EXPECT().
					ListMergeRequestDiffs(gomock.Any(), gomock.Eq(789), gomock.Any(), gomock.Any()).
					Return([]*gitlab.MergeRequestDiff{}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil).AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "pagination handling for diffs",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 101,
			},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface, mockDisc *glabtest.MockDiscussionsServiceInterface) {
				// Mock GetMergeRequest
				mockMR.EXPECT().
					GetMergeRequest(gomock.Eq("test/project"), gomock.Eq(101), gomock.Any(), gomock.Any()).
					Return(&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:    10101,
							IID:   101,
							Title: "Test merge request with paginated diffs",
						},
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)

				// Mock ListMergeRequestDiscussions
				mockDisc.EXPECT().
					ListMergeRequestDiscussions(gomock.Any(), gomock.Eq(101), gomock.Any(), gomock.Any()).
					Return([]*gitlab.Discussion{}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)

				// First page of diffs
				firstPage := mockMR.EXPECT().
					ListMergeRequestDiffs(gomock.Any(), gomock.Eq(101), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.ListMergeRequestDiffsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.MergeRequestDiff, *gitlab.Response, error) {
						if opts.Page != 0 {
							t.Errorf("expected first page request, got page %d", opts.Page)
						}

						return []*gitlab.MergeRequestDiff{
								{
									OldPath: "file1.go",
									NewPath: "file1.go",
									Diff:    "@@ -1,3 +1,3 @@\n Line 1\n-Line 2\n+Updated Line 2\n",
								},
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
								NextPage: 2,
							}, nil
					})

				// Second page of diffs
				mockMR.EXPECT().
					ListMergeRequestDiffs(gomock.Any(), gomock.Eq(101), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.ListMergeRequestDiffsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.MergeRequestDiff, *gitlab.Response, error) {
						if opts.Page != 2 {
							t.Errorf("expected second page request, got page %d", opts.Page)
						}

						return []*gitlab.MergeRequestDiff{
								{
									OldPath: "file2.go",
									NewPath: "file2.go",
									Diff:    "@@ -10,5 +10,6 @@\n func Example() {\n+    // Add comment\n     fmt.Println(\"test\")\n }\n",
								},
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
								NextPage: 0, // No more pages
							}, nil
					}).
					After(firstPage.Call)
			},
			want: &detailedMergeRequest{
				MergeRequest: &gitlab.MergeRequest{
					BasicMergeRequest: gitlab.BasicMergeRequest{
						ID:    10101,
						IID:   101,
						Title: "Test merge request with paginated diffs",
					},
				},
				Diffs: []*gitlab.MergeRequestDiff{
					{
						OldPath: "file1.go",
						NewPath: "file1.go",
						Diff:    "@@ -1,3 +1,3 @@\n Line 1\n-Line 2\n+Updated Line 2\n",
					},
					{
						OldPath: "file2.go",
						NewPath: "file2.go",
						Diff:    "@@ -10,5 +10,6 @@\n func Example() {\n+    // Add comment\n     fmt.Println(\"test\")\n }\n",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)
			tt.setupMock(gitlabClient.MockMergeRequests, gitlabClient.MockDiscussions)

			mergeRequestsService := NewMergeRequestsTools(gitlabClient.Client, "test_user")

			srv, err := mcptest.NewServer(t, mergeRequestsService.GetMergeRequest())
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}
			defer srv.Close()

			// Test direct function call
			var req mcp.CallToolRequest
			req.Params.Name = "get_merge_request"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)

			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("getIssue error mismatch, got: %v, want error: %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			var got detailedMergeRequest
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if diff := cmp.Diff(tt.want, &got); diff != "" {
				t.Errorf("getIssue() mismatch (-want/+got):\n%s", diff)
			}
		})
	}
}

func TestListUserMergeRequests(t *testing.T) {
	defaultOpts := gitlab.ListMergeRequestsOptions{
		AuthorUsername: gitlab.Ptr("test_user"),
		Scope:          gitlab.Ptr("all"),
		ListOptions: gitlab.ListOptions{
			PerPage: maxPerPage,
		},
	}

	tests := []struct {
		name      string
		args      map[string]any
		setupMock func(*glabtest.MockMergeRequestsServiceInterface)
		want      []*gitlab.BasicMergeRequest
		wantErr   bool
	}{
		{
			name: "successful fetch with default user",
			args: map[string]any{},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListMergeRequests(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						if diff := cmp.Diff(&defaultOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want/+got):\n%s", diff)
						}

						return []*gitlab.BasicMergeRequest{{ID: 1, Title: "MR by test_user"}},
							&gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			want: []*gitlab.BasicMergeRequest{{ID: 1, Title: "MR by test_user"}},
		},
		{
			name: "fetch for a specific user as author",
			args: map[string]any{"username": "another_user", "role": "author"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListMergeRequests(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						wantOpts := defaultOpts
						wantOpts.AuthorUsername = gitlab.Ptr("another_user")
						if diff := cmp.Diff(&wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want/+got):\n%s", diff)
						}

						return []*gitlab.BasicMergeRequest{{ID: 2, Title: "MR by another_user"}},
							&gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			want: []*gitlab.BasicMergeRequest{{ID: 2, Title: "MR by another_user"}},
		},
		{
			name: "fetch for a specific user as reviewer",
			args: map[string]any{"username": "another_user", "role": "reviewer"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListMergeRequests(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						wantOpts := defaultOpts
						wantOpts.AuthorUsername = nil
						wantOpts.ReviewerUsername = gitlab.Ptr("another_user")

						if diff := cmp.Diff(&wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want/+got):\n%s", diff)
						}

						return []*gitlab.BasicMergeRequest{{ID: 3, Title: "MR for another_user to review"}},
							&gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			want: []*gitlab.BasicMergeRequest{{ID: 3, Title: "MR for another_user to review"}},
		},
		{
			name: "fetch with state filter",
			args: map[string]any{"state": "merged"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListMergeRequests(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						wantOpts := defaultOpts
						wantOpts.State = gitlab.Ptr("merged")

						if diff := cmp.Diff(&wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want/+got):\n%s", diff)
						}

						return []*gitlab.BasicMergeRequest{{ID: 4, Title: "Merged MR"}},
							&gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			want: []*gitlab.BasicMergeRequest{{ID: 4, Title: "Merged MR"}},
		},
		{
			name: "pagination",
			args: map[string]any{},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				first := mockMR.EXPECT().
					ListMergeRequests(gomock.Any(), gomock.Any()).
					Return([]*gitlab.BasicMergeRequest{{ID: 1}}, &gitlab.Response{NextPage: 2}, nil)
				mockMR.EXPECT().
					ListMergeRequests(gomock.Any(), gomock.Any()).
					Return([]*gitlab.BasicMergeRequest{{ID: 2}}, &gitlab.Response{NextPage: 0}, nil).
					After(first.Call)
			},
			want: []*gitlab.BasicMergeRequest{{ID: 1}, {ID: 2}},
		},
		{
			name: "fetch with limit flag",
			args: map[string]any{
				"project_id": "test-project",
				// 349 is prime, so it is not a multiple of maxPerPage.
				"limit": 349,
			},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListMergeRequests(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						page := opts.ListOptions.Page
						if page == 0 {
							page = 1
						}

						if opts.ListOptions.PerPage <= 0 {
							return nil, nil, fmt.Errorf("per_page must be greater than 0")
						}

						index := (page - 1) * opts.ListOptions.PerPage
						if index >= 349 {
							return nil, nil, fmt.Errorf("page index %d exceeds limit %d (page = %d, per_page = %d)", index, 349, page, opts.ListOptions.PerPage)
						}

						return makeMergeRequests(index, index+opts.ListOptions.PerPage), &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: page + 1,
						}, nil
					}).AnyTimes()
			},
			want: makeMergeRequests(0, 349),
		},
		{
			name: "API error",
			args: map[string]any{},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListMergeRequests(gomock.Any(), gomock.Any()).
					Return(nil, nil, fmt.Errorf("api error"))
			},
			wantErr: true,
		},
		{
			name:    "invalid role",
			args:    map[string]any{"role": "invalid_role"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)
			if tt.setupMock != nil {
				tt.setupMock(gitlabClient.MockMergeRequests)
			}

			service := NewMergeRequestsTools(gitlabClient.Client, "test_user")

			srv, err := mcptest.NewServer(t, service.ListUserMergeRequests())
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}

			defer srv.Close()

			var req mcp.CallToolRequest
			req.Params.Name = "list_user_merge_requests"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("listUserMergeRequests() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			var got []*gitlab.BasicMergeRequest
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("listUserMergeRequests() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListProjectMergeRequests(t *testing.T) {
	defaultOpts := gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: maxPerPage,
		},
	}

	tests := []struct {
		name      string
		args      map[string]any
		setupMock func(*glabtest.MockMergeRequestsServiceInterface)
		want      []*gitlab.BasicMergeRequest
		wantErr   bool
	}{
		{
			name: "successful fetch with default state",
			args: map[string]any{"project_id": "test/project"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListProjectMergeRequests(gomock.Eq("test/project"), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, opts *gitlab.ListProjectMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						if diff := cmp.Diff(&defaultOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want/+got):\n%s", diff)
						}

						return []*gitlab.BasicMergeRequest{{ID: 1, Title: "Project MR"}},
							&gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			want: []*gitlab.BasicMergeRequest{{ID: 1, Title: "Project MR"}},
		},
		{
			name: "fetch with state filter",
			args: map[string]any{"project_id": "test/project", "state": "closed"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListProjectMergeRequests(gomock.Eq("test/project"), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, opts *gitlab.ListProjectMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						wantOpts := defaultOpts
						wantOpts.State = gitlab.Ptr("closed")

						if diff := cmp.Diff(&wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want/+got):\n%s", diff)
						}

						return []*gitlab.BasicMergeRequest{{ID: 2, Title: "Closed Project MR"}},
							&gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			want: []*gitlab.BasicMergeRequest{{ID: 2, Title: "Closed Project MR"}},
		},
		{
			name: "pagination",
			args: map[string]any{"project_id": "test/project"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				first := mockMR.EXPECT().
					ListProjectMergeRequests(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*gitlab.BasicMergeRequest{{ID: 1}}, &gitlab.Response{NextPage: 2}, nil)
				mockMR.EXPECT().
					ListProjectMergeRequests(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*gitlab.BasicMergeRequest{{ID: 2}}, &gitlab.Response{NextPage: 0}, nil).
					After(first.Call)
			},
			want: []*gitlab.BasicMergeRequest{{ID: 1}, {ID: 2}},
		},
		{
			name: "fetch with limit flag",
			args: map[string]any{
				"project_id": "test/project",
				// 349 is prime, so it is not a multiple of maxPerPage.
				"limit": 349,
			},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListProjectMergeRequests(gomock.Eq("test/project"), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, opts *gitlab.ListProjectMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						page := opts.ListOptions.Page
						if page == 0 {
							page = 1
						}

						if opts.ListOptions.PerPage <= 0 {
							return nil, nil, fmt.Errorf("per_page must be greater than 0")
						}

						index := (page - 1) * opts.ListOptions.PerPage
						if index >= 349 {
							return nil, nil, fmt.Errorf("page index %d exceeds limit %d (page = %d, per_page = %d)", index, 349, page, opts.ListOptions.PerPage)
						}

						return makeMergeRequests(index, index+opts.ListOptions.PerPage), &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: page + 1,
						}, nil
					}).AnyTimes()
			},
			want: makeMergeRequests(0, 349),
		},
		{
			name: "API error",
			args: map[string]any{"project_id": "test/project"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListProjectMergeRequests(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil, fmt.Errorf("api error"))
			},
			wantErr: true,
		},
		{
			name:    "missing project id",
			args:    map[string]any{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)
			if tt.setupMock != nil {
				tt.setupMock(gitlabClient.MockMergeRequests)
			}

			service := NewMergeRequestsTools(gitlabClient.Client, "test_user")

			srv, err := mcptest.NewServer(t, service.ListProjectMergeRequests())
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}

			defer srv.Close()

			var req mcp.CallToolRequest
			req.Params.Name = "list_project_merge_requests"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)

			if (err != nil) != tt.wantErr {
				t.Fatalf("listProjectMergeRequests() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			var got []*gitlab.BasicMergeRequest
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("listProjectMergeRequests() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListGroupMergeRequests(t *testing.T) {
	defaultOpts := gitlab.ListGroupMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: maxPerPage,
		},
	}

	tests := []struct {
		name      string
		args      map[string]any
		setupMock func(*glabtest.MockMergeRequestsServiceInterface)
		want      []*gitlab.BasicMergeRequest
		wantErr   bool
	}{
		{
			name: "successful fetch with default state",
			args: map[string]any{"group_id": "test-group"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListGroupMergeRequests(gomock.Eq("test-group"), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, opts *gitlab.ListGroupMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						if diff := cmp.Diff(&defaultOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want/+got):\n%s", diff)
						}

						return []*gitlab.BasicMergeRequest{{ID: 1, Title: "Group MR"}}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			want: []*gitlab.BasicMergeRequest{{ID: 1, Title: "Group MR"}},
		},
		{
			name: "fetch with state filter",
			args: map[string]any{"group_id": "test-group", "state": "opened"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListGroupMergeRequests(gomock.Eq("test-group"), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, opts *gitlab.ListGroupMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						wantOpts := defaultOpts
						wantOpts.State = gitlab.Ptr("opened")
						if diff := cmp.Diff(&wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want/+got):\n%s", diff)
						}

						return []*gitlab.BasicMergeRequest{{ID: 2, Title: "Opened Group MR"}}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			want: []*gitlab.BasicMergeRequest{{ID: 2, Title: "Opened Group MR"}},
		},
		{
			name: "pagination",
			args: map[string]any{"group_id": "test-group"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				first := mockMR.EXPECT().
					ListGroupMergeRequests(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*gitlab.BasicMergeRequest{{ID: 1}}, &gitlab.Response{NextPage: 2}, nil)
				mockMR.EXPECT().
					ListGroupMergeRequests(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*gitlab.BasicMergeRequest{{ID: 2}}, &gitlab.Response{NextPage: 0}, nil).
					After(first.Call)
			},
			want: []*gitlab.BasicMergeRequest{{ID: 1}, {ID: 2}},
		},
		{
			name: "fetch with limit flag",
			args: map[string]any{
				"group_id": "test-group",
				// 349 is prime, so it is not a multiple of maxPerPage.
				"limit": 349,
			},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListGroupMergeRequests(gomock.Eq("test-group"), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, opts *gitlab.ListGroupMergeRequestsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.BasicMergeRequest, *gitlab.Response, error) {
						page := opts.ListOptions.Page
						if page == 0 {
							page = 1
						}

						if opts.ListOptions.PerPage <= 0 {
							return nil, nil, fmt.Errorf("per_page must be greater than 0")
						}

						index := (page - 1) * opts.ListOptions.PerPage
						if index >= 349 {
							return nil, nil, fmt.Errorf("page index %d exceeds limit %d (page = %d, per_page = %d)", index, 349, page, opts.ListOptions.PerPage)
						}

						return makeMergeRequests(index, index+opts.ListOptions.PerPage), &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: page + 1,
						}, nil
					}).AnyTimes()
			},
			want: makeMergeRequests(0, 349),
		},
		{
			name: "API error",
			args: map[string]any{"group_id": "test-group"},
			setupMock: func(mockMR *glabtest.MockMergeRequestsServiceInterface) {
				mockMR.EXPECT().
					ListGroupMergeRequests(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil, fmt.Errorf("api error"))
			},
			wantErr: true,
		},
		{
			name:    "missing group id",
			args:    map[string]any{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)
			if tt.setupMock != nil {
				tt.setupMock(gitlabClient.MockMergeRequests)
			}

			service := NewMergeRequestsTools(gitlabClient.Client, "test_user")

			srv, err := mcptest.NewServer(t, service.ListGroupMergeRequests())
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}

			defer srv.Close()

			var req mcp.CallToolRequest
			req.Params.Name = "list_group_merge_requests"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("listGroupMergeRequests() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			var got []*gitlab.BasicMergeRequest
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("listGroupMergeRequests() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListDraftNotes(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		setupMock func(*glabtest.MockDraftNotesServiceInterface)
		want      []*gitlab.DraftNote
		wantErr   bool
	}{
		{
			name: "successful fetch of draft notes",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 123,
			},
			setupMock: func(mockDraftNotes *glabtest.MockDraftNotesServiceInterface) {
				mockDraftNotes.EXPECT().
					ListDraftNotes(gomock.Eq("test/project"), gomock.Eq(123), gomock.Any(), gomock.Any()).
					Return([]*gitlab.DraftNote{
						{
							ID:       1,
							AuthorID: 100,
							Note:     "This is a draft comment",
							Position: &gitlab.NotePosition{
								BaseSHA:      "abc123",
								StartSHA:     "def456",
								HeadSHA:      "ghi789",
								OldPath:      "file.go",
								NewPath:      "file.go",
								PositionType: "text+",
								OldLine:      10,
								NewLine:      12,
							},
							LineCode:          "abc123_10_12",
							ResolveDiscussion: false,
						},
						{
							ID:                2,
							AuthorID:          101,
							Note:              "Another draft comment",
							Position:          nil, // General comment without specific position
							ResolveDiscussion: true,
						},
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
						NextPage: 0,
					}, nil)
			},
			want: []*gitlab.DraftNote{
				{
					ID:       1,
					AuthorID: 100,
					Note:     "This is a draft comment",
					Position: &gitlab.NotePosition{
						BaseSHA:      "abc123",
						StartSHA:     "def456",
						HeadSHA:      "ghi789",
						OldPath:      "file.go",
						NewPath:      "file.go",
						PositionType: "text+",
						OldLine:      10,
						NewLine:      12,
					},
					LineCode:          "abc123_10_12",
					ResolveDiscussion: false,
				},
				{
					ID:                2,
					AuthorID:          101,
					Note:              "Another draft comment",
					Position:          nil,
					ResolveDiscussion: true,
				},
			},
		},
		{
			name: "no draft notes found",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 456,
			},
			setupMock: func(mockDraftNotes *glabtest.MockDraftNotesServiceInterface) {
				mockDraftNotes.EXPECT().
					ListDraftNotes(gomock.Eq("test/project"), gomock.Eq(456), gomock.Any(), gomock.Any()).
					Return([]*gitlab.DraftNote{}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
						NextPage: 0,
					}, nil)
			},
			want: []*gitlab.DraftNote{},
		},
		{
			name: "error in fetching draft notes",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 789,
			},
			setupMock: func(mockDraftNotes *glabtest.MockDraftNotesServiceInterface) {
				mockDraftNotes.EXPECT().
					ListDraftNotes(gomock.Eq("test/project"), gomock.Eq(789), gomock.Any(), gomock.Any()).
					Return(nil, nil, fmt.Errorf("failed to fetch draft notes"))
			},
			wantErr: true,
		},
		{
			name: "pagination handling for draft notes",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 101,
			},
			setupMock: func(mockDraftNotes *glabtest.MockDraftNotesServiceInterface) {
				// First page
				firstPage := mockDraftNotes.EXPECT().
					ListDraftNotes(gomock.Eq("test/project"), gomock.Eq(101), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.ListDraftNotesOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.DraftNote, *gitlab.Response, error) {
						if opts.Page != 0 {
							t.Errorf("expected first page request, got page %d", opts.Page)
						}

						return []*gitlab.DraftNote{
								{
									ID:       1,
									AuthorID: 100,
									Note:     "First page draft note",
								},
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
								NextPage: 2,
							}, nil
					})

				// Second page
				mockDraftNotes.EXPECT().
					ListDraftNotes(gomock.Eq("test/project"), gomock.Eq(101), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.ListDraftNotesOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.DraftNote, *gitlab.Response, error) {
						if opts.Page != 2 {
							t.Errorf("expected second page request, got page %d", opts.Page)
						}

						return []*gitlab.DraftNote{
								{
									ID:       2,
									AuthorID: 101,
									Note:     "Second page draft note",
								},
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
								NextPage: 0, // No more pages
							}, nil
					}).
					After(firstPage.Call)
			},
			want: []*gitlab.DraftNote{
				{
					ID:       1,
					AuthorID: 100,
					Note:     "First page draft note",
				},
				{
					ID:       2,
					AuthorID: 101,
					Note:     "Second page draft note",
				},
			},
		},
		{
			name: "missing project ID",
			args: map[string]any{
				// project_id is missing
				"merge_request_iid": 123,
			},
			setupMock: func(mockDraftNotes *glabtest.MockDraftNotesServiceInterface) {
				// No mock expectations needed as this should fail validation
			},
			wantErr: true,
		},
		{
			name: "missing merge request IID",
			args: map[string]any{
				"project_id": "test/project",
				// merge_request_iid is missing
			},
			setupMock: func(mockDraftNotes *glabtest.MockDraftNotesServiceInterface) {
				// No mock expectations needed as this should fail validation
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)
			tt.setupMock(gitlabClient.MockDraftNotes)

			mergeRequestsService := NewMergeRequestsTools(gitlabClient.Client, "test_user")

			srv, err := mcptest.NewServer(t, mergeRequestsService.ListDraftNotes())
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}
			defer srv.Close()

			// Test direct function call
			var req mcp.CallToolRequest
			req.Params.Name = "list_draft_notes"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)

			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("listDraftNotes error mismatch, got: %v, want error: %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			var got []*gitlab.DraftNote
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("listDraftNotes() mismatch (-want/+got):\n%s", diff)
			}
		})
	}
}

func makeMergeRequests(from, to int) []*gitlab.BasicMergeRequest {
	mrs := make([]*gitlab.BasicMergeRequest, 0, to-from)

	for i := from; i < to; i++ {
		mrs = append(mrs, &gitlab.BasicMergeRequest{
			ID:    1000000 + i,
			IID:   i,
			Title: fmt.Sprintf("Merge request %d", i),
			State: "opened",
		})
	}

	return mrs
}
