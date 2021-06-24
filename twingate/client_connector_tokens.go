package twingate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hasura/go-graphql-client"
)

type connectorTokens struct {
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

type generateConnectorTokensQuery struct {
	ConnectorGenerateTokens struct {
		ConnectorTokens struct {
			AccessToken  graphql.String
			RefreshToken graphql.String
		}
		OkError
	} `graphql:"connectorGenerateTokens(connectorId: $connectorId)"`
}

func (client *Client) generateConnectorTokens(connector *Connector) error {
	variables := map[string]interface{}{
		"connectorId": connector.ID,
	}

	r := generateConnectorTokensQuery{}

	err := client.GraphqlClient.Mutate(context.Background(), &r, variables)
	if err != nil {
		return NewAPIError(err, "generate", connectorTokensResourceName)
	}

	if !r.ConnectorGenerateTokens.Ok {
		message := r.ConnectorGenerateTokens.Error

		return NewAPIErrorWithID(NewMutationError(message), "generate", connectorTokensResourceName, connector.ID)
	}

	connector.ConnectorTokens = &connectorTokens{
		AccessToken:  string(r.ConnectorGenerateTokens.ConnectorTokens.AccessToken),
		RefreshToken: string(r.ConnectorGenerateTokens.ConnectorTokens.RefreshToken),
	}

	return nil
}
