package sweepers

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	sdk.AddTestSweepers(resource.TwingateServiceAccount, &sdk.Sweeper{
		Name: resource.TwingateServiceAccount,
		F: newTestSweeper(resource.TwingateServiceAccount,
			func(providerClient *client.Client, ctx context.Context) ([]Resource, error) {
				resources, err := providerClient.ReadShallowServiceAccounts(ctx)
				if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
					return nil, err
				}

				items := make([]Resource, 0, len(resources))
				for _, r := range resources {
					items = append(items, r)
				}
				return items, nil
			},
			func(providerClient *client.Client, ctx context.Context, id string) error {
				service, err := providerClient.ReadServiceAccount(ctx, id)
				if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
					return err
				}

				if service != nil {
					for _, keyID := range service.Keys {
						key, err := providerClient.ReadServiceKey(ctx, keyID)
						if err != nil {
							return err
						}

						if key.IsActive() {
							err = providerClient.RevokeServiceKey(ctx, keyID)
							if err != nil {
								return err
							}
						}
					}
				}

				return providerClient.DeleteServiceAccount(ctx, id)
			},
		),
	})
}
