package discussions

import (
	"context"
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

// IssueDiscussion provides methods for managing discussions on GitLab issues.
type IssueDiscussion struct {
	client    *gitlab.Client
	projectID mcpargs.ID
	issueIID  int
}

var _ Manager = (*IssueDiscussion)(nil)

// NewIssueDiscussion creates a new instance for managing discussions on a specific GitLab issue.
func NewIssueDiscussion(client *gitlab.Client, projectID mcpargs.ID, issueIID int) (*IssueDiscussion, error) {
	if client == nil {
		return nil, fmt.Errorf("%w: client cannot be nil", ErrInvalidArgument)
	}

	if projectID.IsZero() {
		return nil, fmt.Errorf("%w: projectID cannot be zero", ErrInvalidArgument)
	}

	if issueIID == 0 {
		return nil, fmt.Errorf("%w: issueIID cannot be zero", ErrInvalidArgument)
	}

	return &IssueDiscussion{
		client:    client,
		projectID: projectID,
		issueIID:  issueIID,
	}, nil
}

// List returns all discussions for the issue.
func (d *IssueDiscussion) List(ctx context.Context, confidential bool) ([]*gitlab.Discussion, error) {
	var (
		opt = &gitlab.ListIssueDiscussionsOptions{
			PerPage: maxPerPage,
		}
		allDiscussions []*gitlab.Discussion
	)

	for {
		discussions, resp, err := d.client.Discussions.ListIssueDiscussions(d.projectID.Value(), d.issueIID, opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to list discussions: %w", err)
		}

		allDiscussions = append(allDiscussions, discussions...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return filterDiscussions(allDiscussions, confidential), nil
}

// NewDiscussion creates a new discussion thread on the issue.
func (d *IssueDiscussion) NewDiscussion(ctx context.Context, body string) (*gitlab.Discussion, error) {
	opt := &gitlab.CreateIssueDiscussionOptions{
		Body: gitlab.Ptr(body),
	}

	discussion, _, err := d.client.Discussions.CreateIssueDiscussion(d.projectID.Value(), d.issueIID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create discussion: %w", err)
	}

	return discussion, nil
}

// AddNote adds a new note to an existing discussion thread.
func (d *IssueDiscussion) AddNote(ctx context.Context, discussionID string, body string) (*gitlab.Note, error) {
	opt := &gitlab.AddIssueDiscussionNoteOptions{
		Body: gitlab.Ptr(body),
	}

	note, _, err := d.client.Discussions.AddIssueDiscussionNote(d.projectID.Value(), d.issueIID, discussionID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to add note to discussion: %w", err)
	}

	return note, nil
}

// ModifyNote updates an existing note in a discussion thread.
func (d *IssueDiscussion) ModifyNote(ctx context.Context, discussionID string, noteID int, body string) (*gitlab.Note, error) {
	opt := &gitlab.UpdateIssueDiscussionNoteOptions{
		Body: gitlab.Ptr(body),
	}

	note, _, err := d.client.Discussions.UpdateIssueDiscussionNote(d.projectID.Value(), d.issueIID, discussionID, noteID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to modify note: %w", err)
	}

	return note, nil
}

// DeleteNote removes a note from a discussion thread.
func (d *IssueDiscussion) DeleteNote(ctx context.Context, discussionID string, noteID int) error {
	_, err := d.client.Discussions.DeleteIssueDiscussionNote(d.projectID.Value(), d.issueIID, discussionID, noteID, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}
