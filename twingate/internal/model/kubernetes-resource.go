package model

type KubernetesResource struct {
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

func (r KubernetesResource) GetID() string {
	return r.ID
}

func (r KubernetesResource) GetName() string {
	return r.Name
}
