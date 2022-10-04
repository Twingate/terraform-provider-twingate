package twingate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/twingate/go-graphql-client"
)

type ConnectorTokens struct {
	AccessToken  string
	RefreshToken string
}

const connectorTokensResourceName = "connector tokens"

func (client *Client) verifyConnectorTokens(ctx context.Context, refreshToken, accessToken string) error {
	jsonValue, err := json.Marshal(
		map[string]string{
			"refresh_token": refreshToken,
		})
	if err != nil {
		return NewAPIError(err, "verify", connectorTokensResourceName)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/access_node/refresh", client.APIServerURL),
		bytes.NewBuffer(jsonValue))
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

func (client *Client) generateConnectorTokens(ctx context.Context, connector *Connector) (*ConnectorTokens, error) {
	variables := map[string]interface{}{
		"connectorId": connector.ID,
	}

	response := generateConnectorTokensQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "generateConnectorTokens", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "generate", connectorTokensResourceName)
	}

	if !response.ConnectorGenerateTokens.Ok {
		message := response.ConnectorGenerateTokens.Error

		return nil, NewAPIErrorWithID(NewMutationError(message), "generate", connectorTokensResourceName, connector.ID)
	}

	return &ConnectorTokens{
		AccessToken:  string(response.ConnectorGenerateTokens.ConnectorTokens.AccessToken),
		RefreshToken: string(response.ConnectorGenerateTokens.ConnectorTokens.RefreshToken),
	}, nil
}
