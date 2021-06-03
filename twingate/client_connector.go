package twingate

import (
	"fmt"
	"log"
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
			Entity struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"entity"`
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
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
	log.Println(r.Data.ConnectorCreate.Ok)
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
			Edges []struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"edges"`
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
	Data *readConnectorResponseData `json:"data"`
}

type readConnectorResponseData struct {
	Id        string                              `json:"id"`
	Name      string                              `json:"name"`
	Connector *readConnectorResponseDataConnector `json:"connector"`
}

type readConnectorResponseDataConnector struct {
	Id            string                                           `json:"id"`
	Name          string                                           `json:"name"`
	RemoteNetwork *readConnectorResponseDataConnectorRemoteNetwork `json:"remoteNetwork"`
}

type readConnectorResponseDataConnectorRemoteNetwork struct {
	Id   string `json:"id"`
	Name string `json:"name"`
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
	rn.ID = r.Data.Connector.RemoteNetwork.Id
	rn.Name = r.Data.Connector.RemoteNetwork.Name
	connector := Connector{
		ID:            r.Data.Connector.Id,
		Name:          r.Data.Connector.Name,
		RemoteNetwork: rn,
	}

	return &connector, nil
}

type deleteConnectorResponse struct {
	Data struct {
		ConnectorDelete struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		} `json:"connectorDelete"`
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
