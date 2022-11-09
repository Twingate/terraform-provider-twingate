package client

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const connectorTokensResourceName = "connector tokens"

type gqlConnectorTokens struct {
	AccessToken  graphql.String
	RefreshToken graphql.String
}

func (client *Client) VerifyConnectorTokens(ctx context.Context, refreshToken, accessToken string) error {
	payload := map[string]string{
		"refresh_token": refreshToken,
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}

	_, err := client.post(ctx, "/connector/validate_tokens", payload, headers)
	if err != nil {
		return NewAPIError(err, "verify", connectorTokensResourceName)
	}

	return nil
}

type generateConnectorTokensQuery struct {
	ConnectorGenerateTokens struct {
		ConnectorTokens gqlConnectorTokens
		OkError
	} `graphql:"connectorGenerateTokens(connectorId: $connectorId)"`
}

func (client *Client) GenerateConnectorTokens(ctx context.Context, connectorID string) (*model.ConnectorTokens, error) {
	variables := newVars(gqlID(connectorID, "connectorId"))
	response := generateConnectorTokensQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "generateConnectorTokens", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "generate", connectorTokensResourceName)
	}

	if !response.ConnectorGenerateTokens.Ok {
		message := response.ConnectorGenerateTokens.Error

		return nil, NewAPIErrorWithID(NewMutationError(message), "generate", connectorTokensResourceName, connectorID)
	}

	return response.ToModel(), nil
}
