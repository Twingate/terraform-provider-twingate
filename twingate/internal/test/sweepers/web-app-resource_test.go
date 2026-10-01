package sweepers

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/provider/resource"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	sdk.AddTestSweepers(resource.TwingateWebAppResource, &sdk.Sweeper{
		Name: resource.TwingateWebAppResource,
		F: newTestSweeper(resource.TwingateWebAppResource,
			func(providerClient *client.Client, ctx context.Context) ([]Resource, error) {
				resources, err := providerClient.ReadWebAppResources(ctx)
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
				return providerClient.DeleteWebAppResource(ctx, id)
			},
		),
	})
}
