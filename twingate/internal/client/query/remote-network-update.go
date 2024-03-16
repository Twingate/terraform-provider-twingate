package query

import "github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"

type UpdateRemoteNetwork struct {
	RemoteNetworkEntityResponse `graphql:"remoteNetworkUpdate(id: $id, name: $name, location: $location)"`
}

func (q UpdateRemoteNetwork) IsEmpty() bool {
	return q.Entity == nil
}

func (q UpdateRemoteNetwork) ToModel() *model.RemoteNetwork {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
