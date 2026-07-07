package query

import (
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/utils"
)

const CursorResources = "resourcesEndCursor"

type ReadResources struct {
	Resources `graphql:"resources(after: $resourcesEndCursor, first: $pageLimit)"`
}

func (r ReadResources) IsEmpty() bool {
	return len(r.Edges) == 0
}

type Resources struct {
	PaginatedResource[*ResourceEdge]
}

type ResourceEdge struct {
	Node *gqlNetworkResource
}

// gqlNetworkResource selects the NetworkResource fields of the Resource interface
// returned by the `resources` connection. The `... on NetworkResource` type
// condition must sit on an embedded field (as in gqlResource), not on the named
// `node` field, otherwise the decoder has no `node` field to unmarshal into.
type gqlNetworkResource struct {
	ResourceNode `graphql:"... on NetworkResource"`
}

func (r Resources) ToModel() []*model.Resource {
	return utils.Map[*ResourceEdge, *model.Resource](r.Edges, func(edge *ResourceEdge) *model.Resource {
		return edge.Node.ToModel()
	})
}

type ReadFullResourcesByName struct {
	FullResources `graphql:"resources(filter: $filter, after: $resourcesEndCursor, first: $pageLimit)"`
}

func (r ReadFullResourcesByName) IsEmpty() bool {
	return len(r.Edges) == 0
}

func (r ReadFullResourcesByName) ToModel() ([]*model.Resource, error) {
	return utils.MapWithError[*FullResourceEdge, *model.Resource](r.Edges, func(edge *FullResourceEdge) (*model.Resource, error) {
		return edge.Node.ToModel()
	})
}

type ReadFullResources struct {
	FullResources `graphql:"resources(after: $resourcesEndCursor, first: $pageLimit)"`
}

func (r ReadFullResources) IsEmpty() bool {
	return len(r.Edges) == 0
}

type FullResources struct {
	PaginatedResource[*FullResourceEdge]
}

type FullResourceEdge struct {
	Node *gqlResource
}

func (r ReadFullResources) ToModel() ([]*model.Resource, error) {
	return utils.MapWithError[*FullResourceEdge, *model.Resource](r.Edges, func(edge *FullResourceEdge) (*model.Resource, error) {
		return edge.Node.ToModel()
	})
}
