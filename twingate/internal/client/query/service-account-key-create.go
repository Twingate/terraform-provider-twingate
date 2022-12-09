package query

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

type CreateServiceAccountKey struct {
	ServiceAccountKeyEntityResponse `graphql:"serviceAccountKeyCreate(expirationTime: $expirationTime, serviceAccountId: $serviceAccountId, name: $name)"`
}

type ServiceAccountKeyEntityResponse struct {
	Entity *gqlServiceKey
	OkError
}

func (q CreateServiceAccountKey) ToModel() (*model.ServiceKey, error) {
	if q.Entity == nil {
		return nil, nil //nolint
	}

	return q.Entity.ToModel()
}
