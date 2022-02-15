package twingate

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type Connector struct {
	ID              graphql.ID
	RemoteNetwork   *remoteNetwork
	Name            graphql.String
	ConnectorTokens *connectorTokens
}

type Connectors struct {
	ID   string
	Name string
}

const connectorResourceName = "connector"

type createConnectorQuery struct {
	ConnectorCreate struct {
		Entity IDName
		OkError
	} `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId)"`
}

type updateConnectorQuery struct {
	ConnectorUpdate struct {
		Entity IDName
		OkError
	} `graphql:"connectorUpdate(id: $connectorId, name: $connectorName )"`
}

func (client *Client) createConnector(ctx context.Context, remoteNetworkID graphql.ID) (*Connector, error) {
	if remoteNetworkID.(string) == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "create", connectorResourceName)
	}

	variables := map[string]interface{}{
		"remoteNetworkId": remoteNetworkID,
	}
	response := createConnectorQuery{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("createConnector"))
	if err != nil {
		return nil, NewAPIError(err, "create", connectorResourceName)
	}

	if !response.ConnectorCreate.Ok {
		return nil, NewAPIError(NewMutationError(response.ConnectorCreate.Error), "create", connectorResourceName)
	}

	connector := Connector{
		ID:   response.ConnectorCreate.Entity.ID,
		Name: response.ConnectorCreate.Entity.Name,
	}

	return &connector, nil
}

func (client *Client) updateConnector(ctx context.Context, connectorID graphql.ID, connectorName graphql.String) error {
	if connectorID.(string) == "" {
		return NewAPIError(ErrGraphqlConnectorIDIsEmpty, "update", connectorResourceName)
	}

	variables := map[string]interface{}{
		"connectorId":   connectorID,
		"connectorName": connectorName,
	}
	response := updateConnectorQuery{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("updateConnector"))
	if err != nil {
		return NewAPIErrorWithID(err, "update", connectorResourceName, connectorID)
	}

	if !response.ConnectorUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ConnectorUpdate.Error), "update", connectorResourceName, connectorID)
	}

	return nil
}

type readConnectorsQuery struct { //nolint
	Connectors struct {
		Edges []*Edges
	}
}

func (client *Client) readConnectors(ctx context.Context) (map[int]*Connectors, error) { //nolint
	response := readConnectorsQuery{}

	err := client.GraphqlClient.Query(ctx, &response, nil, graphql.OperationName("readConnectors"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	var connectors = make(map[int]*Connectors)

	for i, elem := range response.Connectors.Edges {
		c := &Connectors{ID: elem.Node.StringID(), Name: elem.Node.StringName()}
		connectors[i] = c
	}

	return connectors, nil
}

type readConnectorQuery struct {
	Connector *struct {
		IDName
		RemoteNetwork IDName
	} `graphql:"connector(id: $id)"`
}

func (client *Client) readConnector(ctx context.Context, connectorID graphql.ID) (*Connector, error) {
	if connectorID.(string) == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", connectorResourceName)
	}

	variables := map[string]interface{}{
		"id": connectorID,
	}

	response := readConnectorQuery{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readConnector"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	if response.Connector == nil {
		return nil, NewAPIErrorWithID(nil, "read", connectorResourceName, connectorID)
	}

	connectorRemoteNetwork := &remoteNetwork{
		ID:   response.Connector.RemoteNetwork.ID,
		Name: response.Connector.RemoteNetwork.Name,
	}

	connector := Connector{
		ID:            response.Connector.ID,
		Name:          response.Connector.Name,
		RemoteNetwork: connectorRemoteNetwork,
	}

	return &connector, nil
}

type deleteConnectorQuery struct {
	ConnectorDelete *OkError `graphql:"connectorDelete(id: $id)" json:"connectorDelete"`
}

func (client *Client) deleteConnector(ctx context.Context, connectorID graphql.ID) error {
	if connectorID.(string) == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", connectorResourceName)
	}

	variables := map[string]interface{}{
		"id": connectorID,
	}

	response := deleteConnectorQuery{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("deleteConnector"))
	if err != nil {
		return NewAPIErrorWithID(err, "delete", connectorResourceName, connectorID)
	}

	if !response.ConnectorDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ConnectorDelete.Error), "delete", connectorResourceName, connectorID)
	}

	return nil
}
