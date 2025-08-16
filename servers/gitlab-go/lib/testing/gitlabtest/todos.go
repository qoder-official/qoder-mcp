package gitlabtest

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// FakeTodosService is a partial implementation of the gitlab.TodosServiceInterface for unit tests.
type FakeTodosService struct {
	Todos []*gitlab.Todo

	gitlab.TodosServiceInterface
}

// ListTodos filters todos by state (if provided) and paginates results.
func (f *FakeTodosService) ListTodos(opt *gitlab.ListTodosOptions, _ ...gitlab.RequestOptionFunc) ([]*gitlab.Todo, *gitlab.Response, error) {
	var todos []*gitlab.Todo

	state := ""
	if opt != nil && opt.State != nil {
		state = *opt.State
	}

	for _, todo := range f.Todos {
		if state != "" && todo.State != state {
			continue
		}

		todos = append(todos, todo)
	}

	return todos, &gitlab.Response{NextPage: 0}, nil
}
