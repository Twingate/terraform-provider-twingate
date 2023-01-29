package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/twingate/go-graphql-client"
)

type ReadGroup struct {
	Group *gqlGroup `graphql:"group(id: $id)"`
}

type gqlGroup struct {
	IDName
	IsActive graphql.Boolean
	Type     graphql.String
	Users    Users
}

func (g gqlGroup) ToModel() *model.Group {
	return &model.Group{
		ID:       g.StringID(),
		Name:     g.StringName(),
		Type:     string(g.Type),
		IsActive: bool(g.IsActive),
		Users: utils.Map[*UserEdge, string](g.Users.Edges, func(edge *UserEdge) string {
			return idToString(edge.Node.ID)
		}),
	}
}

func (q ReadGroup) ToModel() *model.Group {
	if q.Group == nil {
		return nil
	}

	return q.Group.ToModel()
}
