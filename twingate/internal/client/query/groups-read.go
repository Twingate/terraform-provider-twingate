package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/twingate/go-graphql-client"
)

const CursorGroups = "groupsEndCursor"

type ReadGroups struct {
	Groups `graphql:"groups(filter: $filter, after: $groupsEndCursor)"`
}

type Groups struct {
	PaginatedResource[*GroupEdge]
}

type GroupEdge struct {
	Node *gqlGroup
}

func (u Groups) ToModel() []*model.Group {
	return utils.Map[*GroupEdge, *model.Group](u.Edges, func(edge *GroupEdge) *model.Group {
		return edge.Node.ToModel()
	})
}

type GroupFilterInput struct {
	Name     *StringFilterOperationInput  `json:"name"`
	Type     GroupTypeFilterOperatorInput `json:"type"`
	IsActive BooleanFilterOperatorInput   `json:"isActive"`
}

type StringFilterOperationInput struct {
	Eq graphql.String `json:"eq"`
}

type GroupTypeFilterOperatorInput struct {
	In []graphql.String `json:"in"`
}

type BooleanFilterOperatorInput struct {
	Eq graphql.Boolean `json:"eq"`
}

func NewGroupFilterInput(input *model.GroupsFilter) *GroupFilterInput {
	if input == nil {
		return nil
	}

	// default filter settings
	filter := &GroupFilterInput{
		Type: GroupTypeFilterOperatorInput{
			In: []graphql.String{
				model.GroupTypeManual,
				model.GroupTypeSynced,
				model.GroupTypeSystem,
			},
		},
		IsActive: BooleanFilterOperatorInput{Eq: true},
	}

	if input.Name != nil {
		filter.Name = &StringFilterOperationInput{
			Eq: graphql.String(*input.Name),
		}
	}

	if input.Type != nil {
		filter.Type.In = []graphql.String{graphql.String(*input.Type)}
	}

	if input.IsActive != nil {
		filter.IsActive.Eq = graphql.Boolean(*input.IsActive)
	}

	return filter
}
