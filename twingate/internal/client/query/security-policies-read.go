package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

const CursorPolicies = "policiesEndCursor"

type ReadSecurityPolicies struct {
	SecurityPolicies `graphql:"securityPolicies(filter: $filter, after: $policiesEndCursor, first: $pageLimit)"`
}

func (q ReadSecurityPolicies) IsEmpty() bool {
	return len(q.Edges) == 0
}

type SecurityPolicies struct {
	PaginatedResource[*SecurityPolicyEdge]
}

type SecurityPolicyEdge struct {
	Node *gqlSecurityPolicy
}

func (q ReadSecurityPolicies) ToModel() []*model.SecurityPolicy {
	return utils.Map[*SecurityPolicyEdge, *model.SecurityPolicy](q.SecurityPolicies.Edges,
		func(edge *SecurityPolicyEdge) *model.SecurityPolicy {
			return edge.Node.ToModel()
		})
}

type SecurityPolicyFilterField struct {
	Name *StringFilterOperationInput `json:"name"`
}

func NewSecurityPolicyFilterField(name, filter string) *SecurityPolicyFilterField {
	return &SecurityPolicyFilterField{
		Name: NewStringFilterOperationInput(name, filter),
	}
}
