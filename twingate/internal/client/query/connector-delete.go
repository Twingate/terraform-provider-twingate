package query

type DeleteConnector struct {
	OkError `graphql:"connectorDelete(id: $id)"`
}

func (r *DeleteConnector) IsEmpty() bool {
	return false
}
