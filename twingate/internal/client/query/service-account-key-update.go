package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

type UpdateServiceAccountKey struct {
	ServiceAccountKeyEntityResponse `graphql:"serviceAccountKeyUpdate(id: $id, name: $name)"`
}

func (q UpdateServiceAccountKey) IsEmpty() bool {
	return q.Entity == nil
}

type ServiceAccountKeyEntityResponse struct {
	Entity *gqlServiceKey
	OkError
}

func (q UpdateServiceAccountKey) ToModel() (*model.ServiceKey, error) {
	if q.Entity == nil {
		return nil, nil
	}

	return q.Entity.ToModel()
}
