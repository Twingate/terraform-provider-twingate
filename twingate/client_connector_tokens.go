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

type generateConnectorTokensResponse struct {
	Data struct {
		ConnectorGenerateTokens struct {
			ConnectorTokens struct {
				AccessToken  string `json:"accessToken"`
				RefreshToken string `json:"refreshToken"`
			} `json:"connectorTokens"`
			*OkErrorResponse
		} `json:"connectorGenerateTokens"`
	} `json:"data"`
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
        `, connector.ID),
	}

	r := generateConnectorTokensResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIError(err, "generate", connectorTokensResourceName)
	}

	if !r.Data.ConnectorGenerateTokens.Ok {
		message := r.Data.ConnectorGenerateTokens.Error

		return NewAPIErrorWithID(NewMutationError(message), "generate", connectorTokensResourceName, connector.ID)
	}

	connector.ConnectorTokens = &ConnectorTokens{
		AccessToken:  r.Data.ConnectorGenerateTokens.ConnectorTokens.AccessToken,
		RefreshToken: r.Data.ConnectorGenerateTokens.ConnectorTokens.RefreshToken,
	}

	return nil
}
