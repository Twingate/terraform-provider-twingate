package query

import "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"

type CreateGroup struct {
	GroupEntityResponse `graphql:"groupCreate(name: $name, userIds: $userIds, securityPolicyId: $securityPolicyId)"`
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

func (q CreateGroup) IsEmpty() bool {
	return q.Entity == nil
}
