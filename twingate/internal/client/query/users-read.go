package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const CursorUsers = "usersEndCursor"

type ReadUsers struct {
	Users `graphql:"users(filter: $filter, after: $usersEndCursor, first: $pageLimit)"`
}

func (q ReadUsers) IsEmpty() bool {
	return len(q.Edges) == 0
}

type Users struct {
	PaginatedResource[*UserEdge]
}

type UserEdge struct {
	Node *gqlUser
}

func (u Users) ToModel() []*model.User {
	return utils.Map[*UserEdge, *model.User](u.Edges, func(edge *UserEdge) *model.User {
		return edge.Node.ToModel()
	})
}

type UserFilterInput struct {
	FirstName *StringFilterOperationInput   `json:"firstName"`
	LastName  *StringFilterOperationInput   `json:"lastName"`
	Email     *StringFilterOperationInput   `json:"email"`
	Role      *UserRoleFilterOperationInput `json:"role"`
}

type UserRoleFilterOperationInput struct {
	In []UserRole `json:"in"`
}
