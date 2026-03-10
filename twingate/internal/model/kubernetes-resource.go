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
