// Package discussions provides client-side interfaces for working with GitLab discussions.
//
// This package offers a set of type-safe, idiomatic Go wrappers around GitLab's
// Discussions API endpoints, allowing for the creation, retrieval, modification, and
// deletion of discussion threads and notes across different GitLab resources:
// issues, merge requests, epics, snippets, and commits.
//
// Each resource type has its own specialized manager type (IssueDiscussion, MergeRequestDiscussion, etc.)
// that implements common methods for working with discussions, while also providing
// resource-specific functionality where applicable (such as resolving merge request discussions).
//
// Example usage:
//
//	// Create a new discussion on an issue
//	issueDiscussions, err := discussions.NewIssueDiscussion(client, "mygroup/myproject", 123)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a new discussion thread
//	discussion, err := issueDiscussions.NewDiscussion(ctx, "This is a new comment")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Add a note to the discussion
//	n, err := issueDiscussions.AddNote(ctx, discussion.ID, "This is a reply")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// All methods accept a context.Context parameter to support timeouts, cancellation,
// and other context-driven behaviors.
package discussions

import (
	"context"
	"errors"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const maxPerPage = 100

var ErrInvalidArgument = errors.New("invalid argument")

// Manager is an interface that defines common methods for managing discussions
// across different GitLab resources (issues, merge requests, epics, snippets, commits).
// This interface allows for consistent handling of discussions regardless of the resource type.
type Manager interface {
	// List returns all discussions for a specific resource.
	List(ctx context.Context, confidential bool) ([]*gitlab.Discussion, error)

	// NewDiscussion creates a new discussion thread on a resource.
	NewDiscussion(ctx context.Context, body string) (*gitlab.Discussion, error)

	// AddNote adds a new note to an existing discussion thread.
	AddNote(ctx context.Context, discussionID string, body string) (*gitlab.Note, error)

	// ModifyNote updates an existing note in a discussion thread.
	ModifyNote(ctx context.Context, discussionID string, noteID int, body string) (*gitlab.Note, error)

	// DeleteNote removes a note from a discussion thread.
	DeleteNote(ctx context.Context, discussionID string, noteID int) error
}

// ResolvableManager extends Manager with methods for resolving discussions,
// which is only applicable to merge request discussions.
type ResolvableManager interface {
	Manager

	// ResolveDiscussion resolves or unresolves a thread of discussion.
	ResolveDiscussion(ctx context.Context, discussionID string, resolved bool) (*gitlab.Discussion, error)
}

// PositionedManager extends Manager with methods for working with discussion
// threads that have a specific position reference, such as diff discussions on merge requests.
type PositionedManager interface {
	Manager

	NewPositionDiscussion(ctx context.Context, body string, position *gitlab.PositionOptions, commitID string) (*gitlab.Discussion, error)
}

// filterDiscussions filters discussions and their notes based on the 'includeConfidential' flag.
func filterDiscussions(discussions []*gitlab.Discussion, includeConfidential bool) []*gitlab.Discussion {
	if includeConfidential {
		return discussions
	}

	var ret []*gitlab.Discussion

	for _, discussion := range discussions {
		d := *discussion
		d.Notes = filterNotes(d.Notes)

		if len(d.Notes) > 0 {
			ret = append(ret, &d)
		}
	}

	return ret
}

// filterNotes filters out internal notes from a list of notes.
func filterNotes(notes []*gitlab.Note) []*gitlab.Note {
	var ret []*gitlab.Note

	for _, note := range notes {
		if !note.Internal {
			ret = append(ret, note)
		}
	}

	return ret
}
