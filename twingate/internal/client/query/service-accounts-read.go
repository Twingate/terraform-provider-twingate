package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hasura/go-graphql-client"
)

const (
	CursorServices    = "servicesEndCursor"
	CursorServiceKeys = "keysEndCursor"
)

type ReadServiceAccounts struct {
	Services `graphql:"serviceAccounts(filter: $filter, after: $servicesEndCursor, first: $pageLimit)"`
}

func (q ReadServiceAccounts) IsEmpty() bool {
	return len(q.Edges) == 0
}

type Services struct {
	PaginatedResource[*ServiceEdge]
}

func (s *Services) ToModel() []*model.ServiceAccount {
	return utils.Map[*ServiceEdge, *model.ServiceAccount](s.Edges, func(edge *ServiceEdge) *model.ServiceAccount {
		return edge.Node.ToModel()
	})
}

type ServiceEdge struct {
	Node *GqlService
}

type GqlService struct {
	IDName
	Resources gqlResourceIDs `graphql:"resources(after: $resourcesEndCursor, first: $pageLimit)"`
	Keys      gqlKeyIDs      `graphql:"keys(after: $keysEndCursor, first: $pageLimit)"`
}

func (s *GqlService) ToModel() *model.ServiceAccount {
	return &model.ServiceAccount{
		ID:        string(s.ID),
		Name:      s.Name,
		Resources: s.Resources.listIDs(),
		Keys:      s.Keys.listIDs(),
	}
}

type gqlResourceIDs struct {
	PaginatedResource[*GqlResourceIDEdge]
}

func (q gqlResourceIDs) listIDs() []string {
	return utils.Map[*GqlResourceIDEdge, string](q.Edges, func(edge *GqlResourceIDEdge) string {
		return string(edge.Node.ID)
	})
}

type GqlResourceIDEdge struct {
	Node *gqlResourceID
}

type gqlResourceID struct {
	ID       graphql.ID
	IsActive bool
}

func IsGqlResourceActive(item *GqlResourceIDEdge) bool {
	if item == nil || item.Node == nil {
		return false
	}

	return item.Node.IsActive
}

type gqlKeyIDs struct {
	PaginatedResource[*GqlKeyIDEdge]
}

func (q gqlKeyIDs) listIDs() []string {
	return utils.Map[*GqlKeyIDEdge, string](q.Edges, func(edge *GqlKeyIDEdge) string {
		return string(edge.Node.ID)
	})
}

type GqlKeyIDEdge struct {
	Node *gqlKeyID
}

type gqlKeyID struct {
	ID     graphql.ID
	Status string
}

func (k gqlKeyID) isActive() bool {
	return k.Status == model.StatusActive
}

func IsGqlKeyActive(item *GqlKeyIDEdge) bool {
	if item == nil || item.Node == nil {
		return false
	}

	return item.Node.isActive()
}

type ServiceAccountFilterInput struct {
	Name *StringFilterOperationInput `json:"name"`
}

func NewServiceAccountFilterInput(name, filter string) *ServiceAccountFilterInput {
	if name == "" {
		return nil
	}

	return &ServiceAccountFilterInput{
		Name: NewStringFilterOperationInput(name, filter),
	}
}
