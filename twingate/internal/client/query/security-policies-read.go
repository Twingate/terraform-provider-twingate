package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const (
	CursorPolicies    = "policiesEndCursor"
	PageLimitPolicies = "policiesPageLimit"
)

type ReadSecurityPolicies struct {
	SecurityPolicies `graphql:"securityPolicies(after: $policiesEndCursor, first: $policiesPageLimit)"`
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
