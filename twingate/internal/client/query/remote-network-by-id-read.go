package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

type ReadRemoteNetworkByID struct {
	RemoteNetwork *gqlRemoteNetwork `graphql:"remoteNetwork(id: $id)"`
}

func (r ReadRemoteNetworkByID) IsEmpty() bool {
	return r.RemoteNetwork == nil
}

type gqlRemoteNetwork struct {
	IDName
	Location string
}

func (g gqlRemoteNetwork) ToModel() *model.RemoteNetwork {
	return &model.RemoteNetwork{
		ID:       string(g.ID),
		Name:     g.Name,
		Location: g.Location,
	}
}

func (r ReadRemoteNetworkByID) ToModel() *model.RemoteNetwork {
	if r.RemoteNetwork == nil {
		return nil
	}

	return r.RemoteNetwork.ToModel()
}
