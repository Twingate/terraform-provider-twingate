package model

import "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"

type Connector struct {
	ID                   string
	Name                 string
	NetworkID            string
	StatusUpdatesEnabled *bool
	State                string
	Version              string
	Hostname             string
	PublicIP             string
	PrivateIPs           []string
}

func (c Connector) GetName() string {
	return c.Name
}

func (c Connector) GetID() string {
	return c.ID
}

func (c Connector) ToTerraform() any {
	return map[string]any{
		attr.ID:                   c.ID,
		attr.Name:                 c.Name,
		attr.RemoteNetworkID:      c.NetworkID,
		attr.StatusUpdatesEnabled: *c.StatusUpdatesEnabled,
		attr.State:                c.State,
		attr.Version:              c.Version,
		attr.Hostname:             c.Hostname,
		attr.PublicIP:             c.PublicIP,
		attr.PrivateIPs:           c.PrivateIPs,
	}
}
