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

func TestGetUser(t *testing.T) {
	tests := []struct {
		name              string
		args              map[string]any
		setupMock         func(*glabtest.MockUsersServiceInterface)
		wantError         bool
		wantErrorResponse bool
		want              *gitlab.User
	}{
		{
			name: "get current user",
			args: map[string]any{},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					CurrentUser(gomock.Any()).
					Return(&gitlab.User{
						ID:       1,
						Username: "test_user",
						Name:     "Test User",
						Email:    "test@example.com",
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
			},
			want: &gitlab.User{
				ID:       1,
				Username: "test_user",
				Name:     "Test User",
				Email:    "test@example.com",
			},
		},
		{
			name: "get user by ID",
			args: map[string]any{
				"user_id": "2",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					GetUser(gomock.Eq(2), gomock.Any(), gomock.Any()).
					Return(&gitlab.User{
						ID:       2,
						Username: "other_user",
						Name:     "Other User",
						Email:    "other@example.com",
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
			},
			want: &gitlab.User{
				ID:       2,
				Username: "other_user",
				Name:     "Other User",
				Email:    "other@example.com",
			},
		},
		{
			name: "get user by username",
			args: map[string]any{
				"user_id": "search_user",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListUsersOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.User, *gitlab.Response, error) {
						if opts.Username == nil || *opts.Username != "search_user" {
							t.Errorf("opts.Username = %v, want 'search_user'", opts.Username)
						}

						return []*gitlab.User{
								{
									ID:       3,
									Username: "search_user",
									Name:     "Search User",
									Email:    "search@example.com",
								},
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
							}, nil
					})
			},
			want: &gitlab.User{
				ID:       3,
				Username: "search_user",
				Name:     "Search User",
				Email:    "search@example.com",
			},
		},
		{
			name: "get user by username with @ prefix",
			args: map[string]any{
				"user_id": "@prefixed_user",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.ListUsersOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.User, *gitlab.Response, error) {
						if opts.Username == nil || *opts.Username != "prefixed_user" {
							t.Errorf("opts.Username = %v, want 'prefixed_user'", opts.Username)
						}

						return []*gitlab.User{
								{
									ID:       4,
									Username: "prefixed_user",
									Name:     "Prefixed User",
									Email:    "prefixed@example.com",
								},
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
							}, nil
					})
			},
			want: &gitlab.User{
				ID:       4,
				Username: "prefixed_user",
				Name:     "Prefixed User",
				Email:    "prefixed@example.com",
			},
		},
		{
			name: "user not found by username",
			args: map[string]any{
				"user_id": "nonexistent",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					Return([]*gitlab.User{}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
			},
			wantErrorResponse: true,
		},
		{
			name: "error getting user by ID",
			args: map[string]any{
				"user_id": "999",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					GetUser(gomock.Eq(999), gomock.Any(), gomock.Any()).
					Return(nil, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusNotFound},
					}, fmt.Errorf("user not found"))
			},
			wantError: true,
		},
		{
			name: "error getting current user",
			args: map[string]any{},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					CurrentUser(gomock.Any()).
					Return(nil, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusInternalServerError},
					}, fmt.Errorf("API error"))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)
			tt.setupMock(gitlabClient.MockUsers)

			usersService := NewUsersTools(gitlabClient.Client, "test_user")

			srv, err := mcptest.NewServer(t, usersService.GetUser())
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}
			defer srv.Close()

			var req mcp.CallToolRequest
			req.Params.Name = "get_user"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)

			if gotErr := err != nil; gotErr != tt.wantError {
				t.Errorf("getUser error mismatch, got: %v, want error: %v", err, tt.wantError)
			}

			if err != nil {
				return
			}

			if result.IsError != tt.wantErrorResponse {
				t.Errorf("unexpected inline error status, got: %v, want: %v", result.IsError, tt.wantErrorResponse)
			}

			if result.IsError {
				return
			}

			var got *gitlab.User
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("user mismatch (-want/+got):\n%s", diff)
			}
		})
	}
}

