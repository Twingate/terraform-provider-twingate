package query

type DeleteRemoteNetwork struct {
	OkError `graphql:"remoteNetworkDelete(id: $id)"`
}

func (q DeleteRemoteNetwork) IsEmpty() bool {
	return false
}
