package query

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchPages(t *testing.T) {
	cases := []struct {
		name            string
		initialResource *PaginatedResource[string]
		fetchNextPage   func(ctx context.Context, variables map[string]interface{}, cursor string) (*PaginatedResource[string], error)
		expectedEdges   []string
		expectedError   error
	}{
		{
			name: "No additional pages (HasNextPage false)",
			initialResource: &PaginatedResource[string]{
				PageInfo: PageInfo{
					EndCursor:   "cursor1",
					HasNextPage: false,
				},
				Edges: []string{"edge1", "edge2"},
			},
			fetchNextPage: func(ctx context.Context, variables map[string]interface{}, cursor string) (*PaginatedResource[string], error) {
				return nil, nil // Should not be called since HasNextPage is false
			},
			expectedEdges: []string{"edge1", "edge2"},
			expectedError: nil,
		},
		{
			name: "Single additional page",
			initialResource: &PaginatedResource[string]{
				PageInfo: PageInfo{
					EndCursor:   "cursor1",
					HasNextPage: true,
				},
				Edges: []string{"edge1"},
			},
			fetchNextPage: func(ctx context.Context, variables map[string]interface{}, cursor string) (*PaginatedResource[string], error) {
				if cursor == "cursor1" {
					return &PaginatedResource[string]{
						PageInfo: PageInfo{
							EndCursor:   "cursor2",
							HasNextPage: false,
						},
						Edges: []string{"edge2"},
					}, nil
				}
				return nil, errors.New("unexpected cursor")
			},
			expectedEdges: []string{"edge1", "edge2"},
			expectedError: nil,
		},
		{
			name: "Multiple additional pages",
			initialResource: &PaginatedResource[string]{
				PageInfo: PageInfo{
					EndCursor:   "cursor1",
					HasNextPage: true,
				},
				Edges: []string{"edge1"},
			},
			fetchNextPage: func(ctx context.Context, variables map[string]interface{}, cursor string) (*PaginatedResource[string], error) {
				switch cursor {
				case "cursor1":
					return &PaginatedResource[string]{
						PageInfo: PageInfo{
							EndCursor:   "cursor2",
							HasNextPage: true,
						},
						Edges: []string{"edge2"},
					}, nil
				case "cursor2":
					return &PaginatedResource[string]{
						PageInfo: PageInfo{
							EndCursor:   "cursor3",
							HasNextPage: false,
						},
						Edges: []string{"edge3"},
					}, nil
				default:
					return nil, errors.New("unexpected cursor")
				}
			},
			expectedEdges: []string{"edge1", "edge2", "edge3"},
			expectedError: nil,
		},
		{
			name: "Error fetching next page",
			initialResource: &PaginatedResource[string]{
				PageInfo: PageInfo{
					EndCursor:   "cursor1",
					HasNextPage: true,
				},
				Edges: []string{"edge1"},
			},
			fetchNextPage: func(ctx context.Context, variables map[string]interface{}, cursor string) (*PaginatedResource[string], error) {
				return nil, errors.New("fetch error")
			},
			expectedEdges: []string{"edge1"},
			expectedError: errors.New("fetch error"),
		},
		{
			name:            "Nil initial resource",
			initialResource: nil,
			fetchNextPage: func(ctx context.Context, variables map[string]interface{}, cursor string) (*PaginatedResource[string], error) {
				return nil, nil // Should not be called since resource is nil
			},
			expectedEdges: nil,
			expectedError: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.initialResource.FetchPages(context.Background(), c.fetchNextPage, nil)

			if c.expectedError != nil {
				assert.EqualError(t, err, c.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			if c.initialResource != nil {
				assert.Equal(t, c.expectedEdges, c.initialResource.Edges)
			}
		})
	}
}
