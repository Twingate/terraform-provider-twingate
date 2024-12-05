package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

type ReadRemoteNetworkByID struct {
	RemoteNetwork *gqlRemoteNetwork `graphql:"remoteNetwork(id: $id)"`
}

func (r ReadRemoteNetworkByID) IsEmpty() bool {
	return r.RemoteNetwork == nil
}

type gqlRemoteNetwork struct {
	IDName
	Location    string
	NetworkType string
}

func (g gqlRemoteNetwork) ToModel() *model.RemoteNetwork {
	return &model.RemoteNetwork{
		ID:       string(g.ID),
		Name:     g.Name,
		Location: g.Location,
		ExitNode: g.NetworkType == model.NetworkTypeExit,
	}
}

func (r ReadRemoteNetworkByID) ToModel() *model.RemoteNetwork {
	if r.RemoteNetwork == nil {
		return nil
	}

	return r.RemoteNetwork.ToModel()
}
