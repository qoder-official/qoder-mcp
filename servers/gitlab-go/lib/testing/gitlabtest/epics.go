package gitlabtest

import (
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// FakeEpicsService is a partial implementation of the gitlab.EpicsServiceInterface to be used in unit tests.
// Calling unimplemented methods will panic.
type FakeEpicsService struct {
	Epics []*gitlab.Epic

	gitlab.EpicsServiceInterface
}

func (f *FakeEpicsService) ListGroupEpics(groupID any, opt *gitlab.ListGroupEpicsOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Epic, *gitlab.Response, error) {
	id, ok := groupID.(int)
	if !ok {
		return nil, nil, ErrNumericIDRequired
	}

	var epics []*gitlab.Epic

	for _, epic := range f.Epics {
		if epic.GroupID != id {
			continue
		}

		if opt.State != nil && *opt.State != "" && epic.State != *opt.State {
			continue
		}

		epics = append(epics, epic)
	}

	return epics, &gitlab.Response{}, nil
}

func (f *FakeEpicsService) GetEpic(groupID any, epicIID int, _ ...gitlab.RequestOptionFunc) (*gitlab.Epic, *gitlab.Response, error) {
	id, ok := groupID.(int)
	if !ok {
		return nil, nil, ErrNumericIDRequired
	}

	for _, epic := range f.Epics {
		if epic.GroupID == id && epic.IID == epicIID {
			return epic, &gitlab.Response{}, nil
		}
	}

	return nil, &gitlab.Response{}, nil
}

func (f *FakeEpicsService) GetEpicLinks(groupID any, epicIID int, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Epic, *gitlab.Response, error) {
	id, ok := groupID.(int)
	if !ok {
		return nil, nil, ErrNumericIDRequired
	}

	epicID := -1

	for _, epic := range f.Epics {
		if epic.GroupID == id && epic.IID == epicIID {
			epicID = epic.ID
			break
		}
	}

	if epicID == -1 {
		return nil, nil, fmt.Errorf("%w: epicIID=%d, groupID=%d", ErrNotFound, epicIID, id)
	}

	var epics []*gitlab.Epic

	for _, epic := range f.Epics {
		if epic.ParentID != epicID {
			continue
		}

		epics = append(epics, epic)
	}

	return epics, &gitlab.Response{}, nil
}
