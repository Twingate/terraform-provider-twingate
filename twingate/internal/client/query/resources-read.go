package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const (
	CursorResources    = "resourcesEndCursor"
	PageLimitResources = "resourcesPageLimit"
)

type ReadResources struct {
	Resources `graphql:"resources(after: $resourcesEndCursor, first: $resourcesPageLimit)"`
}

func (r ReadResources) IsEmpty() bool {
	return len(r.Edges) == 0
}

type Resources struct {
	PaginatedResource[*ResourceEdge]
}

type ResourceEdge struct {
	Node *ResourceNode
}

func (r Resources) ToModel() []*model.Resource {
	return utils.Map[*ResourceEdge, *model.Resource](r.Edges, func(edge *ResourceEdge) *model.Resource {
		return edge.Node.ToModel()
	})
}
