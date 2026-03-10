package query

import (
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

//nolint:lll
type CreateKubernetesResource struct {
	KubernetesResourceEntityResponse `graphql:"kubernetesResourceCreate(name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId, isVisible: $isVisible, alias: $alias, securityPolicyId: $securityPolicyId, tags: $tags, protocols: $protocols, accessPolicy: $accessPolicy, approvalMode: $approvalMode)"`
}

func (q CreateKubernetesResource) IsEmpty() bool {
	return q.Entity == nil
}

func (q CreateKubernetesResource) ToModel() *model.KubernetesResource {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}

type KubernetesResourceEntityResponse struct {
	Entity *gqlKubernetesResource
	OkError
}

type gqlKubernetesResource struct {
	IDName
	Address struct {
		Value string
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	Gateway struct {
		ID graphql.ID
	}
	Protocols      *Protocols
	IsVisible      bool
	Alias          string
	SecurityPolicy *gqlSecurityPolicy
	Tags           []Tag
	ApprovalMode   string
	AccessPolicy   *AccessPolicy
}

func (g gqlKubernetesResource) ToModel() *model.KubernetesResource {
	return &model.KubernetesResource{
		ID:               string(g.ID),
		Name:             g.Name,
		Address:          g.Address.Value,
		GatewayID:        string(g.Gateway.ID),
		RemoteNetworkID:  string(g.RemoteNetwork.ID),
		IsVisible:        &g.IsVisible,
		Alias:            optionalString(g.Alias),
		SecurityPolicyID: securityPolicyID(g.SecurityPolicy),
		Tags:             tagsToModel(g.Tags),
		Protocols:        protocolsToModel(g.Protocols),
		AccessPolicy:     accessPolicyToModel(g.AccessPolicy, &g.ApprovalMode),
	}
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
