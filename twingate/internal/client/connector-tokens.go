package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

const connectorTokensResourceName = "connector tokens"

func (client *Client) VerifyConnectorTokens(ctx context.Context, refreshToken, accessToken string) error {
	payload := map[string]string{
		"refresh_token": refreshToken,
	}

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	_, err := client.post(ctx, "/connector/validate_tokens", payload, headers)
	if err != nil {
		return NewAPIError(err, "verify", connectorTokensResourceName)
	}

	return nil
}

func (client *Client) GenerateConnectorTokens(ctx context.Context, connectorID string) (*model.ConnectorTokens, error) {
	opr := resourceConnectorToken.generate()

	variables := newVars(gqlID(connectorID, "connectorId"))
	response := query.GenerateConnectorTokens{}

	err := client.mutate(ctx, &response, variables, opr.withCustomName("generateConnectorTokens"), attr{id: connectorID})
	if err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}
