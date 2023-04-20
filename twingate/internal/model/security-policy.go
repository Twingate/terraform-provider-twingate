package model

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"

type SecurityPolicy struct {
	ID   string
	Name string
}

func (s SecurityPolicy) ToTerraform() interface{} {
	return map[string]interface{}{
		attr.ID:   s.ID,
		attr.Name: s.Name,
	}
}
