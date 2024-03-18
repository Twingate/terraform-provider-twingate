package query

import "github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"

type CreateServiceAccount struct {
	ServiceAccountEntityResponse `graphql:"serviceAccountCreate(name: $name)"`
}

func (q CreateServiceAccount) IsEmpty() bool {
	return q.Entity == nil
}

type ServiceAccountEntityResponse struct {
	Entity *gqlServiceAccount
	OkError
}

func (q CreateServiceAccount) ToModel() *model.ServiceAccount {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
