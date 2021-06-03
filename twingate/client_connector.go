package twingate

import (
	"fmt"
	"log"
)

type Connector struct {
	ID              string
	RemoteNetwork   *RemoteNetwork
	Name            string
	ConnectorTokens *ConnectorTokens
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

type readConnectorResponse struct {
	Data *readConnectorResponseData `json:"data"`
}

type readConnectorResponseData struct {
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

func newReadConnectorResponse() *readConnectorResponse {
	return &readConnectorResponse{
		Data: &readConnectorResponseData{
			Connector: &readConnectorResponseDataConnector{
				RemoteNetwork: &readConnectorResponseDataConnectorRemoteNetwork{},
			},
		},
	}
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

	r := newReadConnectorResponse()

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", connectorResourceName, connectorID)
	}

	if r.Data == nil {
		return nil, NewAPIErrorWithID(nil, "reed", connectorResourceName, connectorID)
	}

	connector := Connector{
		ID:   r.Data.Connector.Id,
		Name: r.Data.Connector.Name,
		RemoteNetwork: &RemoteNetwork{
			ID:   r.Data.Connector.RemoteNetwork.Id,
			Name: r.Data.Connector.RemoteNetwork.Name,
		},
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
