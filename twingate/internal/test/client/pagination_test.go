package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/stretchr/testify/assert"
)

func TestPagination(t *testing.T) {
	badError := errors.New("bad error")

	cases := []struct {
		resource *query.PaginatedResource[int]
		nextPage query.NextPageFunc[int]

		expected    *query.PaginatedResource[int]
		expectedErr error
	}{
		{},
		{
			resource: &query.PaginatedResource[int]{
				PageInfo: query.PageInfo{
					HasNextPage: false,
				},
			},
			expected: &query.PaginatedResource[int]{
				PageInfo: query.PageInfo{
					HasNextPage: false,
				},
			},
		},
		{
			resource: &query.PaginatedResource[int]{
				PageInfo: query.PageInfo{
					HasNextPage: true,
				},
			},
			nextPage: func(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[int], error) {
				return nil, badError
			},
			expected: &query.PaginatedResource[int]{
				PageInfo: query.PageInfo{
					HasNextPage: true,
				},
			},
			expectedErr: badError,
		},
		{
			resource: &query.PaginatedResource[int]{
				PageInfo: query.PageInfo{
					HasNextPage: true,
				},
				Edges: []int{1, 2},
			},
			nextPage: func(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[int], error) {
				return &query.PaginatedResource[int]{
					PageInfo: query.PageInfo{
						HasNextPage: false,
					},
					Edges: []int{3, 4},
				}, nil
			},
			expected: &query.PaginatedResource[int]{
				PageInfo: query.PageInfo{
					HasNextPage: true,
				},
				Edges: []int{1, 2, 3, 4},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			err := c.resource.FetchPages(context.TODO(), c.nextPage, map[string]interface{}{})

			assert.Equal(t, c.expected, c.resource)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}
