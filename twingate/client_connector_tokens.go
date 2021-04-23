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

func (client *Client) verifyConnectorTokens(refreshToken, accessToken string) error {
	jsonValue, _ := json.Marshal(
		map[string]string{
			"refresh_token": refreshToken,
		})

	req, err := retryablehttp.NewRequest("POST", fmt.Sprintf("%s/access_node/refresh", client.APIServerURL), bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("could not create Api request : %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	body, err := client.doRequest(req)
	_ = body
	if err != nil {
		return fmt.Errorf("connector tokens are invalid : %w", err)
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
		return fmt.Errorf("can't generate tokens : %w", err)
	}
	createTokensResult := mutationConnector.Path("data.connectorGenerateTokens")
	status := createTokensResult.Path("ok").Data().(bool)
	if !status {
		errorString := createTokensResult.Path("error").Data().(string)

		return APIError("can't create tokens for connector %s, error: %s", connector.Id, errorString)
	}

	connector.ConnectorTokens = &ConnectorTokens{
		AccessToken:  createTokensResult.Path("connectorTokens.accessToken").Data().(string),
		RefreshToken: createTokensResult.Path("connectorTokens.refreshToken").Data().(string),
	}

	return nil
}
