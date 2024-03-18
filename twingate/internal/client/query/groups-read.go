package query

import (
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/utils"
)

const CursorGroups = "groupsEndCursor"

type ReadGroups struct {
	Groups `graphql:"groups(filter: $filter, after: $groupsEndCursor, first: $pageLimit)"`
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
	Eq         *string  `json:"eq"`
	Ne         *string  `json:"ne"`
	StartsWith *string  `json:"startsWith"`
	EndsWith   *string  `json:"endsWith"`
	Regexp     *string  `json:"regexp"`
	Contains   *string  `json:"contains"`
	In         []string `json:"in"`
}

func NewStringFilterOperationInput(name, filter string) *StringFilterOperationInput {
	if filter == "" && name == "" {
		return nil
	}

	var stringFilter StringFilterOperationInput

	switch filter {
	case attr.FilterByRegexp:
		stringFilter.Regexp = &name
	case attr.FilterByContains:
		stringFilter.Contains = &name
	case attr.FilterByExclude:
		stringFilter.Ne = &name
	case attr.FilterByPrefix:
		stringFilter.StartsWith = &name
	case attr.FilterBySuffix:
		stringFilter.EndsWith = &name
	default:
		stringFilter.Eq = &name
	}

	return &stringFilter
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
		filter.Name = NewStringFilterOperationInput(*input.Name, input.NameFilter)
	}

	if len(input.Types) > 0 {
		filter.Type.In = input.Types
	}

	if input.IsActive != nil {
		filter.IsActive.Eq = *input.IsActive
	}

	return filter
}
