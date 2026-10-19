package model

const (
	TLSClientModeVerifyFull = "VERIFY_FULL"
	TLSClientModeVerifyCA   = "VERIFY_CA"
	TLSClientModeInsecure   = "INSECURE"
	TLSClientModeNone       = "NONE"

	TLSServerModeTLS13 = "TLS13"
	TLSServerModeNone  = "NONE"
)

var TLSClientModes = []string{TLSClientModeVerifyFull, TLSClientModeVerifyCA, TLSClientModeInsecure, TLSClientModeNone} //nolint

var TLSServerModes = []string{TLSServerModeTLS13, TLSServerModeNone} //nolint

type WebAppUpstream struct {
	Port    int64
	TLSMode string
}

type WebAppDownstream struct {
	Port    int64
	TLSMode string
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
