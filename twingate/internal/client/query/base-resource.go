package query

import (
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

// gqlBaseResource holds the fields every resource kind returns.
type gqlBaseResource struct {
	IDName
	Address struct {
		Value string
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	IsVisible      bool
	Alias          string
	SecurityPolicy *gqlSecurityPolicy
	Tags           []Tag
	ApprovalMode   string
	AccessPolicy   *AccessPolicy
}

func (g gqlBaseResource) ToModel() model.BaseResource {
	return model.BaseResource{
		ID:               string(g.ID),
		Name:             g.Name,
		Address:          g.Address.Value,
		RemoteNetworkID:  string(g.RemoteNetwork.ID),
		IsVisible:        &g.IsVisible,
		Alias:            optionalString(g.Alias),
		SecurityPolicyID: securityPolicyID(g.SecurityPolicy),
		Tags:             tagsToModel(g.Tags),
		AccessPolicy:     accessPolicyToModel(g.AccessPolicy, &g.ApprovalMode),
	}
}

// gqlPAMResource is the entity the create and update mutations of the PAM
// resources return: the base fields plus the Gateway.
type gqlPAMResource struct {
	gqlBaseResource

	Gateway struct {
		ID graphql.ID
	}
}

func (g gqlPAMResource) ToModel() model.PAMResource {
	return model.PAMResource{
		BaseResource: g.gqlBaseResource.ToModel(),
		GatewayID:    string(g.Gateway.ID),
	}
}

// toPAMModel converts a resource node of the read query. The Gateway and the
// access live outside of the base fields there: the Gateway in a type fragment,
// the access in a paginated connection.
func (g gqlBaseResource) toPAMModel(gatewayID graphql.ID, access Access) (model.PAMResource, error) {
	res := model.PAMResource{
		BaseResource: g.ToModel(),
		GatewayID:    string(gatewayID),
	}

	groups, serviceAccounts, err := access.ToModel()
	if err != nil {
		return model.PAMResource{}, err
	}

	res.GroupsAccess = groups
	res.ServiceAccounts = serviceAccounts

	return res, nil
}

func securityPolicyID(policy *gqlSecurityPolicy) *string {
	if policy == nil {
		return nil
	}

	id := string(policy.ID)
	if id == "" {
		return nil
	}

	return &id
}
