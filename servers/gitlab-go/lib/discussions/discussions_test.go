//nolint:err113
package discussions_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/discussions"
	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

func TestCommitDiscussion(t *testing.T) {
	client := newFakeClient(map[string][]*gitlab.Discussion{})

	mgr, err := discussions.NewCommitDiscussion(client, mcpargs.ID{Integer: 1234}, "deadbeef")
	if err != nil {
		t.Fatal(err)
	}

	testManager(t, mgr)
}

func TestEpicDiscussion(t *testing.T) {
	client := newFakeClient(map[string][]*gitlab.Discussion{})

	mgr, err := discussions.NewEpicDiscussion(client, mcpargs.ID{Integer: 1234}, 42)
	if err != nil {
		t.Fatal(err)
	}

	testManager(t, mgr)
}

func TestIssueDiscussion(t *testing.T) {
	client := newFakeClient(map[string][]*gitlab.Discussion{})

	mgr, err := discussions.NewEpicDiscussion(client, mcpargs.ID{Integer: 1234}, 42)
	if err != nil {
		t.Fatal(err)
	}

	testManager(t, mgr)
}

func TestMergeRequestDiscussion(t *testing.T) {
	client := newFakeClient(map[string][]*gitlab.Discussion{})

	mgr, err := discussions.NewMergeRequestDiscussion(client, mcpargs.ID{Integer: 1234}, 42)
	if err != nil {
		t.Fatal(err)
	}

	testManager(t, mgr)
}

func TestSnippetDiscussion(t *testing.T) {
	client := newFakeClient(map[string][]*gitlab.Discussion{})

	mgr, err := discussions.NewSnippetDiscussion(client, mcpargs.ID{Integer: 1234}, 42)
	if err != nil {
		t.Fatal(err)
	}

	testManager(t, mgr)
}

//nolint:funlen
func testManager(t *testing.T, mgr discussions.Manager) {
	t.Helper()

	// NewDiscussion
	discussion, err := mgr.NewDiscussion(t.Context(), "New comment")
	if err != nil {
		t.Fatal(err)
	}

	if want := "0"; discussion.ID != want {
		t.Fatalf("discussion.ID = %q, want %q", discussion.ID, want)
	}

	want := []*gitlab.Discussion{
		{
			ID: "0",
			Notes: []*gitlab.Note{
				{
					ID:   0,
					Body: "New comment",
				},
			},
		},
	}

	// List
	got, err := mgr.List(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(gitlab.Discussion{})); diff != "" {
		t.Errorf("List issue discussion differs (-want/+got):\n%s", diff)
	}

	// AddNote
	note, err := mgr.AddNote(t.Context(), "0", "First reply")
	if err != nil {
		t.Fatal(err)
	}

	if want := 1; note.ID != want {
		t.Fatalf("note.ID = %d, want %d", note.ID, want)
	}

	want[0].Notes = append(want[0].Notes, &gitlab.Note{
		ID:   1,
		Body: "First reply",
	})

	checkDiscussion(t, mgr, want)

	// ModifyNote
	note, err = mgr.ModifyNote(t.Context(), discussion.ID, 1, "Updated reply")
	if err != nil {
		t.Fatal(err)
	}

	if want := 1; note.ID != want {
		t.Fatalf("note.ID = %d, want %d", note.ID, want)
	}

	want[0].Notes[1] = &gitlab.Note{
		ID:   1,
		Body: "Updated reply",
	}

	checkDiscussion(t, mgr, want)

	// DeleteNote
	err = mgr.DeleteNote(t.Context(), discussion.ID, 1)
	if err != nil {
		t.Fatal(err)
	}

	want[0].Notes = want[0].Notes[:1]

	checkDiscussion(t, mgr, want)
}

func checkDiscussion(t *testing.T, mgr discussions.Manager, want []*gitlab.Discussion) {
	t.Helper()

	// Get current discussions from manager
	got, err := mgr.List(t.Context(), false)
	if err != nil {
		t.Fatalf("List discussions failed: %v", err)
	}

	// Compare discussions using cmp.Diff
	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(gitlab.Discussion{})); diff != "" {
		t.Errorf("List discussions differs (-want/+got):\n%s", diff)
	}
}

