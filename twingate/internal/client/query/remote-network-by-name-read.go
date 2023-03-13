package query

type ReadRemoteNetworkByName struct {
	RemoteNetworks gqlRemoteNetworks `graphql:"remoteNetworks(filter: {name: {eq: $name}})"`
}

func (q ReadRemoteNetworkByName) IsEmpty() bool {
	return len(q.RemoteNetworks.Edges) == 0 || q.RemoteNetworks.Edges[0] == nil
}

type gqlRemoteNetworks struct {
	Edges []*RemoteNetworkEdge
}
