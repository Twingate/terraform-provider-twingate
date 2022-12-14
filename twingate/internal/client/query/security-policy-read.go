package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

type ReadSecurityPolicy struct {
	SecurityPolicy *gqlSecurityPolicy `graphql:"securityPolicy(id: $id, name: $name)"`
}

type gqlSecurityPolicy struct {
	IDName
}

func (q ReadSecurityPolicy) ToModel() *model.SecurityPolicy {
	if q.SecurityPolicy == nil {
		return nil
	}

	return q.SecurityPolicy.ToModel()
}

func (q *gqlSecurityPolicy) ToModel() *model.SecurityPolicy {
	return &model.SecurityPolicy{
		ID:   q.StringID(),
		Name: q.StringName(),
	}
}
