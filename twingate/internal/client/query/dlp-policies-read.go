package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

type DataLossPreventionPolicyFilterInput struct {
	Name *StringFilterOperationInput `json:"name"`
}

func NewDLPPoliciesFilterInput(name, filter string) *DataLossPreventionPolicyFilterInput {
	return &DataLossPreventionPolicyFilterInput{
		Name: NewStringFilterOperationInput(name, filter),
	}
}

const CursorDLPPolicies = "dlpPoliciesEndCursor"

type ReadDLPPolicies struct {
	DLPPolicies `graphql:"dlpPolicies(filter: $filter, after: $dlpPoliciesEndCursor, first: $pageLimit)"`
}

type DLPPolicies struct {
	PaginatedResource[*DLPPolicyEdge]
}

type DLPPolicyEdge struct {
	Node *gqlDLPPolicy
}

func (r ReadDLPPolicies) ToModel() []*model.DLPPolicy {
	return utils.Map[*DLPPolicyEdge, *model.DLPPolicy](r.Edges, func(edge *DLPPolicyEdge) *model.DLPPolicy {
		return edge.Node.ToModel()
	})
}

func (r ReadDLPPolicies) IsEmpty() bool {
	return len(r.DLPPolicies.Edges) == 0
}
