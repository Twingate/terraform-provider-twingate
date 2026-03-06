package query

import "github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"

type CreateGateway struct {
	GatewayEntityResponse `graphql:"gatewayCreate(address: $address, remoteNetworkId: $remoteNetworkId, x509CAId: $x509CAId, sshCAId: $sshCAId)"`
}

type GatewayEntityResponse struct {
	Entity *gqlGateway
	OkError
}

func (q CreateGateway) IsEmpty() bool {
	return q.Entity == nil
}

func (q CreateGateway) ToModel() *model.Gateway {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
