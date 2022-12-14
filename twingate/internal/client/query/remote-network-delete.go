package query

type DeleteRemoteNetwork struct {
	OkError `graphql:"remoteNetworkDelete(id: $id)"`
}
