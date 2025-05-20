package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

const CursorDLPPolicies = "dlpPoliciesEndCursor"

type ReadDLPPolicies struct {
	DLPPolicies `graphql:"dlpPolicies(filter: $filter, after: $dlpPoliciesEndCursor, first: $pageLimit)"`
}

func (q ReadDLPPolicies) IsEmpty() bool {
	return len(q.Edges) == 0
}

type DLPPolicies struct {
	PaginatedResource[*DLPPolicyEdge]
}

type DLPPolicyEdge struct {
	Node *gqlDLPPolicy
}

func (q ReadDLPPolicies) ToModel() []*model.DLPPolicy {
	return utils.Map(q.Edges,
		func(edge *DLPPolicyEdge) *model.DLPPolicy {
			return edge.Node.ToModel()
		})
}

type DLPPolicyFilterField struct {
	Name *StringFilterOperationInput `json:"name"`
}

func NewDLPPolicyFilterField(name, filter string) *DLPPolicyFilterField {
	return &DLPPolicyFilterField{
		Name: NewStringFilterOperationInput(name, filter),
	}
}