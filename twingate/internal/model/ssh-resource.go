package model

type SSHDownstream struct {
	Port int64
}

type SSHUpstream struct {
	Port int64
}

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
	Downstream       *SSHDownstream
	Upstream         *SSHUpstream
	AccessPolicy     *AccessPolicy
	GroupsAccess     []AccessGroup
}

func (r SSHResource) GetID() string {
	return r.ID
}

func (r SSHResource) GetName() string {
	return r.Name
}
