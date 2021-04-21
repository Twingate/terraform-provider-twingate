package twingate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ConnectorTokens struct {
	AccessToken  string
	RefreshToken string
}

func (client *Client) verifyConnectorTokens(refreshToken, accessToken *string) error {
	jsonValue, _ := json.Marshal(
		map[string]string{
			"refresh_token": *refreshToken,
		})

	req, err := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("%s/access_node/refresh", client.APIServerURL), bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("can't create context : %w ", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *accessToken))
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
		return err
	}
	createTokensResult := mutationConnector.Path("data.connectorGenerateTokens")
	status := createTokensResult.Path("ok").Data().(bool)
	if !status {
		errorString := createTokensResult.Path("error").Data().(string)

		return fmt.Errorf("cant create tokens for connector %s, error:  %w", connector.Id, APIError(errorString))
	}

	connector.ConnectorTokens = &ConnectorTokens{
		AccessToken:  createTokensResult.Path("connectorTokens.accessToken").Data().(string),
		RefreshToken: createTokensResult.Path("connectorTokens.refreshToken").Data().(string),
	}

	return nil
}
