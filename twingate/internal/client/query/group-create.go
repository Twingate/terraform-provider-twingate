package query

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

type CreateGroup struct {
	GroupEntityResponse `graphql:"groupCreate(name: $name, securityPolicyId: $securityPolicyId)"`
}

type GroupEntityResponse struct {
	Entity *gqlGroup
	OkError
}

func (q CreateGroup) ToModel() *model.Group {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
