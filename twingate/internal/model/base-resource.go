package model

// BaseResource holds what every resource kind has in common: the Network resource
// and the PAM resources (Kubernetes, SSH and Web App).
type BaseResource struct {
	ID               string
	Name             string
	Address          string
	RemoteNetworkID  string
	IsVisible        *bool
	Alias            *string
	SecurityPolicyID *string
	Tags             map[string]string
	AccessPolicy     *AccessPolicy
	GroupsAccess     []AccessGroup
	ServiceAccounts  []string
}

func (r BaseResource) GetID() string {
	return r.ID
}

func (r BaseResource) GetName() string {
	return r.Name
}

// PAMResource holds what the PAM resources have in common: they are all
// reached through a Gateway.
type PAMResource struct {
	BaseResource

	GatewayID string
}
