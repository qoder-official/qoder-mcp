package discussions

import (
	"context"
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"gitlab.com/fforster/gitlab-mcp/lib/mcpargs"
)

// EpicDiscussion provides methods for managing discussions on GitLab epics.
type EpicDiscussion struct {
	client  *gitlab.Client
	groupID mcpargs.ID
	epicID  int
}

var _ Manager = (*EpicDiscussion)(nil)

// NewEpicDiscussion creates a new instance for managing discussions on a specific GitLab epic.
func NewEpicDiscussion(client *gitlab.Client, groupID mcpargs.ID, epicID int) (*EpicDiscussion, error) {
	if client == nil {
		return nil, fmt.Errorf("%w: client cannot be nil", ErrInvalidArgument)
	}

	if groupID.IsZero() {
		return nil, fmt.Errorf("%w: groupID cannot be zero", ErrInvalidArgument)
	}

	if epicID == 0 {
		return nil, fmt.Errorf("%w: epicID cannot be zero", ErrInvalidArgument)
	}

	return &EpicDiscussion{
		client:  client,
		groupID: groupID,
		epicID:  epicID,
	}, nil
}

// List returns all discussions for the epic.
func (d *EpicDiscussion) List(ctx context.Context, confidential bool) ([]*gitlab.Discussion, error) {
	var (
		opt = &gitlab.ListGroupEpicDiscussionsOptions{
			PerPage: maxPerPage,
		}
		allDiscussions []*gitlab.Discussion
	)

	for {
		discussions, resp, err := d.client.Discussions.ListGroupEpicDiscussions(d.groupID.Value(), d.epicID, opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("ListGroupEpicDiscussions(%v, %d): %w", d.groupID.Value(), d.epicID, err)
		}

		allDiscussions = append(allDiscussions, discussions...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return filterDiscussions(allDiscussions, confidential), nil
}

// NewDiscussion creates a new discussion thread on the epic.
func (d *EpicDiscussion) NewDiscussion(ctx context.Context, body string) (*gitlab.Discussion, error) {
	opt := &gitlab.CreateEpicDiscussionOptions{
		Body: gitlab.Ptr(body),
	}

	discussion, _, err := d.client.Discussions.CreateEpicDiscussion(d.groupID.Value(), d.epicID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create epic discussion: %w", err)
	}

	return discussion, nil
}

// AddNote adds a new note to an existing discussion thread.
func (d *EpicDiscussion) AddNote(ctx context.Context, discussionID string, body string) (*gitlab.Note, error) {
	opt := &gitlab.AddEpicDiscussionNoteOptions{
		Body: gitlab.Ptr(body),
	}

	note, _, err := d.client.Discussions.AddEpicDiscussionNote(d.groupID.Value(), d.epicID, discussionID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to add note to epic discussion: %w", err)
	}

	return note, nil
}

// ModifyNote updates an existing note in a discussion thread.
func (d *EpicDiscussion) ModifyNote(ctx context.Context, discussionID string, noteID int, body string) (*gitlab.Note, error) {
	opt := &gitlab.UpdateEpicDiscussionNoteOptions{
		Body: gitlab.Ptr(body),
	}

	note, _, err := d.client.Discussions.UpdateEpicDiscussionNote(d.groupID.Value(), d.epicID, discussionID, noteID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to modify epic note: %w", err)
	}

	return note, nil
}

// DeleteNote removes a note from a discussion thread.
func (d *EpicDiscussion) DeleteNote(ctx context.Context, discussionID string, noteID int) error {
	_, err := d.client.Discussions.DeleteEpicDiscussionNote(d.groupID.Value(), d.epicID, discussionID, noteID, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete epic note: %w", err)
	}

	return nil
}
