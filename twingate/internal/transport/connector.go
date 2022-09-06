package transport

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const connectorResourceName = "connector"

type gqlConnector struct {
	IDName
	RemoteNetwork struct {
		ID graphql.ID
	}
}

type createConnectorQuery struct {
	ConnectorCreate struct {
		Entity gqlConnector
		OkError
	} `graphql:"connectorCreate(remoteNetworkId: $remoteNetworkId)"`
}

func (client *Client) CreateConnector(ctx context.Context, remoteNetworkID string) (*model.Connector, error) {
	if remoteNetworkID == "" {
		return nil, NewAPIError(ErrGraphqlNetworkIDIsEmpty, "create", connectorResourceName)
	}

	variables := newVariables().
		withID(remoteNetworkID, "remoteNetworkId").value()

	response := createConnectorQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "createConnector", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", connectorResourceName)
	}

	if !response.ConnectorCreate.Ok {
		return nil, NewAPIError(NewMutationError(response.ConnectorCreate.Error), "create", connectorResourceName)
	}

	return response.ToModel(), nil
}

type updateConnectorQuery struct {
	ConnectorUpdate struct {
		Entity IDName
		OkError
	} `graphql:"connectorUpdate(id: $connectorId, name: $connectorName )"`
}

func (client *Client) UpdateConnector(ctx context.Context, connectorID string, connectorName string) error {
	if connectorID == "" {
		return NewAPIError(ErrGraphqlConnectorIDIsEmpty, "update", connectorResourceName)
	}

	variables := newVariables().
		withID(connectorID, "connectorId").
		withField(connectorName, "connectorName").value()

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

type readConnectorsQuery struct {
	Connectors struct {
		Edges []*struct {
			Node gqlConnector
		}
	}
}

func (client *Client) ReadConnectors(ctx context.Context) ([]*model.Connector, error) {
	response := readConnectorsQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readConnectors", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	if response.Connectors.Edges == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, "All")
	}

	return response.ToModel(), nil
}

type readConnectorQuery struct {
	Connector *gqlConnector `graphql:"connector(id: $id)"`
}

func (client *Client) ReadConnector(ctx context.Context, connectorID string) (*model.Connector, error) {
	if connectorID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", connectorResourceName)
	}

	variables := newVariables().withID(connectorID).value()
	response := readConnectorQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readConnector", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	if response.Connector == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", connectorResourceName, connectorID)
	}

	return response.ToModel(), nil
}

type deleteConnectorQuery struct {
	ConnectorDelete *OkError `graphql:"connectorDelete(id: $id)" json:"connectorDelete"`
}

func (client *Client) DeleteConnector(ctx context.Context, connectorID string) error {
	if connectorID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", connectorResourceName)
	}

	variables := newVariables().withID(connectorID).value()
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
