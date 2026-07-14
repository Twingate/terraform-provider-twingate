package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
)

func (client *Client) ReadSyncToS3OidcURL(ctx context.Context) (string, error) {
	opr := resourceSyncToS3.read()

	response := query.ReadSyncToS3OidcURL{}

	if err := client.query(ctx, &response, newVars(), opr); err != nil {
		return "", err
	}

	return response.OidcURL, nil
}
