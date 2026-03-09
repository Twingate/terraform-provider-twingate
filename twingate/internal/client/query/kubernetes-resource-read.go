package query

import (
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadKubernetesResource struct {
	Resource *gqlKubernetesResourceNode `graphql:"resource(id: $id)"`
}

func (q ReadKubernetesResource) IsEmpty() bool {
	return q.Resource == nil
}

func (q ReadKubernetesResource) ToModel() *model.KubernetesResource {
	if q.Resource == nil {
		return nil
	}

	return q.Resource.ToModel()
}

type gqlKubernetesResourceNode struct {
	IDName
	Address struct {
		Value string
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	KubernetesResourceFragment struct {
		Gateway struct {
			ID graphql.ID
		}
	} `graphql:"... on KubernetesResource"`
}

func (n gqlKubernetesResourceNode) ToModel() *model.KubernetesResource {
	return &model.KubernetesResource{
		ID:              string(n.ID),
		Name:            n.Name,
		Address:         n.Address.Value,
		GatewayID:       string(n.KubernetesResourceFragment.Gateway.ID),
		RemoteNetworkID: string(n.RemoteNetwork.ID),
	}
}
