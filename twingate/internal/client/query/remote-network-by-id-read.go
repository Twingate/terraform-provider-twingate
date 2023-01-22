package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadRemoteNetworkByID struct {
	RemoteNetwork *gqlRemoteNetwork `graphql:"remoteNetwork(id: $id)"`
}

type gqlRemoteNetwork struct {
	IDName
	Location graphql.String
}

func (g gqlRemoteNetwork) ToModel() *model.RemoteNetwork {
	return &model.RemoteNetwork{
		ID:       g.StringID(),
		Name:     g.StringName(),
		Location: string(g.Location),
	}
}

func (r ReadRemoteNetworkByID) ToModel() *model.RemoteNetwork {
	if r.RemoteNetwork == nil {
		return nil
	}

	return r.RemoteNetwork.ToModel()
}
