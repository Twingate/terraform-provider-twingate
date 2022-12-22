package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const CursorPolicies = "policiesEndCursor"

type ReadSecurityPolicies struct {
	SecurityPolicies `graphql:"securityPolicies(after: $policiesEndCursor)"`
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
