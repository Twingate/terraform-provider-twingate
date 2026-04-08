package model

type SSHResource struct {
	ID               string
	Name             string
	Address          string
	GatewayID        string
	RemoteNetworkID  string
	IsVisible        *bool
	Alias            *string
	SecurityPolicyID *string
	Tags             map[string]string
	Protocols        *Protocols
	AccessPolicy     *AccessPolicy
	GroupsAccess     []AccessGroup
}

func (r SSHResource) GetID() string {
	return r.ID
}

func (r SSHResource) GetName() string {
	return r.Name
}
