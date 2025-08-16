package tools

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	glabtest "gitlab.com/gitlab-org/api/client-go/testing"
)

func TestListUserEvents(t *testing.T) {
	// Create test dates
	testDate1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	testDate2 := time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC)

	const testUser = "testuser"

	// Helper function to create test events
	createTestEvent := func(id int, actionName string, targetType string, createdAt time.Time) *gitlab.ContributionEvent {
		return &gitlab.ContributionEvent{
			ID:         id,
			ActionName: actionName,
			TargetType: targetType,
			CreatedAt:  &createdAt,
			Author: struct {
				Name      string `json:"name"`
				Username  string `json:"username"`
				ID        int    `json:"id"`
				State     string `json:"state"`
				AvatarURL string `json:"avatar_url"`
				WebURL    string `json:"web_url"`
			}{
				ID:       100,
				Username: testUser,
				Name:     "Test User",
			},
		}
	}

	cmpOpts := []cmp.Option{
		cmp.Comparer(func(i1, i2 gitlab.ISOTime) bool {
			return i1.String() == i2.String()
		}),
	}

	tests := []struct {
		name              string
		args              map[string]any
		setupMock         func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent)
		wantError         bool
		wantErrorResponse bool
		wantEvents        []*gitlab.ContributionEvent
	}{
		{
			name: "list events for current user",
			args: map[string]any{},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				// Mock CurrentUser call
				mockUsers.EXPECT().
					CurrentUser(gomock.Any()).
					Return(&gitlab.User{
						ID:       100,
						Username: testUser,
						Name:     "Test User",
					}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil)

				// Mock ListUserContributionEvents call
				mockUsers.EXPECT().
					ListUserContributionEvents(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(uid any, opts *gitlab.ListContributionEventsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ContributionEvent, *gitlab.Response, error) {
						// Verify username is passed correctly
						if uid != testUser {
							return nil, nil, fmt.Errorf("expected username %q, got %v", testUser, uid)
						}

						// Verify expected options
						wantOpts := &gitlab.ListContributionEventsOptions{
							Sort: gitlab.Ptr("desc"),
							ListOptions: gitlab.ListOptions{
								PerPage: maxPerPage,
							},
						}
						if diff := cmp.Diff(wantOpts, opts, cmpOpts...); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want +got):\n%s", diff)
						}

						return events, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: 0,
						}, nil
					})
			},
			wantEvents: []*gitlab.ContributionEvent{
				createTestEvent(1, "created", "issue", testDate1),
				createTestEvent(2, "commented", "merge_request", testDate2),
			},
		},
		{
			name: "list events for specific user",
			args: map[string]any{
				"username": "specificuser",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				// Should not call CurrentUser when username is provided
				mockUsers.EXPECT().
					ListUserContributionEvents(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(uid any, opts *gitlab.ListContributionEventsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ContributionEvent, *gitlab.Response, error) {
						if uid != "specificuser" {
							return nil, nil, fmt.Errorf("expected username 'specificuser', got %v", uid)
						}

						wantOpts := &gitlab.ListContributionEventsOptions{
							Sort: gitlab.Ptr("desc"),
							ListOptions: gitlab.ListOptions{
								PerPage: maxPerPage,
							},
						}
						if diff := cmp.Diff(wantOpts, opts, cmpOpts...); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want +got):\n%s", diff)
						}

						return events, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: 0,
						}, nil
					})
			},
			wantEvents: []*gitlab.ContributionEvent{
				createTestEvent(3, "updated", "project", testDate1),
			},
		},
		{
			name: "filter by target type",
			args: map[string]any{
				"username":    testUser,
				"target_type": "issue",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				mockUsers.EXPECT().
					ListUserContributionEvents(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(uid any, opts *gitlab.ListContributionEventsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ContributionEvent, *gitlab.Response, error) {
						if uid != testUser {
							return nil, nil, fmt.Errorf("expected username %q, got %v", testUser, uid)
						}

						targetType := gitlab.EventTargetTypeValue("issue")
						wantOpts := &gitlab.ListContributionEventsOptions{
							Sort:       gitlab.Ptr("desc"),
							TargetType: &targetType,
							ListOptions: gitlab.ListOptions{
								PerPage: maxPerPage,
							},
						}
						if diff := cmp.Diff(wantOpts, opts, cmpOpts...); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want +got):\n%s", diff)
						}

						return events, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: 0,
						}, nil
					})
			},
			wantEvents: []*gitlab.ContributionEvent{
				createTestEvent(4, "created", "issue", testDate1),
				createTestEvent(5, "closed", "issue", testDate2),
			},
		},
		{
			name: "filter by action type",
			args: map[string]any{
				"username":    testUser,
				"action_type": "created",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				mockUsers.EXPECT().
					ListUserContributionEvents(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(uid any, opts *gitlab.ListContributionEventsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ContributionEvent, *gitlab.Response, error) {
						if uid != testUser {
							return nil, nil, fmt.Errorf("expected username %q, got %v", testUser, uid)
						}

						eventType := gitlab.EventTypeValue("created")
						wantOpts := &gitlab.ListContributionEventsOptions{
							Sort:   gitlab.Ptr("desc"),
							Action: &eventType,
							ListOptions: gitlab.ListOptions{
								PerPage: maxPerPage,
							},
						}
						if diff := cmp.Diff(wantOpts, opts, cmpOpts...); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want +got):\n%s", diff)
						}

						return events, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: 0,
						}, nil
					})
			},
			wantEvents: []*gitlab.ContributionEvent{
				createTestEvent(6, "created", "merge_request", testDate1),
			},
		},
		{
			name: "filter by date range",
			args: map[string]any{
				"username": testUser,
				"before":   "2024-01-20",
				"after":    "2024-01-01",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				mockUsers.EXPECT().
					ListUserContributionEvents(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(uid any, opts *gitlab.ListContributionEventsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ContributionEvent, *gitlab.Response, error) {
						if uid != testUser {
							return nil, nil, fmt.Errorf("expected username %q, got %v", testUser, uid)
						}

						beforeDate, _ := gitlab.ParseISOTime("2024-01-20")
						afterDate, _ := gitlab.ParseISOTime("2024-01-01")

						wantOpts := &gitlab.ListContributionEventsOptions{
							Sort:   gitlab.Ptr("desc"),
							Before: &beforeDate,
							After:  &afterDate,
							ListOptions: gitlab.ListOptions{
								PerPage: maxPerPage,
							},
						}
						if diff := cmp.Diff(wantOpts, opts, cmpOpts...); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want +got):\n%s", diff)
						}

						return events, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: 0,
						}, nil
					})
			},
			wantEvents: []*gitlab.ContributionEvent{
				createTestEvent(7, "updated", "issue", testDate1),
				createTestEvent(8, "commented", "merge_request", testDate2),
			},
		},
		{
			name: "multiple filters combined",
			args: map[string]any{
				"username":    testUser,
				"target_type": "merge_request",
				"action_type": "created",
				"before":      "2024-01-20",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				mockUsers.EXPECT().
					ListUserContributionEvents(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(uid any, opts *gitlab.ListContributionEventsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ContributionEvent, *gitlab.Response, error) {
						if uid != testUser {
							return nil, nil, fmt.Errorf("expected username %q, got %v", testUser, uid)
						}

						beforeDate, _ := gitlab.ParseISOTime("2024-01-20")
						targetType := gitlab.EventTargetTypeValue("merge_request")
						eventType := gitlab.EventTypeValue("created")

						wantOpts := &gitlab.ListContributionEventsOptions{
							Sort:       gitlab.Ptr("desc"),
							Before:     &beforeDate,
							TargetType: &targetType,
							Action:     &eventType,
							ListOptions: gitlab.ListOptions{
								PerPage: maxPerPage,
							},
						}
						if diff := cmp.Diff(wantOpts, opts, cmpOpts...); diff != "" {
							return nil, nil, fmt.Errorf("options mismatch (-want +got):\n%s", diff)
						}

						return events, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: 0,
						}, nil
					})
			},
			wantEvents: []*gitlab.ContributionEvent{
				createTestEvent(9, "created", "merge_request", testDate1),
			},
		},
		{
			name: "pagination with date boundaries",
			args: map[string]any{
				"username": testUser,
				"before":   "2024-01-20",
				"after":    "2024-01-01",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				// First page
				mockUsers.EXPECT().
					ListUserContributionEvents(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(uid any, opts *gitlab.ListContributionEventsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ContributionEvent, *gitlab.Response, error) {
						if uid != testUser {
							return nil, nil, fmt.Errorf("expected username %q, got %v", testUser, uid)
						}

						if opts.Page == 0 {
							// First page
							return events[:1], &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
								NextPage: 2,
							}, nil
						} else if opts.Page == 2 {
							// Second page
							return events[1:], &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
								NextPage: 0, // No more pages
							}, nil
						}

						return nil, nil, fmt.Errorf("unexpected page: %d", opts.Page)
					}).
					Times(2)
			},
			wantEvents: []*gitlab.ContributionEvent{
				createTestEvent(10, "created", "issue", testDate1),
				createTestEvent(11, "updated", "merge_request", testDate2),
			},
		},
		{
			name: "empty results",
			args: map[string]any{
				"username": "emptyuser",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				mockUsers.EXPECT().
					ListUserContributionEvents(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(uid any, opts *gitlab.ListContributionEventsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ContributionEvent, *gitlab.Response, error) {
						if uid != "emptyuser" {
							return nil, nil, fmt.Errorf("expected username 'emptyuser', got %v", uid)
						}

						return []*gitlab.ContributionEvent{}, &gitlab.Response{
							Response: &http.Response{StatusCode: http.StatusOK},
							NextPage: 0,
						}, nil
					})
			},
			wantEvents: []*gitlab.ContributionEvent{},
		},
		{
			name: "invalid before date",
			args: map[string]any{
				"username": testUser,
				"before":   "invalid-date",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				// No mock setup needed as this should fail validation
			},
			wantError: true,
		},
		{
			name: "invalid after date",
			args: map[string]any{
				"username": testUser,
				"after":    "not-a-date",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				// No mock setup needed as this should fail validation
			},
			wantError: true,
		},
		{
			name: "CurrentUser API error",
			args: map[string]any{}, // No username provided, so it should call CurrentUser
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				mockUsers.EXPECT().
					CurrentUser(gomock.Any()).
					Return(nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusUnauthorized}}, fmt.Errorf("unauthorized"))
			},
			wantError: true,
		},
		{
			name: "ListUserContributionEvents API error",
			args: map[string]any{
				"username": testUser,
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface, events []*gitlab.ContributionEvent) {
				mockUsers.EXPECT().
					ListUserContributionEvents(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(uid any, opts *gitlab.ListContributionEventsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.ContributionEvent, *gitlab.Response, error) {
						if uid != testUser {
							return nil, nil, fmt.Errorf("expected username %q, got %v", testUser, uid)
						}

						return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, fmt.Errorf("server error")
					})
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)

			tt.setupMock(gitlabClient.MockUsers, tt.wantEvents)

			eventsService := &EventsService{
				client: gitlabClient.Client,
			}

			srv := mcptest.NewUnstartedServer(t)
			srv.AddTools(eventsService.ListUserEvents())

			if err := srv.Start(); err != nil {
				t.Fatalf("failed to start server: %v", err)
			}
			defer srv.Close()

			var req mcp.CallToolRequest
			req.Params.Name = "list_user_events"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)

			if gotErr := err != nil; gotErr != tt.wantError {
				t.Errorf("listUserEvents error mismatch, got: %v (%v), want error: %v", err, result, tt.wantError)
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

			var got []*gitlab.ContributionEvent
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v -- result was: %s", err, result.Content)
			}

			if diff := cmp.Diff(tt.wantEvents, got, cmpOpts...); diff != "" {
				t.Errorf("events mismatch (-want/+got):\n%s", diff)
			}
		})
	}
}
