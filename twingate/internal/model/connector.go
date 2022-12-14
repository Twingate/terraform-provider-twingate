package model

type Connector struct {
	ID        string
	Name      string
	NetworkID string
}

func (c Connector) GetName() string {
	return c.Name
}

func (c Connector) GetID() string {
	return c.ID
}

func (c Connector) ToTerraform() interface{} {
	return map[string]interface{}{
		"id":                c.ID,
		"name":              c.Name,
		"remote_network_id": c.NetworkID,
	}
}
