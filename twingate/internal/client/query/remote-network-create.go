package query

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

type CreateRemoteNetwork struct {
	RemoteNetworkEntityResponse `graphql:"remoteNetworkCreate(name: $name, isActive: $isActive, location: $location)"`
}

type RemoteNetworkEntityResponse struct {
	Entity *gqlRemoteNetwork
	OkError
}

func (q CreateRemoteNetwork) ToModel() *model.RemoteNetwork {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}
