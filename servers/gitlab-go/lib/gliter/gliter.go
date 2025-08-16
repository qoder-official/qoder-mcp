// Package gliter implements iterators for GitLab objects.
package gliter

import (
	"context"
	"iter"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// All func iterates over all items fetched by the getter function, handling pagination automatically.
// It yields each item and stops when either an error occurs or the yield function returns false.
// The nextPage function is called with the current options and the next page number to continue pagination.
func All[T any, O any](ctx context.Context, getter func(opts *O, options ...gitlab.RequestOptionFunc) ([]T, *gitlab.Response, error), opts O, nextPage func(*O, int)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for {
			items, resp, err := getter(&opts, gitlab.WithContext(ctx))
			if err != nil {
				var zero T

				yield(zero, err)

				return
			}

			for _, item := range items {
				if !yield(item, nil) {
					return
				}
			}

			if resp.NextPage == 0 {
				break
			}

			nextPage(&opts, resp.NextPage)
		}
	}
}

// AllWithID is a wrapper for All that includes an ID in the getter function.
// It is useful for fetching items that are related to a specific ID, such as merge requests within a project or group.
func AllWithID[T any, O any](ctx context.Context, id any, getter func(id any, opts *O, options ...gitlab.RequestOptionFunc) ([]T, *gitlab.Response, error), opts O, nextPage func(*O, int)) iter.Seq2[T, error] {
	wrappedGetter := func(opts *O, options ...gitlab.RequestOptionFunc) ([]T, *gitlab.Response, error) {
		return getter(id, opts, options...)
	}

	return All(ctx, wrappedGetter, opts, nextPage)
}

// Limited is a utility function that limits the number of items yielded by a sequence.
// It stops yielding items after the specified number of items has been reached.
func Limited[T any](seq iter.Seq2[T, error], n int) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var count int

		for item, err := range seq {
			if err != nil {
				yield(item, err)
				return
			}

			if !yield(item, nil) {
				return
			}

			count++
			if count >= n {
				return
			}
		}
	}
}
