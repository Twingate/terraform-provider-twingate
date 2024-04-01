package model

import "github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/attr"

const (
	GroupTypeManual = "MANUAL"
	GroupTypeSynced = "SYNCED"
	GroupTypeSystem = "SYSTEM"
)

type Group struct {
	ID               string
	Name             string
	Type             string
	IsActive         bool
	Users            []string
	IsAuthoritative  bool
	SecurityPolicyID string
}

func (g Group) GetName() string {
	return g.Name
}

func (g Group) GetID() string {
	return g.ID
}

func (g Group) ToTerraform() interface{} {
	return map[string]interface{}{
		attr.ID:               g.ID,
		attr.Name:             g.Name,
		attr.Type:             g.Type,
		attr.IsActive:         g.IsActive,
		attr.SecurityPolicyID: g.SecurityPolicyID,
	}
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
