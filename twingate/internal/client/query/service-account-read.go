package query

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

type ReadShallowServiceAccount struct {
	ServiceAccount *gqlServiceAccount `graphql:"serviceAccount(id: $id)"`
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
