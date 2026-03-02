package query

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"

type UpdateGateway struct {
	GatewayEntityResponse `graphql:"gatewayUpdate(id: $id, address: $address, remoteNetworkId: $remoteNetworkId, x509CAId: $x509CAId, sshCAId: $sshCAId)"`
}

func (q UpdateGateway) IsEmpty() bool {
	return q.Entity == nil
}

func (q UpdateGateway) ToModel() *model.Gateway {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
