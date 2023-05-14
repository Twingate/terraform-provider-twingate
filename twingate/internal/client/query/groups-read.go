package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const (
	CursorGroups    = "groupsEndCursor"
	PageLimitGroups = "groupsPageLimit"
)

type ReadGroups struct {
	Groups `graphql:"groups(filter: $filter, after: $groupsEndCursor, first: $groupsPageLimit)"`
}

func (q ReadGroups) IsEmpty() bool {
	return len(q.Edges) == 0
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
	Eq string `json:"eq"`
}

type GroupTypeFilterOperatorInput struct {
	In []string `json:"in"`
}

type BooleanFilterOperatorInput struct {
	Eq bool `json:"eq"`
}

func NewGroupFilterInput(input *model.GroupsFilter) *GroupFilterInput {
	if input == nil {
		return nil
	}

	// default filter settings
	filter := &GroupFilterInput{
		Type: GroupTypeFilterOperatorInput{
			In: []string{
				model.GroupTypeManual,
				model.GroupTypeSynced,
				model.GroupTypeSystem,
			},
		},
		IsActive: BooleanFilterOperatorInput{Eq: true},
	}

	if input.Name != nil {
		filter.Name = &StringFilterOperationInput{
			Eq: *input.Name,
		}
	}

	if input.Type != nil {
		filter.Type.In = []string{*input.Type}
	}

	if input.IsActive != nil {
		filter.IsActive.Eq = *input.IsActive
	}

	return filter
}
