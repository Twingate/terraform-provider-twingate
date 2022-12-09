package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const CursorServiceAccounts = "serviceAccountsEndCursor"

type ReadServiceAccounts struct {
	ServiceAccounts `graphql:"serviceAccounts(after: $serviceAccountsEndCursor)"`
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
