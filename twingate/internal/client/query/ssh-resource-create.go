package query

import (
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type CreateSSHResource struct {
	SSHResourceEntityResponse `graphql:"sshResourceCreate(name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId)"`
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
}

func (g gqlSSHResource) ToModel() *model.SSHResource {
	return &model.SSHResource{
		ID:              string(g.ID),
		Name:            g.Name,
		Address:         g.Address.Value,
		GatewayID:       string(g.Gateway.ID),
		RemoteNetworkID: string(g.RemoteNetwork.ID),
	}
}
