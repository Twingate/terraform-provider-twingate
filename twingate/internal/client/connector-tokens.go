package client

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

const connectorTokensResourceName = "connector tokens"

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

func (client *Client) GenerateConnectorTokens(ctx context.Context, connectorID string) (*model.ConnectorTokens, error) {
	variables := newVars(gqlID(connectorID, "connectorId"))
	response := query.GenerateConnectorTokens{}

	err := client.GraphqlClient.NamedMutate(ctx, "generateConnectorTokens", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "generate", connectorTokensResourceName)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), "generate", connectorTokensResourceName, connectorID)
	}

	return response.ToModel(), nil
}