var _ gitlab.DiscussionsServiceInterface = (*fakeDiscussionsService)(nil)

type fakeDiscussionsService struct {
	gitlab.DiscussionsServiceInterface

	nextID      int
	discussions map[string][]*gitlab.Discussion
}

func newFakeClient(discussions map[string][]*gitlab.Discussion) *gitlab.Client {
	return &gitlab.Client{
		Discussions: newFakeDiscussionsService(discussions),
	}
}

func newFakeDiscussionsService(discussions map[string][]*gitlab.Discussion) *fakeDiscussionsService {
	return &fakeDiscussionsService{
		discussions: discussions,
	}
}

// Issues

func (fds *fakeDiscussionsService) ListIssueDiscussions(projectID any, issueIID int, opt *gitlab.ListIssueDiscussionsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Discussion, *gitlab.Response, error) {
	return fds.list("Issue", projectID, issueIID)
}

func (fds *fakeDiscussionsService) CreateIssueDiscussion(projectID any, issueIID int, opt *gitlab.CreateIssueDiscussionOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Discussion, *gitlab.Response, error) {
	return fds.create("Issue", projectID, issueIID, *opt.Body)
}

func (fds *fakeDiscussionsService) AddIssueDiscussionNote(projectID any, issueIID int, discussionID string, opt *gitlab.AddIssueDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.addNote("Issue", projectID, issueIID, discussionID, *opt.Body)
}

func (fds *fakeDiscussionsService) UpdateIssueDiscussionNote(projectID any, issueIID int, discussionID string, noteID int, opt *gitlab.UpdateIssueDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.updateNote("Issue", projectID, issueIID, discussionID, noteID, *opt.Body)
}

func (fds *fakeDiscussionsService) DeleteIssueDiscussionNote(projectID any, issueIID int, discussionID string, noteID int, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return fds.deleteNote("Issue", projectID, issueIID, discussionID, noteID)
}

// Snippets

func (fds *fakeDiscussionsService) ListSnippetDiscussions(projectID any, snippetID int, opt *gitlab.ListSnippetDiscussionsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Discussion, *gitlab.Response, error) {
	return fds.list("Snippet", projectID, snippetID)
}

func (fds *fakeDiscussionsService) CreateSnippetDiscussion(projectID any, snippetID int, opt *gitlab.CreateSnippetDiscussionOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Discussion, *gitlab.Response, error) {
	return fds.create("Snippet", projectID, snippetID, *opt.Body)
}

func (fds *fakeDiscussionsService) AddSnippetDiscussionNote(projectID any, snippetID int, discussionID string, opt *gitlab.AddSnippetDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.addNote("Snippet", projectID, snippetID, discussionID, *opt.Body)
}

func (fds *fakeDiscussionsService) UpdateSnippetDiscussionNote(projectID any, snippetID int, discussionID string, noteID int, opt *gitlab.UpdateSnippetDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.updateNote("Snippet", projectID, snippetID, discussionID, noteID, *opt.Body)
}

func (fds *fakeDiscussionsService) DeleteSnippetDiscussionNote(projectID any, snippetID int, discussionID string, noteID int, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return fds.deleteNote("Snippet", projectID, snippetID, discussionID, noteID)
}

// Merge Requests

func (fds *fakeDiscussionsService) ListMergeRequestDiscussions(projectID any, mergeRequestIID int, opt *gitlab.ListMergeRequestDiscussionsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Discussion, *gitlab.Response, error) {
	return fds.list("MergeRequest", projectID, mergeRequestIID)
}

func (fds *fakeDiscussionsService) CreateMergeRequestDiscussion(projectID any, mergeRequestIID int, opt *gitlab.CreateMergeRequestDiscussionOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Discussion, *gitlab.Response, error) {
	return fds.create("MergeRequest", projectID, mergeRequestIID, *opt.Body)
}

func (fds *fakeDiscussionsService) AddMergeRequestDiscussionNote(projectID any, mergeRequestIID int, discussionID string, opt *gitlab.AddMergeRequestDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.addNote("MergeRequest", projectID, mergeRequestIID, discussionID, *opt.Body)
}

