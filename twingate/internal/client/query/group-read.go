package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

type ReadGroup struct {
	Group *gqlGroup `graphql:"group(id: $id)"`
}

type gqlGroup struct {
	IDName
	IsActive       bool
	Type           string
	SecurityPolicy gqlSecurityPolicy
}

func (g gqlGroup) ToModel() *model.Group {
	return &model.Group{
		ID:               string(g.ID),
		Name:             g.Name,
		Type:             g.Type,
		IsActive:         g.IsActive,
		SecurityPolicyID: string(g.SecurityPolicy.ID),
	}
}

func (q ReadGroup) ToModel() *model.Group {
	if q.Group == nil {
		return nil
	}

	return q.Group.ToModel()
}
