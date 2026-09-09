package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGroupModel(t *testing.T) {
	cases := []struct {
		group model.Group

		expectedName string
		expectedID   string
		expected     any
	}{
		{
			group: model.Group{},
			expected: map[string]any{
				attr.ID:       "",
				attr.Name:     "",
				attr.Type:     "",
				attr.IsActive: false,
			},
		},
		{
			group: model.Group{
				ID:       "id",
				Name:     "name",
				Type:     "type",
				IsActive: true,
			},
			expectedID:   "id",
			expectedName: "name",
			expected: map[string]any{
				attr.ID:       "id",
				attr.Name:     "name",
				attr.Type:     "type",
				attr.IsActive: true,
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expectedID, c.group.GetID())
			assert.Equal(t, c.expectedName, c.group.GetName())
			assert.Equal(t, c.expected, c.group.ToTerraform())
		})
	}
}

func TestGroupMatchByNameIn(t *testing.T) {
	group := model.Group{ID: "id", Name: "Group A", Type: model.GroupTypeManual, IsActive: true}

	cases := []struct {
		filter   *model.GroupsFilter
		expected bool
	}{
		{
			filter:   nil,
			expected: true,
		},
		{
			filter:   &model.GroupsFilter{NameIn: []string{"Group A"}, NameFilter: attr.FilterByIn},
			expected: true,
		},
		{
			filter:   &model.GroupsFilter{NameIn: []string{"Group B", "Group A"}, NameFilter: attr.FilterByIn},
			expected: true,
		},
		{
			filter:   &model.GroupsFilter{NameIn: []string{"Group B"}, NameFilter: attr.FilterByIn},
			expected: false,
		},
		{
			filter:   &model.GroupsFilter{NameIn: []string{"group a"}, NameFilter: attr.FilterByIn},
			expected: false,
		},
		{
			filter: &model.GroupsFilter{
				NameIn:     []string{"Group A"},
				NameFilter: attr.FilterByIn,
				Types:      []string{model.GroupTypeSynced},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.False(t, c.filter.HasNotSupportedFilters())
			assert.Equal(t, c.expected, group.Match(c.filter))
		})
	}
}

func TestGroupsFilterWithNameInString(t *testing.T) {
	filter := &model.GroupsFilter{
		NameIn:     []string{"Group A", "Group B"},
		NameFilter: attr.FilterByIn,
		IsActive:   optionalBool(true),
	}

	assert.Equal(t, `GroupsFilter{Name(in)=[Group A Group B], IsActive=true}`, filter.String())
}

func optionalBool(val bool) *bool {
	return &val
}
