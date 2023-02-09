package query

import (
	"context"
)

type PageInfo struct {
	EndCursor   string
	HasNextPage bool
}

type PaginatedResource[E any] struct {
	PageInfo PageInfo
	Edges    []E
}

type NextPageFunc[E any] func(ctx context.Context, variables map[string]interface{}, cursor string) (*PaginatedResource[E], error)

func (r *PaginatedResource[E]) FetchPages(ctx context.Context, fetchNextPage NextPageFunc[E], variables map[string]interface{}) error {
	if r == nil {
		return nil
	}

	page := r.PageInfo
	for page.HasNextPage {
		next, err := fetchNextPage(ctx, variables, page.EndCursor)
		if err != nil {
			return err
		}

		r.Edges = append(r.Edges, next.Edges...)
		page = next.PageInfo
	}

	return nil
}
