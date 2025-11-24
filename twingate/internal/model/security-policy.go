package model

import "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"

type SecurityPolicy struct {
	ID   string
	Name string
}

func (s SecurityPolicy) ToTerraform() any {
	return map[string]any{
		attr.ID:   s.ID,
		attr.Name: s.Name,
	}
}
