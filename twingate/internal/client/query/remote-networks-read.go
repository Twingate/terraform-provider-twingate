package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

const CursorRemoteNetworks = "remoteNetworksEndCursor"

type ReadRemoteNetworks struct {
	RemoteNetworks `graphql:"remoteNetworks(filter: $filter, after: $remoteNetworksEndCursor, first: $pageLimit)"`
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

type RemoteNetworkType string

func ConvertNetworkType(exitNodeNetwork bool) RemoteNetworkType {
	if exitNodeNetwork {
		return model.NetworkTypeExit
	}

	return model.NetworkTypeRegular
}

type RemoteNetworkTypeFilterOperatorInput struct {
	In []RemoteNetworkType `json:"in"`
}

type RemoteNetworkFilterInput struct {
	Name        *StringFilterOperationInput           `json:"name"`
	NetworkType *RemoteNetworkTypeFilterOperatorInput `json:"networkType"`
}

func NewRemoteNetworkFilterInput(name, filter string, exitNode bool) *RemoteNetworkFilterInput {
	return &RemoteNetworkFilterInput{
		Name: NewStringFilterOperationInput(name, filter),
		NetworkType: &RemoteNetworkTypeFilterOperatorInput{
			In: []RemoteNetworkType{ConvertNetworkType(exitNode)},
		},
	}
}
