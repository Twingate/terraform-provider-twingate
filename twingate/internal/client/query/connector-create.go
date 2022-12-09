package query

type CreateConnector struct {
	ConnectorEntityResponse `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId, name: $connectorName)"`
}

type ConnectorEntityResponse struct {
	Entity *gqlConnector
	OkError
}
