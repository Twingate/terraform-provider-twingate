package model

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"

type ServiceAccount struct {
	ID        string
	Name      string
	Resources []string
	Keys      []string
}

func (s ServiceAccount) GetID() string {
	return s.ID
}

func (s ServiceAccount) GetName() string {
	return s.Name
}

func (s ServiceAccount) ToTerraform() interface{} {
	return map[string]interface{}{
		attr.ID:          s.ID,
		attr.Name:        s.Name,
		attr.ResourceIDs: s.Resources,
		attr.KeyIDs:      s.Keys,
	}
}
