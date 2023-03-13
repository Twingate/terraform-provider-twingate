package model

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
		"id":                 g.ID,
		"name":               g.Name,
		"type":               g.Type,
		"is_active":          g.IsActive,
		"security_policy_id": g.SecurityPolicyID,
	}
}

type GroupsFilter struct {
	Name     *string
	Type     *string
	IsActive *bool
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
