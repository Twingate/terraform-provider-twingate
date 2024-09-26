package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

type ReadDLPPolicy struct {
	DLPPolicy *gqlDLPPolicy `graphql:"dlpPolicy(id: $id, name: $name)"`
}

type gqlDLPPolicy struct {
	IDName
}

func (r gqlDLPPolicy) ToModel() *model.DLPPolicy {
	return &model.DLPPolicy{
		ID:   string(r.ID),
		Name: r.Name,
	}
}

func (r ReadDLPPolicy) ToModel() *model.DLPPolicy {
	if r.DLPPolicy == nil {
		return nil
	}

	return r.DLPPolicy.ToModel()
}

func (r ReadDLPPolicy) IsEmpty() bool {
	return r.DLPPolicy == nil
}
