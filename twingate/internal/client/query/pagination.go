package query

import (
	"context"

	"github.com/twingate/go-graphql-client"
)

type PageInfo struct {
	EndCursor   graphql.String
	HasNextPage graphql.Boolean
}

type PaginatedResource[E any] struct {
	PageInfo PageInfo
	Edges    []E
}

type NextPageFunc[E any] func(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[E], error)

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
