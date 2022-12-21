package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/twingate/go-graphql-client"
)

const (
	CursorServices         = "servicesEndCursor"
	CursorServiceResources = "resourcesEndCursor"
	CursorServiceKeys      = "keysEndCursor"
)

type ReadServiceAccounts struct {
	Services `graphql:"serviceAccounts(filter: $filter, after: $servicesEndCursor)"`
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
	Resources gqlResourceIDs `graphql:"resources(after: $resourcesEndCursor)"`
	Keys      gqlKeyIDs      `graphql:"keys(after: $keysEndCursor)"`
}

func (s *GqlService) ToModel() *model.ServiceAccount {
	return &model.ServiceAccount{
		ID:        s.StringID(),
		Name:      s.StringName(),
		Resources: s.Resources.listIDs(),
		Keys:      s.Keys.listIDs(),
	}
}

type gqlResourceIDs struct {
	PaginatedResource[*GqlResourceIDEdge]
}

func (q gqlResourceIDs) listIDs() []string {
	return utils.Map[*GqlResourceIDEdge, string](q.Edges, func(edge *GqlResourceIDEdge) string {
		return edge.Node.ID.(string)
	})
}

type GqlResourceIDEdge struct {
	Node *gqlResourceID
}

type gqlResourceID struct {
	ID       graphql.ID
	IsActive graphql.Boolean
}

func (r gqlResourceID) isActive() bool {
	return bool(r.IsActive)
}

func IsGqlResourceActive(item *GqlResourceIDEdge) bool {
	return item.Node.isActive()
}

type gqlKeyIDs struct {
	PaginatedResource[*GqlKeyIDEdge]
}

func (q gqlKeyIDs) listIDs() []string {
	return utils.Map[*GqlKeyIDEdge, string](q.Edges, func(edge *GqlKeyIDEdge) string {
		return edge.Node.ID.(string)
	})
}

type GqlKeyIDEdge struct {
	Node *gqlKeyID
}

type gqlKeyID struct {
	ID     graphql.ID
	Status graphql.String
}

func (k gqlKeyID) isActive() bool {
	return string(k.Status) == model.StatusActive
}

func IsGqlKeyActive(item *GqlKeyIDEdge) bool {
	return item.Node.isActive()
}

type ServiceAccountFilterInput struct {
	Name StringFilter `json:"name"`
}

type StringFilter struct {
	Eq graphql.String `json:"eq"`
}

func NewServiceAccountFilterInput(name string) *ServiceAccountFilterInput {
	if name == "" {
		return nil
	}

	return &ServiceAccountFilterInput{
		Name: StringFilter{
			Eq: graphql.String(name),
		},
	}
}
