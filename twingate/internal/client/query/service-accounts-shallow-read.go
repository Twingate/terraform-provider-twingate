package query

import (
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/utils"
)

type ReadShallowServiceAccounts struct {
	ServiceAccounts `graphql:"serviceAccounts(after: $servicesEndCursor, first: $pageLimit)"`
}

func (q ReadShallowServiceAccounts) IsEmpty() bool {
	return len(q.Edges) == 0
}

type ServiceAccounts struct {
	PaginatedResource[*ServiceAccountEdge]
}

type ServiceAccountEdge struct {
	Node *gqlServiceAccount
}

func (s ServiceAccounts) ToModel() []*model.ServiceAccount {
	return utils.Map[*ServiceAccountEdge, *model.ServiceAccount](s.Edges, func(edge *ServiceAccountEdge) *model.ServiceAccount {
		return edge.Node.ToModel()
	})
}
