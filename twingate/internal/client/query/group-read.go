package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

type ReadGroup struct {
	Group *gqlGroup `graphql:"group(id: $id)"`
}

type gqlGroup struct {
	IDName
	IsActive bool
	Type     string
	Users    Users `graphql:"users(after: $usersEndCursor, first: $pageLimit)"`
}

func (g gqlGroup) ToModel() *model.Group {
	return &model.Group{
		ID:       string(g.ID),
		Name:     g.Name,
		Type:     g.Type,
		IsActive: g.IsActive,
		Users: utils.Map[*UserEdge, string](g.Users.Edges, func(edge *UserEdge) string {
			return string(edge.Node.ID)
		}),
	}
}

func (q ReadGroup) ToModel() *model.Group {
	if q.Group == nil {
		return nil
	}

	return q.Group.ToModel()
}

func (q ReadGroup) IsEmpty() bool {
	return q.Group == nil
}
