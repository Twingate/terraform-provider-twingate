package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type CreateServiceAccountKey struct {
	ServiceAccountKeyEntityCreateResponse `graphql:"serviceAccountKeyCreate(expirationTime: $expirationTime, serviceAccountId: $serviceAccountId, name: $name)"`
}

type ServiceAccountKeyEntityCreateResponse struct {
	ServiceAccountKeyEntityResponse
	Token graphql.String
}

func (q CreateServiceAccountKey) ToModel() (*model.ServiceKey, error) {
	if q.Entity == nil {
		return nil, nil //nolint
	}

	serviceKey, err := q.Entity.ToModel()
	if err != nil {
		return nil, err
	}

	serviceKey.Token = string(q.Token)

	return serviceKey, nil
}
