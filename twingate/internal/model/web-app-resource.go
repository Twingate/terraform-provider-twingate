package model

type WebAppResource struct {
	ID                    string
	Name                  string
	Address               string
	GatewayID             string
	RemoteNetworkID       string
	IsVisible             *bool
	Alias                 *string
	SecurityPolicyID      *string
	Tags                  map[string]string
	UpstreamPort          int64
	DownstreamPort        int64
	RequestHeaderRewrites map[string]string
	AccessPolicy          *AccessPolicy
	GroupsAccess          []AccessGroup
}

func (r WebAppResource) GetID() string {
	return r.ID
}

func (r WebAppResource) GetName() string {
	return r.Name
}
