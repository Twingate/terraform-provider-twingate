package query

import (
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

//nolint:lll
type CreateSSHResource struct {
	SSHResourceEntityResponse `graphql:"sshResourceCreate(name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId, isVisible: $isVisible, alias: $alias, securityPolicyId: $securityPolicyId, tags: $tags, downstream: $downstream, upstream: $upstream, accessPolicy: $accessPolicy, approvalMode: $approvalMode)"`
}

func (q CreateSSHResource) IsEmpty() bool {
	return q.Entity == nil
}

func (q CreateSSHResource) ToModel() *model.SSHResource {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}

type SSHResourceEntityResponse struct {
	Entity *gqlSSHResource
	OkError
}

type SSHDownstream struct {
	Port int64
}

type SSHUpstream struct {
	Port int64
}

type gqlSSHResource struct {
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
	IsVisible      bool
	Alias          string
	SecurityPolicy *gqlSecurityPolicy
	Tags           []Tag
	Downstream     SSHDownstream
	Upstream       SSHUpstream
	ApprovalMode   string
	AccessPolicy   *AccessPolicy
}

func (g gqlSSHResource) ToModel() *model.SSHResource {
	return &model.SSHResource{
		ID:               string(g.ID),
		Name:             g.Name,
		Address:          g.Address.Value,
		GatewayID:        string(g.Gateway.ID),
		RemoteNetworkID:  string(g.RemoteNetwork.ID),
		IsVisible:        &g.IsVisible,
		Alias:            optionalString(g.Alias),
		SecurityPolicyID: securityPolicyID(g.SecurityPolicy),
		Tags:             tagsToModel(g.Tags),
		Downstream:       &model.SSHDownstream{Port: g.Downstream.Port},
		Upstream:         &model.SSHUpstream{Port: g.Upstream.Port},
		AccessPolicy:     accessPolicyToModel(g.AccessPolicy, &g.ApprovalMode),
	}
}
