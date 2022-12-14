package sweepers

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	sdk.AddTestSweepers(resource.TwingateServiceAccount, &sdk.Sweeper{
		Name: resource.TwingateServiceAccount,
		F: newTestSweeper(resource.TwingateServiceAccount,
			func(client *client.Client, ctx context.Context) ([]Resource, error) {
				resources, err := client.ReadServiceAccounts(ctx)
				if err != nil {
					return nil, err
				}

				items := make([]Resource, 0, len(resources))
				for _, r := range resources {
					items = append(items, r)
				}
				return items, nil
			},
			func(client *client.Client, ctx context.Context, id string) error {
				return client.DeleteServiceAccount(ctx, id)
			},
		),
	})
}
