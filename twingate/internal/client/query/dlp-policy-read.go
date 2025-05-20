package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

type ReadDLPPolicy struct {
	DLPPolicy *gqlDLPPolicy `graphql:"dlpPolicy(id: $id, name: $name)"`
}

func (q ReadDLPPolicy) IsEmpty() bool {
	return q.DLPPolicy == nil
}

type gqlDLPPolicy struct {
	IDName
}

func (q ReadDLPPolicy) ToModel() *model.DLPPolicy {
	if q.DLPPolicy == nil {
		return nil
	}

	return q.DLPPolicy.ToModel()
}

func (q *gqlDLPPolicy) ToModel() *model.DLPPolicy {
	return &model.DLPPolicy{
		ID:   string(q.ID),
		Name: q.Name,
	}
}