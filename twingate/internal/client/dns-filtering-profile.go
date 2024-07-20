package client

import (
	"context"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

func (client *Client) ReadDNSFilteringProfile(ctx context.Context, profileID string) (*model.Connector, error) {
	opr := resourceDNSFilteringProfile.read()

	if profileID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	// todo
	response := query.ReadConnector{}
	if err := client.query(ctx, &response, newVars(gqlID(profileID)), opr, attr{id: profileID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}