func (fds *fakeDiscussionsService) UpdateMergeRequestDiscussionNote(projectID any, mergeRequestIID int, discussionID string, noteID int, opt *gitlab.UpdateMergeRequestDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.updateNote("MergeRequest", projectID, mergeRequestIID, discussionID, noteID, *opt.Body)
}

func (fds *fakeDiscussionsService) DeleteMergeRequestDiscussionNote(projectID any, mergeRequestIID int, discussionID string, noteID int, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return fds.deleteNote("MergeRequest", projectID, mergeRequestIID, discussionID, noteID)
}

func (fds *fakeDiscussionsService) ResolveMergeRequestDiscussion(projectID any, mergeRequestIID int, discussionID string, opt *gitlab.ResolveMergeRequestDiscussionOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Discussion, *gitlab.Response, error) {
	key := fmt.Sprintf("MergeRequest(%v,%d)", projectID, mergeRequestIID)

	discussion, ok := findDiscussion(fds.discussions[key], discussionID)
	if !ok {
		return nil, nil, fmt.Errorf("ResolveMergeRequestDiscussion(%v, %d, %s): discussion ID does not exist", projectID, mergeRequestIID, discussionID)
	}

	discussion.Notes[0].Resolved = *opt.Resolved

	return discussion, &gitlab.Response{}, nil
}

// Epics

func (fds *fakeDiscussionsService) ListGroupEpicDiscussions(groupID any, epicIID int, opt *gitlab.ListGroupEpicDiscussionsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Discussion, *gitlab.Response, error) {
	return fds.list("Epic", groupID, epicIID)
}

func (fds *fakeDiscussionsService) CreateEpicDiscussion(groupID any, epicIID int, opt *gitlab.CreateEpicDiscussionOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Discussion, *gitlab.Response, error) {
	return fds.create("Epic", groupID, epicIID, *opt.Body)
}

func (fds *fakeDiscussionsService) AddEpicDiscussionNote(groupID any, epicIID int, discussionID string, opt *gitlab.AddEpicDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.addNote("Epic", groupID, epicIID, discussionID, *opt.Body)
}

func (fds *fakeDiscussionsService) UpdateEpicDiscussionNote(groupID any, epicIID int, discussionID string, noteID int, opt *gitlab.UpdateEpicDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.updateNote("Epic", groupID, epicIID, discussionID, noteID, *opt.Body)
}

func (fds *fakeDiscussionsService) DeleteEpicDiscussionNote(groupID any, epicIID int, discussionID string, noteID int, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return fds.deleteNote("Epic", groupID, epicIID, discussionID, noteID)
}

// Commits

func (fds *fakeDiscussionsService) ListCommitDiscussions(projectID any, commitSHA string, opt *gitlab.ListCommitDiscussionsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Discussion, *gitlab.Response, error) {
	return fds.list("Commit", projectID, commitSHA)
}

func (fds *fakeDiscussionsService) CreateCommitDiscussion(projectID any, commitSHA string, opt *gitlab.CreateCommitDiscussionOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Discussion, *gitlab.Response, error) {
	return fds.create("Commit", projectID, commitSHA, *opt.Body)
}

func (fds *fakeDiscussionsService) AddCommitDiscussionNote(projectID any, commitSHA string, discussionID string, opt *gitlab.AddCommitDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.addNote("Commit", projectID, commitSHA, discussionID, *opt.Body)
}

func (fds *fakeDiscussionsService) UpdateCommitDiscussionNote(projectID any, commitSHA string, discussionID string, noteID int, opt *gitlab.UpdateCommitDiscussionNoteOptions, _ ...gitlab.RequestOptionFunc) (*gitlab.Note, *gitlab.Response, error) {
	return fds.updateNote("Commit", projectID, commitSHA, discussionID, noteID, *opt.Body)
}

func (fds *fakeDiscussionsService) DeleteCommitDiscussionNote(projectID any, commitSHA string, discussionID string, noteID int, _ ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	return fds.deleteNote("Commit", projectID, commitSHA, discussionID, noteID)
}

// Fake implementations

