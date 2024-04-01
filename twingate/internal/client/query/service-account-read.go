package query

import "github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"

type ReadShallowServiceAccount struct {
	ServiceAccount *gqlServiceAccount `graphql:"serviceAccount(id: $id)"`
}

func (q ReadShallowServiceAccount) IsEmpty() bool {
	return q.ServiceAccount == nil
}

type gqlServiceAccount struct {
	IDName
}

func (q gqlServiceAccount) ToModel() *model.ServiceAccount {
	return &model.ServiceAccount{
		ID:   string(q.ID),
		Name: q.Name,
	}
}

func (q ReadShallowServiceAccount) ToModel() *model.ServiceAccount {
	if q.ServiceAccount == nil {
		return nil
	}

	return q.ServiceAccount.ToModel()
}

type ReadServiceAccount struct {
	Service *GqlService `graphql:"serviceAccount(id: $id)"`
}

func (q ReadServiceAccount) IsEmpty() bool {
	return q.Service == nil || len(q.Service.Resources.Edges) == 0 && len(q.Service.Keys.Edges) == 0
}
