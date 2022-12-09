package query

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

type UpdateRemoteNetwork struct {
	RemoteNetworkEntityResponse `graphql:"remoteNetworkUpdate(id: $id, name: $name)"`
}

func (q UpdateRemoteNetwork) ToModel() *model.RemoteNetwork {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
