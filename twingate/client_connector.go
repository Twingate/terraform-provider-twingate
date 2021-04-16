package twingate

import (
	"fmt"
)

type Connector struct {
	Id            string
	remoteNetwork *RemoteNetwork
	Name          string
	AccessToken   string
	RefreshToken  string
}

func (client *Client) createConnector(remoteNetworkId *string) (*Connector, error) {
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
        `, *remoteNetworkId),
	}
	mutationConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, err
	}

	connectorResult := mutationConnector.Path("data.connectorCreate")
	status := connectorResult.Path("ok").Data().(bool)

	if !status {
		errorString := connectorResult.Path("error").Data().(string)
		return nil, fmt.Errorf("Cant create connector under the network with name %s, Error:  %s", *remoteNetworkId, errorString)
	}

	connector := Connector{
		Id:   connectorResult.Path("entity.id").Data().(string),
		Name: connectorResult.Path("entity.name").Data().(string),
	}
	return &connector, nil
}

func (client *Client) populateConnectorTokens(connector *Connector) error {

	mutation := map[string]string{
		"query": fmt.Sprintf(`
			mutation{
			  connectorGenerateTokens(connectorId: "%s"){
				connectorTokens {
				  accessToken
				  refreshToken
				}
				ok
				error
			  }
			}
        `, connector.Id),
	}
	mutationConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return err
	}
	createTokensResult := mutationConnector.Path("data.connectorGenerateTokens")
	status := createTokensResult.Path("ok").Data().(bool)
	if !status {
		errorString := createTokensResult.Path("error").Data().(string)
		return fmt.Errorf("Cant create tokens for connector %s, Error:  %s", connector.Id, errorString)
	}

	connector.AccessToken = createTokensResult.Path("connectorTokens.accessToken").Data().(string)
	connector.RefreshToken = createTokensResult.Path("connectorTokens.refreshToken").Data().(string)

	return nil
}

func (client *Client) readConnector(connectorId *string) (*Connector, error) {
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
        `, *connectorId),
	}
	queryConnector, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, err
	}

	connectorRead := queryConnector.Path("data.connector")

	if connectorRead.Data() == nil {
		return nil, fmt.Errorf("Unable to read connector %s information ", *connectorId)
	}

	connector := Connector{
		Id:   connectorRead.Path("id").Data().(string),
		Name: connectorRead.Path("name").Data().(string),
		remoteNetwork: &RemoteNetwork{
			Id:   connectorRead.Path("remoteNetwork.id").Data().(string),
			Name: connectorRead.Path("remoteNetwork.name").Data().(string),
		},
	}

	return &connector, nil

}

func (client *Client) deleteConnector(connectorId *string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  connectorDelete(id: "%s"){
			ok
			error
		  }
		}
		`, *connectorId),
	}
	mutationConnector, err := client.doGraphqlRequest(mutation)

	if err != nil {
		return err
	}
	connectorDelete := mutationConnector.Path("data.connectorDelete")
	status := connectorDelete.Path("ok").Data().(bool)
	if !status {
		return fmt.Errorf("Unable to delete network with Id %s, Error:  %s", *connectorId, connectorDelete.Path("error").Data().(string))
	}

	return nil
}
