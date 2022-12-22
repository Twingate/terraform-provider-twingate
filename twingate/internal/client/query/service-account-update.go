package query

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

type UpdateServiceAccount struct {
	ServiceAccountEntityResponse `graphql:"serviceAccountUpdate(id: $id, name: $name, addedResourceIds: $addedResourceIds)"`
}

func (q UpdateServiceAccount) ToModel() *model.ServiceAccount {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}

type UpdateServiceAccountRemoveResources struct {
	ServiceAccountEntityResponse `graphql:"serviceAccountUpdate(id: $id, removedResourceIds: $removedResourceIds)"`
}
