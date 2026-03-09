package query

import (
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadSSHResource struct {
	Resource *gqlSSHResourceNode `graphql:"resource(id: $id)"`
}

func (q ReadSSHResource) IsEmpty() bool {
	return q.Resource == nil
}

func (q ReadSSHResource) ToModel() *model.SSHResource {
	if q.Resource == nil {
		return nil
	}

	return q.Resource.ToModel()
}

type gqlSSHResourceNode struct {
	IDName
	Address struct {
		Value string
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	SSHResourceFragment struct {
		Gateway struct {
			ID graphql.ID
		}
	} `graphql:"... on SSHResource"`
}

func (n gqlSSHResourceNode) ToModel() *model.SSHResource {
	return &model.SSHResource{
		ID:              string(n.ID),
		Name:            n.Name,
		Address:         n.Address.Value,
		GatewayID:       string(n.SSHResourceFragment.Gateway.ID),
		RemoteNetworkID: string(n.RemoteNetwork.ID),
	}
}
