package query

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

type UpdateServiceAccountKey struct {
	ServiceAccountKeyEntityResponse `graphql:"serviceAccountKeyUpdate(id: $id, name: $name)"`
}

func (q UpdateServiceAccountKey) ToModel() (*model.ServiceKey, error) {
	if q.Entity == nil {
		return nil, nil //nolint
	}

	return q.Entity.ToModel()
}
