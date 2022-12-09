package query

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

type ReadServiceAccount struct {
	ServiceAccount *gqlServiceAccount `graphql:"serviceAccount(id: $id)"`
}

type gqlServiceAccount struct {
	IDName
}

func (q gqlServiceAccount) ToModel() *model.ServiceAccount {
	return &model.ServiceAccount{
		ID:   q.StringID(),
		Name: q.StringName(),
	}
}

func (q ReadServiceAccount) ToModel() *model.ServiceAccount {
	if q.ServiceAccount == nil {
		return nil
	}

	return q.ServiceAccount.ToModel()
}