func TestGetUserStatus(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		setupMock func(*glabtest.MockUsersServiceInterface)
		wantError bool
		want      *gitlab.UserStatus
	}{
		{
			name: "get user status by user ID",
			args: map[string]any{
				"user_id": "1",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					GetUserStatus(gomock.Eq(1), gomock.Any()).
					Return(&gitlab.UserStatus{
						Emoji:        "speech_balloon",
						Message:      "Working on GitLab MCP",
						MessageHTML:  "Working on GitLab MCP",
						Availability: "not_set",
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
			},
			want: &gitlab.UserStatus{
				Emoji:        "speech_balloon",
				Message:      "Working on GitLab MCP",
				MessageHTML:  "Working on GitLab MCP",
				Availability: "not_set",
			},
		},
		{
			name: "get user status by username",
			args: map[string]any{
				"user_id": "test_user",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					GetUserStatus(gomock.Eq("test_user"), gomock.Any()).
					Return(&gitlab.UserStatus{
						Emoji:        "coffee",
						Message:      "May the code be with you",
						MessageHTML:  "May the code be with you",
						Availability: "not_set",
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
			},
			want: &gitlab.UserStatus{
				Emoji:        "coffee",
				Message:      "May the code be with you",
				MessageHTML:  "May the code be with you",
				Availability: "not_set",
			},
		},
		{
			name: "get user status by username with @ prefix",
			args: map[string]any{
				"user_id": "@prefixed_user",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					GetUserStatus(gomock.Eq("prefixed_user"), gomock.Any()).
					Return(&gitlab.UserStatus{
						Emoji:        "pizza",
						Message:      "I'll be back... after lunch",
						MessageHTML:  "I'll be back... after lunch",
						Availability: "not_set",
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
			},
			want: &gitlab.UserStatus{
				Emoji:        "pizza",
				Message:      "I'll be back... after lunch",
				MessageHTML:  "I'll be back... after lunch",
				Availability: "not_set",
			},
		},
		{
			name: "get busy user status",
			args: map[string]any{
				"user_id": "2",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					GetUserStatus(gomock.Eq(2), gomock.Any()).
					Return(&gitlab.UserStatus{
						Emoji:        "calendar",
						Message:      "In a meeting",
						MessageHTML:  "In a meeting",
						Availability: "busy",
					}, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
			},
			want: &gitlab.UserStatus{
				Emoji:        "calendar",
				Message:      "In a meeting",
				MessageHTML:  "In a meeting",
				Availability: "busy",
			},
		},
		{
			name: "user status not found",
			args: map[string]any{
				"user_id": "999",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					GetUserStatus(gomock.Eq(999), gomock.Any()).
					Return(nil, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusNotFound},
					}, fmt.Errorf("user status not found"))
			},
			wantError: true,
		},
		{
			name: "missing required user_id",
			args: map[string]any{},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				// No mock expectations - should fail at validation
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)
			tt.setupMock(gitlabClient.MockUsers)

			usersService := NewUsersTools(gitlabClient.Client, "test_user")

			srv, err := mcptest.NewServer(t, usersService.GetUserStatus())
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}
			defer srv.Close()

			var req mcp.CallToolRequest
			req.Params.Name = "get_user_status"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)

			if gotErr := err != nil; gotErr != tt.wantError {
				t.Errorf("getUserStatus error mismatch, got: %v, want error: %v", err, tt.wantError)
			}

			if err != nil {
				return
			}

			var got *gitlab.UserStatus
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("user status mismatch (-want/+got):\n%s", diff)
			}
		})
	}
}

