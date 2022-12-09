package query

type ReadRemoteNetworkByName struct {
	RemoteNetworks gqlRemoteNetworks `graphql:"remoteNetworks(filter: {name: {eq: $name}})"`
}

type gqlRemoteNetworks struct {
	Edges []*RemoteNetworkEdge
}
