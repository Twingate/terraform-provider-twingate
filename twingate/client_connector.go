package twingate

import (
	"context"

	"github.com/twingate/go-graphql-client"
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

func (client *Client) createConnector(ctx context.Context, remoteNetworkID, connectorName string) (*Connector, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "create", connectorResourceName)
	}

	var (
		err      error
		response *ConnectorCreateResponse
	)

	if connectorName == "" {
		response, err = client.createConnectorWithoutName(ctx, remoteNetworkID)
	} else {
		response, err = client.createConnectorWithName(ctx, remoteNetworkID, connectorName)
	}

	if err != nil {
		return nil, err
	}

	if !response.Ok {
		return nil, NewAPIError(NewMutationError(response.Error), "create", connectorResourceName)
	}

	connector := Connector{
		ID:   response.Entity.ID,
		Name: response.Entity.Name,
	}

	return &connector, nil
}

type ConnectorCreateResponse struct {
	Entity IDName
	OkError
}

type createConnectorQuery struct {
	ConnectorCreate *ConnectorCreateResponse `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId)"`
}

func (client *Client) createConnectorWithoutName(ctx context.Context, remoteNetworkID string) (*ConnectorCreateResponse, error) {
	variables := map[string]interface{}{
		"remoteNetworkId": graphql.ID(remoteNetworkID),
	}

	response := createConnectorQuery{}
	err := client.GraphqlClient.NamedMutate(ctx, "createConnector", &response, variables)

	if err != nil {
		return nil, NewAPIError(err, "create", connectorResourceName)
	}

	return response.ConnectorCreate, nil
}

type createConnectorWithNameQuery struct {
	ConnectorCreate *ConnectorCreateResponse `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId, name: $connectorName)"`
}

func (client *Client) createConnectorWithName(ctx context.Context, remoteNetworkID, connectorName string) (*ConnectorCreateResponse, error) {
	variables := map[string]interface{}{
		"remoteNetworkId": graphql.ID(remoteNetworkID),
		"connectorName":   graphql.String(connectorName),
	}

	response := createConnectorWithNameQuery{}
	err := client.GraphqlClient.NamedMutate(ctx, "createConnector", &response, variables)

	if err != nil {
		return nil, NewAPIError(err, "create", connectorResourceName)
	}

	return response.ConnectorCreate, nil
}

type updateConnectorQuery struct {
	ConnectorUpdate struct {
		Entity IDName
		OkError
	} `graphql:"connectorUpdate(id: $connectorId, name: $connectorName )"`
}

func (client *Client) updateConnector(ctx context.Context, connectorID string, connectorName string) error {
	if connectorID == "" {
		return NewAPIError(ErrGraphqlConnectorIDIsEmpty, "update", connectorResourceName)
	}

	variables := map[string]interface{}{
		"connectorId":   graphql.ID(connectorID),
		"connectorName": graphql.String(connectorName),
	}
	response := updateConnectorQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateConnector", &response, variables)
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

	err := client.GraphqlClient.NamedQuery(ctx, "readConnectors", &response, nil)
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

type readConnectorsWithRemoteNetworkQuery struct {
	Connectors struct {
		Edges []*struct {
			Node struct {
				IDName
				RemoteNetwork struct {
					ID graphql.ID
				}
			}
		}
	}
}

func (client *Client) readConnectorsWithRemoteNetwork(ctx context.Context) ([]*Connector, error) {
	response := readConnectorsWithRemoteNetworkQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readConnectors", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	if response.Connectors.Edges == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, "All")
	}

	connectors := make([]*Connector, 0, len(response.Connectors.Edges))

	for _, elem := range response.Connectors.Edges {
		if elem == nil {
			continue
		}

		conn := elem.Node

		connectors = append(connectors, &Connector{
			ID:   conn.ID,
			Name: conn.Name,
			RemoteNetwork: &remoteNetwork{
				ID: conn.RemoteNetwork.ID,
			},
		})
	}

	return connectors, nil
}

type readConnectorQuery struct {
	Connector *struct {
		IDName
		RemoteNetwork IDName
	} `graphql:"connector(id: $id)"`
}

func (client *Client) readConnector(ctx context.Context, connectorID string) (*Connector, error) {
	if connectorID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", connectorResourceName)
	}

	variables := map[string]interface{}{
		"id": graphql.ID(connectorID),
	}

	response := readConnectorQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readConnector", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	if response.Connector == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, connectorID)
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

func (client *Client) deleteConnector(ctx context.Context, connectorID string) error {
	if connectorID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", connectorResourceName)
	}

	variables := map[string]interface{}{
		"id": graphql.ID(connectorID),
	}

	response := deleteConnectorQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "deleteConnector", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", connectorResourceName, connectorID)
	}

	if !response.ConnectorDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ConnectorDelete.Error), "delete", connectorResourceName, connectorID)
	}

	return nil
}
