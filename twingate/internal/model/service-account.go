package model

type ServiceAccount struct {
	ID   string
	Name string
}

func (s ServiceAccount) GetName() string {
	return s.Name
}

func (s ServiceAccount) GetID() string {
	return s.ID
}

type Service struct {
	ID        string
	Name      string
	Resources []string
	Keys      []string
}

func (s Service) ToTerraform() interface{} {
	return map[string]interface{}{
		"id":           s.ID,
		"name":         s.Name,
		"resource_ids": s.Resources,
		"key_ids":      s.Keys,
	}
}