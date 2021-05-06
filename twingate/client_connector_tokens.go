package twingate

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
)

type ConnectorTokens struct {
	AccessToken  string
	RefreshToken string
}

const connectorTokensResourceName = "connector tokens"

func (client *Client) verifyConnectorTokens(refreshToken, accessToken string) error {
	jsonValue, _ := json.Marshal(
		map[string]string{
			"refresh_token": refreshToken,
		})

	req, err := retryablehttp.NewRequest("POST", fmt.Sprintf("%s/access_node/refresh", client.APIServerURL), bytes.NewBuffer(jsonValue))
	if err != nil {
		return NewAPIError(err, "verify", connectorTokensResourceName)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	_, err = client.doRequest(req)
	if err != nil {
		return NewAPIError(err, "verify", connectorTokensResourceName)
	}

	return nil
}

func (client *Client) generateConnectorTokens(connector *Connector) error {
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
		return NewAPIError(err, "generate", connectorTokensResourceName)
	}
	createTokensResult := mutationConnector.Path("data.connectorGenerateTokens")
	status := createTokensResult.Path("ok").Data().(bool)
	if !status {
		message := createTokensResult.Path("error").Data().(string)

		return NewAPIErrorWithId(NewMutationError(message), "generate", connectorTokensResourceName, connector.Id)
	}

	connector.ConnectorTokens = &ConnectorTokens{
		AccessToken:  createTokensResult.Path("connectorTokens.accessToken").Data().(string),
		RefreshToken: createTokensResult.Path("connectorTokens.refreshToken").Data().(string),
	}

	return nil
}
