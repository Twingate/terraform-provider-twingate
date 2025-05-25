package model

import "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"

type DLPPolicy struct {
	ID   string
	Name string
}

func (d DLPPolicy) ToTerraform() interface{} {
	return map[string]interface{}{
		attr.ID:   d.ID,
		attr.Name: d.Name,
	}
}