package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func TestPagination(t *testing.T) {
	badError := errors.New("bad error")

	cases := []struct {
		resource *PaginatedResource[int]
		nextPage nextPageFunc[int]

		expected    *PaginatedResource[int]
		expectedErr error
	}{
		{},
		{
			resource: &PaginatedResource[int]{
				PageInfo: PageInfo{
					HasNextPage: false,
				},
			},
			expected: &PaginatedResource[int]{
				PageInfo: PageInfo{
					HasNextPage: false,
				},
			},
		},
		{
			resource: &PaginatedResource[int]{
				PageInfo: PageInfo{
					HasNextPage: true,
				},
			},
			nextPage: func(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[int], error) {
				return nil, badError
			},
			expected: &PaginatedResource[int]{
				PageInfo: PageInfo{
					HasNextPage: true,
				},
			},
			expectedErr: badError,
		},
		{
			resource: &PaginatedResource[int]{
				PageInfo: PageInfo{
					HasNextPage: true,
				},
				Edges: []int{1, 2},
			},
			nextPage: func(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[int], error) {
				return &PaginatedResource[int]{
					PageInfo: PageInfo{
						HasNextPage: false,
					},
					Edges: []int{3, 4},
				}, nil
			},
			expected: &PaginatedResource[int]{
				PageInfo: PageInfo{
					HasNextPage: true,
				},
				Edges: []int{1, 2, 3, 4},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			err := c.resource.fetchPages(context.TODO(), c.nextPage, map[string]interface{}{})

			assert.Equal(t, c.expected, c.resource)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}
