package twingate

import (
	"fmt"
)

type Connector struct {
	ID              string
	RemoteNetwork   *remoteNetwork
	Name            string
	ConnectorTokens *ConnectorTokens
}

type Connectors struct {
	ID   string
	Name string
}

const connectorResourceName = "connector"

type createConnectorResponse struct {
	Data struct {
		ConnectorCreate struct {
			Entity *IdNameResponse `json:"entity"`
			*OkErrorResponse
		} `json:"connectorCreate"`
	} `json:"data"`
}

func (client *Client) createConnector(remoteNetworkID string) (*Connector, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
			mutation{
			  connectorCreate(remoteNetworkId: "%s"){
				ok
				error
				entity{
				  id
                  name
				}
			  }
			}
        `, remoteNetworkID),
	}
	r := createConnectorResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return nil, NewAPIError(err, "create", connectorResourceName)
	}

	if !r.Data.ConnectorCreate.Ok {
		message := r.Data.ConnectorCreate.Error
		return nil, NewAPIError(NewMutationError(message), "create", connectorResourceName)
	}

	connector := Connector{
		ID:   r.Data.ConnectorCreate.Entity.ID,
		Name: r.Data.ConnectorCreate.Entity.Name,
	}

	return &connector, nil
}

type readConnectorsResponse struct {
	Data struct {
		Connectors struct {
			Edges []*EdgesResponse `json:"edges"`
		} `json:"connectors"`
	} `json:"data"`
}

func (client *Client) readConnectors() (map[int]*Connectors, error) { //nolint
	query := map[string]string{
		"query": "{ connectors { edges { node { id name } } } }",
	}

	r := readConnectorsResponse{}
	err := client.doGraphqlRequest(query, &r)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, "All")
	}

	var connectors = make(map[int]*Connectors)

	for i, elem := range r.Data.Connectors.Edges {
		c := &Connectors{ID: elem.Node.ID, Name: elem.Node.Name}
		connectors[i] = c
	}

	return connectors, nil
}

type readConnectorResponse struct {
	Data *struct {
		*IdNameResponse
		Connector *struct {
			*IdNameResponse
			RemoteNetwork *IdNameResponse `json:"remoteNetwork"`
		} `json:"connector"`
	} `json:"data"`
}

func (client *Client) readConnector(connectorID string) (*Connector, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  connector(id: "%s") {
			id
			name
			remoteNetwork {
				name
				id
			}
          }
		}
        `, connectorID),
	}

	r := readConnectorResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	if r.Data == nil || r.Data.Connector == nil {
		return nil, NewAPIErrorWithID(nil, "read", connectorResourceName, connectorID)
	}

	rn := &remoteNetwork{}
	rn.ID = r.Data.Connector.RemoteNetwork.ID
	rn.Name = r.Data.Connector.RemoteNetwork.Name
	connector := Connector{
		ID:            r.Data.Connector.ID,
		Name:          r.Data.Connector.Name,
		RemoteNetwork: rn,
	}

	return &connector, nil
}

type deleteConnectorResponse struct {
	Data struct {
		ConnectorDelete *OkErrorResponse `json:"connectorDelete"`
	} `json:"data"`
}

func (client *Client) deleteConnector(connectorID string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  connectorDelete(id: "%s"){
			ok
			error
		  }
		}
		`, connectorID),
	}

	r := deleteConnectorResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", connectorResourceName, connectorID)
	}

	if !r.Data.ConnectorDelete.Ok {
		message := r.Data.ConnectorDelete.Error

		return NewAPIErrorWithID(NewMutationError(message), "delete", connectorResourceName, connectorID)
	}

	return nil
}
