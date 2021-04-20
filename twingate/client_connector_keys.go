package twingate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (client *Client) verifyConnectorTokens(refreshToken, accessToken *string) error {

	jsonValue, _ := json.Marshal(
		map[string]string{
			"refresh_token": *refreshToken,
		})

	req, err := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("%s/access_node/refresh", client.ApiServerURL), bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("Cant create context : %s ", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *accessToken))
	body, err := client.doRequest(req)
	_ = body
	if err != nil {
		return fmt.Errorf("Connector tokens are invalid : %s ", err)
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
