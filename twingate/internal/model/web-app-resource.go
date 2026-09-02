package model

type WebAppUpstream struct {
	Port int64
	TLS  bool
}

type WebAppDownstream struct {
	Port int64
	TLS  bool
}

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
	Upstream              WebAppUpstream
	Downstream            WebAppDownstream
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
