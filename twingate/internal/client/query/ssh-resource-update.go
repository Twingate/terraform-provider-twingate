package query

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"

//nolint:lll
type UpdateSSHResource struct {
	SSHResourceEntityResponse `graphql:"sshResourceUpdate(id: $id, name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId, isVisible: $isVisible, alias: $alias, securityPolicyId: $securityPolicyId, tags: $tags, protocols: $protocols, accessPolicy: $accessPolicy, approvalMode: $approvalMode)"`
}

func (q UpdateSSHResource) IsEmpty() bool {
	return q.Entity == nil
}

func (q UpdateSSHResource) ToModel() *model.SSHResource {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
