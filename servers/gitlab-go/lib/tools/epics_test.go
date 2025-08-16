package tools

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/testing/gitlabtest"
)

func fakeGitlabClient() *gitlab.Client {
	epics := []*gitlab.Epic{
		{
			ID:      1001,
			IID:     1,
			GroupID: 10,
			State:   "opened",
		},
		{
			ID:       1002,
			IID:      2,
			GroupID:  10,
			ParentID: 1001, // parent in same group
			State:    "opened",
		},
		{
			ID:       1003,
			IID:      3,
			GroupID:  10,
			ParentID: 1004, // parent in other group
			State:    "closed",
		},
		{
			ID:      1004,
			IID:     4,
			GroupID: 20,
			State:   "opened",
		},
		{
			ID:       1005,
			IID:      5,
			GroupID:  20,
			ParentID: 1001, // parent in same group
			State:    "opened",
		},
		{
			ID:       1006,
			IID:      6,
			GroupID:  20,
			ParentID: 1002, // parent in other group
			State:    "opened",
		},
	}

	issues := []*gitlab.Issue{
		{
			ID: 101,
			Epic: &gitlab.Epic{
				ID: 1001,
			},
		},
		{
			ID: 102,
			Epic: &gitlab.Epic{
				ID: 1001,
			},
		},
	}

	return &gitlab.Client{
		Epics: &gitlabtest.FakeEpicsService{
			Epics: epics,
		},
		EpicIssues: &gitlabtest.FakeEpicIssuesService{
			Epics:  epics,
			Issues: issues,
		},
	}
}

func TestListGroupEpics(t *testing.T) {
	gitlabClient := fakeGitlabClient()

	epicTools := NewEpicTools(gitlabClient)

	srv, err := mcptest.NewServer(t, epicTools.ListGroupEpics())
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	client := srv.Client()

	var req mcp.CallToolRequest

	req.Params.Name = "list_group_epics"
	req.Params.Arguments = map[string]any{
		"group_id": "10",
		"state":    "opened",
	}

	result, err := client.CallTool(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}

	var epics []*gitlab.Epic
	if err := unmarshalResult(result, &epics); err != nil {
		t.Fatal(err)
	}

	var (
		wantIDs = []int{1001, 1002}
		gotIDs  = epicIDs(epics)
	)

	if !cmp.Equal(gotIDs, wantIDs) {
		t.Errorf("got epics with ids %v, wanted %v", gotIDs, wantIDs)
	}
}

func TestGetEpic(t *testing.T) {
	gitlabClient := fakeGitlabClient()

	epicTools := NewEpicTools(gitlabClient)

	srv, err := mcptest.NewServer(t, epicTools.GetEpic())
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	client := srv.Client()

	var req mcp.CallToolRequest

	req.Params.Name = "get_epic"
	req.Params.Arguments = map[string]any{
		"group_id": "10",
		"epic_iid": 1,
	}

	result, err := client.CallTool(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}

	var epic gitlab.Epic
	if err := unmarshalResult(result, &epic); err != nil {
		t.Fatal(err)
	}

	if wantID := 1001; epic.ID != wantID {
		t.Errorf("got epic id %v, wanted %v", epic.ID, wantID)
	}
}

func TestGetEpicLinks(t *testing.T) {
	gitlabClient := fakeGitlabClient()

	epicTools := NewEpicTools(gitlabClient)

	srv, err := mcptest.NewServer(t, epicTools.GetEpicLinks())
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	client := srv.Client()

	var req mcp.CallToolRequest

	req.Params.Name = "get_epic_links"
	req.Params.Arguments = map[string]any{
		"group_id": "10",
		"epic_iid": 1,
	}

	result, err := client.CallTool(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}

	var epics []*gitlab.Epic
	if err := unmarshalResult(result, &epics); err != nil {
		t.Fatal(err)
	}

	var (
		wantIDs = []int{1002, 1005}
		gotIDs  = epicIDs(epics)
	)

	if !cmp.Equal(gotIDs, wantIDs) {
		t.Errorf("got epics with ids %v, wanted %v", gotIDs, wantIDs)
	}
}

func TestListEpicIssues(t *testing.T) {
	gitlabClient := fakeGitlabClient()

	epicTools := NewEpicTools(gitlabClient)

	srv, err := mcptest.NewServer(t, epicTools.ListEpicIssues())
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	client := srv.Client()

	var req mcp.CallToolRequest

	req.Params.Name = "list_epic_issues"
	req.Params.Arguments = map[string]any{
		"group_id": "10",
		"epic_iid": 1,
	}

	result, err := client.CallTool(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}

	var issues []*gitlab.Issue
	if err := unmarshalResult(result, &issues); err != nil {
		t.Fatal(err)
	}

	var gotIDs []int
	for _, issue := range issues {
		gotIDs = append(gotIDs, issue.ID)
	}

	sort.Ints(gotIDs)

	wantIDs := []int{101, 102}

	if !cmp.Equal(gotIDs, wantIDs) {
		t.Errorf("got issues with ids %v, wanted %v", gotIDs, wantIDs)
	}
}

func epicIDs(epics []*gitlab.Epic) []int {
	var ids []int
	for _, epic := range epics {
		ids = append(ids, epic.ID)
	}

	sort.Ints(ids)

	return ids
}

func unmarshalResult(res *mcp.CallToolResult, v any) error {
	s, err := resultToString(res)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(s), v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func resultToString(res *mcp.CallToolResult) (string, error) {
	var b strings.Builder

	for _, c := range res.Content {
		tc, ok := mcp.AsTextContent(c)
		if !ok {
			return "", fmt.Errorf("content is not text: %T", c)
		}

		b.WriteString(tc.Text)
	}

	if res.IsError {
		return "", fmt.Errorf("tool returned error: %s", b.String())
	}

	return b.String(), nil
}
