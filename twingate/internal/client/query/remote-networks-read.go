package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const CursorRemoteNetworks = "remoteNetworksEndCursor"

type ReadRemoteNetworks struct {
	RemoteNetworks `graphql:"remoteNetworks(after: $remoteNetworksEndCursor, first: $pageLimit)"`
}

func (q ReadRemoteNetworks) IsEmpty() bool {
	return len(q.Edges) == 0
}

type RemoteNetworks struct {
	PaginatedResource[*RemoteNetworkEdge]
}

type RemoteNetworkEdge struct {
	Node gqlRemoteNetwork
}

func (r RemoteNetworks) ToModel() []*model.RemoteNetwork {
	return utils.Map[*RemoteNetworkEdge, *model.RemoteNetwork](r.Edges, func(edge *RemoteNetworkEdge) *model.RemoteNetwork {
		return edge.Node.ToModel()
	})
}
