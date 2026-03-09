package query

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"

type UpdateSSHResource struct {
	SSHResourceEntityResponse `graphql:"sshResourceUpdate(id: $id, name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId)"`
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
