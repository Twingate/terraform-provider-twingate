package query

type UpdateConnector struct {
	ConnectorEntityResponse `graphql:"connectorUpdate(id: $connectorId, name: $connectorName, hasStatusNotificationsEnabled: $hasStatusNotificationsEnabled)"`
}
