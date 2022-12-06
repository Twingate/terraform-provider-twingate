package model

type Group struct {
	ID       string
	Name     string
	Type     string
	IsActive bool
	UserIDs  []string
}

func (g Group) GetName() string {
	return g.Name
}

func (g Group) GetID() string {
	return g.ID
}

func (g Group) ToTerraform() interface{} {
	return map[string]interface{}{
		"id":        g.ID,
		"name":      g.Name,
		"type":      g.Type,
		"is_active": g.IsActive,
	}
}
