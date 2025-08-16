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

func TestEditMergeRequest(t *testing.T) {
	tests := []struct {
		name              string
		args              map[string]any
		setupMock         func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest)
		wantError         bool
		wantErrorResponse bool // If the tool call itself should return an error message
		wantResult        *gitlab.MergeRequest
	}{
		{
			name: "update title",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 1,
				"title":             "New MR Title",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().
					UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(1), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							Title: gitlab.Ptr("New MR Title"),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					}).
					Return(ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:   1,
					Title: "New MR Title",
				},
			},
		},
		{
			name: "update description",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 2,
				"description":       "New MR Description",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(2), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							Description: gitlab.Ptr("New MR Description"),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:         2,
					Description: "New MR Description",
				},
			},
		},
		{
			name: "update target branch",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 3,
				"target_branch":     "develop",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(3), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							TargetBranch: gitlab.Ptr("develop"),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:          3,
					TargetBranch: "develop",
				},
			},
		},
		{
			name: "clear assignees",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 4,
				"assignee_ids":      "-", // Signal to clear assignees
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(4), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							AssigneeIDs: &[]int{},
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:       4,
					Assignees: []*gitlab.BasicUser{},
				},
			},
		},
		{
			name: "set multiple assignees",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 5,
				"assignee_ids":      "101,102",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(5), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							AssigneeIDs: &[]int{101, 102},
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:       5,
					Assignees: []*gitlab.BasicUser{{ID: 101}, {ID: 102}},
				},
			},
		},
		{
			name: "set multiple reviewers",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 6,
				"reviewer_ids":      "201,202",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(6), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							ReviewerIDs: &[]int{201, 202},
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:       6,
					Reviewers: []*gitlab.BasicUser{{ID: 201}, {ID: 202}},
				},
			},
		},
		{
			name: "clear reviewers",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 7,
				"reviewer_ids":      "-",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(7), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							ReviewerIDs: &[]int{},
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:       7,
					Reviewers: []*gitlab.BasicUser{},
				},
			},
		},
		{
			name: "add labels",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 8,
				"add_labels":        "feature,ready-for-review",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(8), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							AddLabels: &gitlab.LabelOptions{"feature", "ready-for-review"},
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:    8,
					Labels: []string{"feature", "ready-for-review"},
				},
			},
		},
		{
			name: "remove labels",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 9,
				"remove_labels":     "draft",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(9), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							RemoveLabels: &gitlab.LabelOptions{"draft"},
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:    9,
					Labels: []string{"feature"}, // Assuming 'draft' is removed
				},
			},
		},
		{
			name: "set milestone",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 10,
				"milestone_id":      42,
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(10), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							MilestoneID: gitlab.Ptr(42),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:       10,
					Milestone: &gitlab.Milestone{ID: 42},
				},
			},
		},
		{
			name: "close merge request",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 11,
				"state_event":       "close",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(11), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							StateEvent: gitlab.Ptr("close"),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:   11,
					State: "closed",
				},
			},
		},
		{
			name: "reopen merge request",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 12,
				"state_event":       "reopen",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(12), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							StateEvent: gitlab.Ptr("reopen"),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:   12,
					State: "opened",
				},
			},
		},
		{
			name: "enable remove source branch",
			args: map[string]any{
				"project_id":           "test/project",
				"merge_request_iid":    13,
				"remove_source_branch": true,
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(13), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							RemoveSourceBranch: gitlab.Ptr(true),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:                     13,
					ForceRemoveSourceBranch: true,
				},
			},
		},
		{
			name: "disable squash",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 14,
				"squash":            false,
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(14), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							Squash: gitlab.Ptr(false),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:    14,
					Squash: false,
				},
			},
		},
		{
			name: "lock discussion",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 15,
				"discussion_locked": true,
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(15), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							DiscussionLocked: gitlab.Ptr(true),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:              15,
					DiscussionLocked: true,
				},
			},
		},
		{
			name: "enable allow collaboration",
			args: map[string]any{
				"project_id":          "test/project",
				"merge_request_iid":   16,
				"allow_collaboration": true,
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(16), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							AllowCollaboration: gitlab.Ptr(true),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:                16,
					AllowCollaboration: true,
				},
			},
		},
		{
			name: "multiple fields update",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 17,
				"title":             "Updated Title",
				"description":       "Updated Description",
				"add_labels":        "enhancement,high-priority",
				"assignee_ids":      "301,302",
				"squash":            true,
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(17), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						wantOpts := &gitlab.UpdateMergeRequestOptions{
							Title:       gitlab.Ptr("Updated Title"),
							Description: gitlab.Ptr("Updated Description"),
							AddLabels:   &gitlab.LabelOptions{"enhancement", "high-priority"},
							AssigneeIDs: &[]int{301, 302},
							Squash:      gitlab.Ptr(true),
						}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return ret, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					})
			},
			wantResult: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					IID:         17,
					Title:       "Updated Title",
					Description: "Updated Description",
					Labels:      []string{"enhancement", "high-priority"},
					Assignees:   []*gitlab.BasicUser{{ID: 301}, {ID: 302}},
					Squash:      true,
				},
			},
		},
		{
			name: "invalid state event",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 18,
				"state_event":       "invalid",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				// No mock setup needed as this should fail validation
			},
			wantError: true,
		},
		{
			name: "UpdateMergeRequest returns error",
			args: map[string]any{
				"project_id":        "test/project",
				"merge_request_iid": 20,
				"title":             "New Title",
			},
			setupMock: func(mockMergeRequests *glabtest.MockMergeRequestsServiceInterface, ret *gitlab.MergeRequest) {
				mockMergeRequests.EXPECT().
					UpdateMergeRequest(gomock.Eq("test/project"), gomock.Eq(20), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ int, opts *gitlab.UpdateMergeRequestOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
						// Check that opts are what we expect even when an error is returned
						wantOpts := &gitlab.UpdateMergeRequestOptions{Title: gitlab.Ptr("New Title")}
						if diff := cmp.Diff(wantOpts, opts); diff != "" {
							return nil, nil, fmt.Errorf("UpdateMergeRequest() options mismatch (-want +got):\n%s", diff)
						}

						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, fmt.Errorf("update failed")
					})
			},
			wantError:  true,
			wantResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)

			tt.setupMock(gitlabClient.MockMergeRequests, tt.wantResult)

			mergeRequestsService := NewMergeRequestsTools(gitlabClient.Client, "test_user")

			srv, err := mcptest.NewServer(t, mergeRequestsService.EditMergeRequest())
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}
			defer srv.Close()

			var req mcp.CallToolRequest
			req.Params.Name = "edit_merge_request"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)

			if gotErr := err != nil; gotErr != tt.wantError {
				t.Errorf("editMergeRequest error mismatch, got: %v (%v), want error: %v", err, result, tt.wantError)
			}

			if err != nil {
				return
			}

			if result.IsError != tt.wantErrorResponse {
				t.Errorf("unexpected inline error status, got: %v (%s), want: %v", result.IsError, result.Content, tt.wantErrorResponse)
			}

			if result.IsError || tt.wantError {
				return
			}

			var got gitlab.MergeRequest
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v -- result was: %s", err, result.Content)
			}

			if diff := cmp.Diff(tt.wantResult, &got); diff != "" {
				t.Errorf("merge request mismatch (-want/+got):\n%s", diff)
			}
		})
	}
}
