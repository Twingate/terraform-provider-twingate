package query

type DeleteConnector struct {
	OkError `graphql:"connectorDelete(id: $id)"`
}
