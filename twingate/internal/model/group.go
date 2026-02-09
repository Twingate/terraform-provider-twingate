package model

import (
	"regexp"
	"slices"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
)

const (
	GroupTypeManual = "MANUAL"
	GroupTypeSynced = "SYNCED"
	GroupTypeSystem = "SYSTEM"
)

type Group struct {
	ID              string
	Name            string
	Type            string
	IsActive        bool
	Users           []string
	IsAuthoritative bool
}

func (g Group) GetName() string {
	return g.Name
}

func (g Group) GetID() string {
	return g.ID
}

func (g Group) ToTerraform() any {
	return map[string]any{
		attr.ID:       g.ID,
		attr.Name:     g.Name,
		attr.Type:     g.Type,
		attr.IsActive: g.IsActive,
	}
}

func (g Group) Match(filter ResourceFilter) bool {
	if filter.IsNil() {
		// matches all groups
		return true
	}

	if filter.HasNotSupportedFilters() {
		// for not supported filters we delegate fetching data from API
		return false
	}

	// filter by isActive
	if filter.GetIsActive() != nil && *filter.GetIsActive() != g.IsActive {
		return false
	}

	// filter by type
	if len(filter.GetTypes()) > 0 && !slices.Contains(filter.GetTypes(), g.Type) {
		return false
	}

	// filter by name
	if name := filter.GetName(); name != "" {
		switch filter.GetFilterBy() {
		case "":
			if g.Name != name {
				return false
			}

		case attr.FilterByContains:
			if !strings.Contains(g.Name, name) {
				return false
			}

		case attr.FilterByExclude:
			if strings.Contains(g.Name, name) {
				return false
			}

		case attr.FilterByPrefix:
			if !strings.HasPrefix(g.Name, name) {
				return false
			}

		case attr.FilterBySuffix:
			if !strings.HasSuffix(g.Name, name) {
				return false
			}

		case attr.FilterByRegexp:
			matched, err := regexp.MatchString(name, g.Name)
			if err != nil || !matched {
				return false
			}
		}
	}

	return true
}

type GroupsFilter struct {
	Name       *string
	NameFilter string
	Types      []string
	IsActive   *bool
}

func (f *GroupsFilter) HasName() bool {
	return f != nil && f.Name != nil && *f.Name != ""
}

func (f *GroupsFilter) GetName() string {
	if f.HasName() {
		return *f.Name
	}

	return ""
}

func (f *GroupsFilter) GetFilterBy() string {
	return f.NameFilter
}

func (f *GroupsFilter) GetTypes() []string {
	return f.Types
}

func (f *GroupsFilter) GetIsActive() *bool {
	return f.IsActive
}

func (f *GroupsFilter) IsNil() bool {
	return f == nil
}

func (f *GroupsFilter) HasNotSupportedFilters() bool {
	return f != nil && !slices.Contains([]string{"", attr.FilterByRegexp, attr.FilterByContains, attr.FilterByExclude, attr.FilterByPrefix, attr.FilterBySuffix}, f.NameFilter)
}

func (f *GroupsFilter) GetTags() map[string]string {
	// not supported
	return nil
}
