package query

import "github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"

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
	gqlPAMResource

	Downstream SSHDownstream
	Upstream   SSHUpstream
}

func (g gqlSSHResource) ToModel() *model.SSHResource {
	return &model.SSHResource{
		PAMResource: g.gqlPAMResource.ToModel(),
		Downstream:  &model.SSHDownstream{Port: g.Downstream.Port},
		Upstream:    &model.SSHUpstream{Port: g.Upstream.Port},
	}
}
