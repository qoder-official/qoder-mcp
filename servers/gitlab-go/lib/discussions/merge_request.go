package discussions

import (
	"context"
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

// MergeRequestDiscussion provides methods for managing discussions on GitLab merge requests.
type MergeRequestDiscussion struct {
	client          *gitlab.Client
	projectID       mcpargs.ID
	mergeRequestIID int
}

var _ ResolvableManager = (*MergeRequestDiscussion)(nil)
var _ PositionedManager = (*MergeRequestDiscussion)(nil)

// NewMergeRequestDiscussion creates a new instance for managing discussions on a specific GitLab merge request.
func NewMergeRequestDiscussion(client *gitlab.Client, projectID mcpargs.ID, mergeRequestIID int) (*MergeRequestDiscussion, error) {
	if client == nil {
		return nil, fmt.Errorf("%w: client cannot be nil", ErrInvalidArgument)
	}

	return &MergeRequestDiscussion{
		client:          client,
		projectID:       projectID,
		mergeRequestIID: mergeRequestIID,
	}, nil
}

// List returns all discussions for the merge request.
func (d *MergeRequestDiscussion) List(ctx context.Context, confidential bool) ([]*gitlab.Discussion, error) {
	var (
		opt = &gitlab.ListMergeRequestDiscussionsOptions{
			PerPage: maxPerPage,
		}
		allDiscussions []*gitlab.Discussion
	)

	for {
		discussions, resp, err := d.client.Discussions.ListMergeRequestDiscussions(d.projectID.Value(), d.mergeRequestIID, opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to list merge request discussions: %w", err)
		}

		allDiscussions = append(allDiscussions, discussions...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return filterDiscussions(allDiscussions, confidential), nil
}

// NewDiscussion creates a new discussion thread on the merge request.
func (d *MergeRequestDiscussion) NewDiscussion(ctx context.Context, body string) (*gitlab.Discussion, error) {
	opt := &gitlab.CreateMergeRequestDiscussionOptions{
		Body: gitlab.Ptr(body),
	}

	discussion, _, err := d.client.Discussions.CreateMergeRequestDiscussion(d.projectID.Value(), d.mergeRequestIID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create merge request discussion: %w", err)
	}

	return discussion, nil
}

var _ PositionedManager = (*MergeRequestDiscussion)(nil)

// NewPositionDiscussion creates a new discussion on a specific position in the merge request diff.
func (d *MergeRequestDiscussion) NewPositionDiscussion(ctx context.Context, body string, pos *gitlab.PositionOptions, commitID string) (*gitlab.Discussion, error) {
	opt := &gitlab.CreateMergeRequestDiscussionOptions{
		Body:     gitlab.Ptr(body),
		Position: pos,
	}

	if commitID != "" {
		opt.CommitID = gitlab.Ptr(commitID)
	}

	discussion, _, err := d.client.Discussions.CreateMergeRequestDiscussion(d.projectID.Value(), d.mergeRequestIID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create diff discussion: %w", err)
	}

	return discussion, nil
}

// AddNote adds a new note to an existing discussion thread.
func (d *MergeRequestDiscussion) AddNote(ctx context.Context, discussionID string, body string) (*gitlab.Note, error) {
	opt := &gitlab.AddMergeRequestDiscussionNoteOptions{
		Body: gitlab.Ptr(body),
	}

	note, _, err := d.client.Discussions.AddMergeRequestDiscussionNote(d.projectID.Value(), d.mergeRequestIID, discussionID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to add note to merge request discussion: %w", err)
	}

	return note, nil
}

// ModifyNote updates an existing note in a discussion thread.
func (d *MergeRequestDiscussion) ModifyNote(ctx context.Context, discussionID string, noteID int, body string) (*gitlab.Note, error) {
	opt := &gitlab.UpdateMergeRequestDiscussionNoteOptions{
		Body: gitlab.Ptr(body),
	}

	note, _, err := d.client.Discussions.UpdateMergeRequestDiscussionNote(d.projectID.Value(), d.mergeRequestIID, discussionID, noteID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to modify merge request note: %w", err)
	}

	return note, nil
}

// DeleteNote removes a note from a discussion thread.
func (d *MergeRequestDiscussion) DeleteNote(ctx context.Context, discussionID string, noteID int) error {
	_, err := d.client.Discussions.DeleteMergeRequestDiscussionNote(d.projectID.Value(), d.mergeRequestIID, discussionID, noteID, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete merge request note: %w", err)
	}

	return nil
}

// ResolveDiscussion resolves or unresolves a thread of discussion in a merge request.
func (d *MergeRequestDiscussion) ResolveDiscussion(ctx context.Context, discussionID string, resolved bool) (*gitlab.Discussion, error) {
	opt := &gitlab.ResolveMergeRequestDiscussionOptions{
		Resolved: gitlab.Ptr(resolved),
	}

	discussion, _, err := d.client.Discussions.ResolveMergeRequestDiscussion(d.projectID.Value(), d.mergeRequestIID, discussionID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to %s discussion: %w",
			resolveActionText(resolved), err)
	}

	return discussion, nil
}

// resolveActionText returns the appropriate text based on the resolved status.
func resolveActionText(resolved bool) string {
	if resolved {
		return "resolve"
	}

	return "unresolve"
}
