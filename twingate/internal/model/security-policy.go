package model

type SecurityPolicy struct {
	ID   string
	Name string
}

func (s SecurityPolicy) ToTerraform() interface{} {
	return map[string]interface{}{
		"id":   s.ID,
		"name": s.Name,
	}
}
