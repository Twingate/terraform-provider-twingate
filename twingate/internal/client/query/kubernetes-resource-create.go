package query

import (
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type CreateKubernetesResource struct {
	KubernetesResourceEntityResponse `graphql:"kubernetesResourceCreate(name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId)"`
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
}

func (g gqlKubernetesResource) ToModel() *model.KubernetesResource {
	return &model.KubernetesResource{
		ID:              string(g.ID),
		Name:            g.Name,
		Address:         g.Address.Value,
		GatewayID:       string(g.Gateway.ID),
		RemoteNetworkID: string(g.RemoteNetwork.ID),
	}
}
