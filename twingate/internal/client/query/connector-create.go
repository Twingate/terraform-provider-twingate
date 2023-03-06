package query

type CreateConnector struct {
	ConnectorEntityResponse `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId, name: $connectorName, hasStatusNotificationsEnabled: $hasStatusNotificationsEnabled)"`
}

type ConnectorEntityResponse struct {
	Entity *gqlConnector
	OkError
}
