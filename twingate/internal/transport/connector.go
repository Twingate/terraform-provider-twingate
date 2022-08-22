package transport

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

func (c Connectors) GetName() string {
	return c.Name
}

func (c Connectors) GetID() string {
	return c.ID
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

func (client *Client) CreateConnector(ctx context.Context, remoteNetworkID string) (*Connector, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "create", connectorResourceName)
	}

	variables := map[string]interface{}{
		"remoteNetworkId": graphql.ID(remoteNetworkID),
	}
	response := createConnectorQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "createConnector", &response, variables)
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

func (client *Client) UpdateConnector(ctx context.Context, connectorID string, connectorName string) error {
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

func (client *Client) ReadConnectors(ctx context.Context) (map[int]*Connectors, error) { //nolint
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

func (client *Client) ReadConnectorsWithRemoteNetwork(ctx context.Context) ([]*Connector, error) {
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

func (client *Client) ReadConnector(ctx context.Context, connectorID string) (*Connector, error) {
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

func (client *Client) DeleteConnector(ctx context.Context, connectorID string) error {
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