func (fds *fakeDiscussionsService) list(resourceType string, parentID, resourceID any) ([]*gitlab.Discussion, *gitlab.Response, error) {
	key := fmt.Sprintf("%s(%v,%v)", resourceType, parentID, resourceID)
	if discussions, ok := fds.discussions[key]; ok {
		return discussions, &gitlab.Response{}, nil
	}

	return nil, nil, fmt.Errorf("List%sDiscussions(%v, %v): unexpected call", resourceType, parentID, resourceID)
}

func (fds *fakeDiscussionsService) create(resourceType string, parentID, resourceID any, body string) (*gitlab.Discussion, *gitlab.Response, error) {
	key := fmt.Sprintf("%s(%v,%v)", resourceType, parentID, resourceID)

	discussion := gitlab.Discussion{
		ID: fmt.Sprintf("%d", fds.nextID),
		Notes: []*gitlab.Note{
			{
				ID:   fds.nextID,
				Body: body,
			},
		},
	}

	fds.discussions[key] = append(fds.discussions[key], &discussion)
	fds.nextID++

	return &discussion, &gitlab.Response{}, nil
}

func (fds *fakeDiscussionsService) addNote(resourceType string, parentID, resourceID any, discussionID, body string) (*gitlab.Note, *gitlab.Response, error) {
	key := fmt.Sprintf("%s(%v,%v)", resourceType, parentID, resourceID)

	discussion, ok := findDiscussion(fds.discussions[key], discussionID)
	if !ok {
		return nil, nil, fmt.Errorf("Add%sDiscussionNote(%v, %v, %s): discussion ID does not exist", resourceType, parentID, resourceID, discussionID)
	}

	note := &gitlab.Note{
		ID:   fds.nextID,
		Body: body,
	}

	discussion.Notes = append(discussion.Notes, note)
	fds.nextID++

	return note, &gitlab.Response{}, nil
}

func (fds *fakeDiscussionsService) updateNote(resourceType string, parentID, resourceID any, discussionID string, noteID int, body string) (*gitlab.Note, *gitlab.Response, error) {
	key := fmt.Sprintf("%s(%v,%v)", resourceType, parentID, resourceID)

	discussion, ok := findDiscussion(fds.discussions[key], discussionID)
	if !ok {
		return nil, nil, fmt.Errorf("Update%sDiscussionNote(%v, %v, %s, %d): discussion ID does not exist", resourceType, parentID, resourceID, discussionID, noteID)
	}

	note, ok := findNote(discussion.Notes, noteID)
	if !ok {
		return nil, nil, fmt.Errorf("Update%sDiscussionNote(%v, %v, %s, %d): note ID does not exist", resourceType, parentID, resourceID, discussionID, noteID)
	}

	note.Body = body

	return note, &gitlab.Response{}, nil
}

func (fds *fakeDiscussionsService) deleteNote(resourceType string, parentID, resourceID any, discussionID string, noteID int) (*gitlab.Response, error) {
	key := fmt.Sprintf("%s(%v,%v)", resourceType, parentID, resourceID)

	discussion, ok := findDiscussion(fds.discussions[key], discussionID)
	if !ok {
		return nil, fmt.Errorf("Delete%sDiscussionNote(%v, %v, %s, %d): discussion ID does not exist", resourceType, parentID, resourceID, discussionID, noteID)
	}

	index := -1

	for i, note := range discussion.Notes {
		if note.ID == noteID {
			index = i
			break
		}
	}

	if index == -1 {
		return nil, fmt.Errorf("Delete%sDiscussionNote(%v, %v, %s, %d): note ID does not exist", resourceType, parentID, resourceID, discussionID, noteID)
	}

	discussion.Notes = append(discussion.Notes[:index], discussion.Notes[index+1:]...)

	return &gitlab.Response{}, nil
}

// Utility functions

func findDiscussion(discussions []*gitlab.Discussion, id string) (*gitlab.Discussion, bool) {
	for _, d := range discussions {
		if d.ID == id {
			return d, true
		}
	}

	return nil, false
}

func findNote(notes []*gitlab.Note, id int) (*gitlab.Note, bool) {
	for _, n := range notes {
		if n.ID == id {
			return n, true
		}
	}

	return nil, false
}
