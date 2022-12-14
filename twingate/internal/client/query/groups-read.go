package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const CursorGroups = "groupsEndCursor"

type ReadGroups struct {
	Groups `graphql:"groups(after: $groupsEndCursor)"`
}

type Groups struct {
	PaginatedResource[*GroupEdge]
}

type GroupEdge struct {
	Node *gqlGroup
}

func (u Groups) ToModel() []*model.Group {
	return utils.Map[*GroupEdge, *model.Group](u.Edges, func(edge *GroupEdge) *model.Group {
		return edge.Node.ToModel()
	})
}