func TestSetUserStatus(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		setupMock func(*glabtest.MockUsersServiceInterface)
		wantError bool
		want      *gitlab.UserStatus
	}{
		{
			name: "set user status with all parameters",
			args: map[string]any{
				"emoji":        "coffee",
				"message":      "Taking a break",
				"availability": "busy",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					SetUserStatus(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.UserStatusOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.UserStatus, *gitlab.Response, error) {
						if opts.Emoji == nil || *opts.Emoji != "coffee" {
							t.Errorf("opts.Emoji = %v, want 'coffee'", opts.Emoji)
						}
						if opts.Message == nil || *opts.Message != "Taking a break" {
							t.Errorf("opts.Message = %v, want 'Taking a break'", opts.Message)
						}
						if opts.Availability == nil || *opts.Availability != "busy" {
							t.Errorf("opts.Availability = %v, want 'busy'", opts.Availability)
						}

						return &gitlab.UserStatus{
								Emoji:        "coffee",
								Message:      "Taking a break",
								MessageHTML:  "Taking a break",
								Availability: "busy",
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
							}, nil
					})
			},
			want: &gitlab.UserStatus{
				Emoji:        "coffee",
				Message:      "Taking a break",
				MessageHTML:  "Taking a break",
				Availability: "busy",
			},
		},
		{
			name: "set user status with message only",
			args: map[string]any{
				"message": "Working remotely",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					SetUserStatus(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.UserStatusOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.UserStatus, *gitlab.Response, error) {
						if opts.Emoji != nil {
							t.Errorf("opts.Emoji = %v, want nil", opts.Emoji)
						}
						if opts.Message == nil || *opts.Message != "Working remotely" {
							t.Errorf("opts.Message = %v, want 'Working remotely'", opts.Message)
						}
						if opts.Availability != nil {
							t.Errorf("opts.Availability = %v, want nil", opts.Availability)
						}

						return &gitlab.UserStatus{
								Emoji:        "speech_balloon", // Default emoji
								Message:      "Working remotely",
								MessageHTML:  "Working remotely",
								Availability: "not_set", // Default availability
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
							}, nil
					})
			},
			want: &gitlab.UserStatus{
				Emoji:        "speech_balloon",
				Message:      "Working remotely",
				MessageHTML:  "Working remotely",
				Availability: "not_set",
			},
		},
		{
			name: "set user status with availability only",
			args: map[string]any{
				"availability": "busy",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					SetUserStatus(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.UserStatusOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.UserStatus, *gitlab.Response, error) {
						if opts.Emoji != nil {
							t.Errorf("opts.Emoji = %v, want nil", opts.Emoji)
						}
						if opts.Message != nil {
							t.Errorf("opts.Message = %v, want nil", opts.Message)
						}
						if opts.Availability == nil || *opts.Availability != "busy" {
							t.Errorf("opts.Availability = %v, want 'busy'", opts.Availability)
						}

						return &gitlab.UserStatus{
								Emoji:        "",
								Message:      "",
								MessageHTML:  "",
								Availability: "busy",
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
							}, nil
					})
			},
			want: &gitlab.UserStatus{
				Emoji:        "",
				Message:      "",
				MessageHTML:  "",
				Availability: "busy",
			},
		},
		{
			name: "clear user status",
			args: map[string]any{
				"emoji":        "",
				"message":      "",
				"availability": "not_set",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					SetUserStatus(gomock.Any(), gomock.Any()).
					DoAndReturn(func(opts *gitlab.UserStatusOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.UserStatus, *gitlab.Response, error) {
						// Empty emoji and message are still sent as empty strings, not nil
						if opts.Emoji != nil && *opts.Emoji != "" {
							t.Errorf("opts.Emoji = %v, want empty string", opts.Emoji)
						}
						if opts.Message != nil && *opts.Message != "" {
							t.Errorf("opts.Message = %v, want empty string", opts.Message)
						}
						if opts.Availability == nil || *opts.Availability != "not_set" {
							t.Errorf("opts.Availability = %v, want 'not_set'", opts.Availability)
						}

						return &gitlab.UserStatus{
								Emoji:        "",
								Message:      "",
								MessageHTML:  "",
								Availability: "not_set",
							}, &gitlab.Response{
								Response: &http.Response{StatusCode: http.StatusOK},
							}, nil
					})
			},
			want: &gitlab.UserStatus{
				Emoji:        "",
				Message:      "",
				MessageHTML:  "",
				Availability: "not_set",
			},
		},
		{
			name: "error setting user status",
			args: map[string]any{
				"message": "Error test",
			},
			setupMock: func(mockUsers *glabtest.MockUsersServiceInterface) {
				mockUsers.EXPECT().
					SetUserStatus(gomock.Any(), gomock.Any()).
					Return(nil, &gitlab.Response{
						Response: &http.Response{StatusCode: http.StatusInternalServerError},
					}, fmt.Errorf("API error"))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitlabClient := glabtest.NewTestClient(t)
			tt.setupMock(gitlabClient.MockUsers)

			usersService := NewUsersTools(gitlabClient.Client, "test_user")

			srv, err := mcptest.NewServer(t, usersService.SetUserStatus())
			if err != nil {
				t.Fatalf("failed to create server: %v", err)
			}
			defer srv.Close()

			var req mcp.CallToolRequest
			req.Params.Name = "set_user_status"
			req.Params.Arguments = tt.args

			result, err := srv.Client().CallTool(t.Context(), req)

			if gotErr := err != nil; gotErr != tt.wantError {
				t.Errorf("setUserStatus error mismatch, got: %v, want error: %v", err, tt.wantError)
			}

			if err != nil {
				return
			}

			var got *gitlab.UserStatus
			if err := unmarshalResult(result, &got); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("user status mismatch (-want/+got):\n%s", diff)
			}
		})
	}
}
