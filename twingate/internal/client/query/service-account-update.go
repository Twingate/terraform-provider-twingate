package query

import "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"

type UpdateServiceAccount struct {
	ServiceAccountEntityResponse `graphql:"serviceAccountUpdate(id: $id, name: $name, addedResourceIds: $addedResourceIds)"`
}

func (q UpdateServiceAccount) IsEmpty() bool {
	return q.Entity == nil
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

func (q UpdateServiceAccountRemoveResources) IsEmpty() bool {
	return q.Entity == nil
}
