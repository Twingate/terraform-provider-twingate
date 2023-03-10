package model

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"

type Connector struct {
	ID                   string
	Name                 string
	NetworkID            string
	StatusUpdatesEnabled *bool
}

func (c Connector) GetName() string {
	return c.Name
}

func (c Connector) GetID() string {
	return c.ID
}

func (c Connector) ToTerraform() interface{} {
	return map[string]interface{}{
		attr.ID:                   c.ID,
		attr.Name:                 c.Name,
		attr.RemoteNetworkID:      c.NetworkID,
		attr.StatusUpdatesEnabled: *c.StatusUpdatesEnabled,
	}
}
