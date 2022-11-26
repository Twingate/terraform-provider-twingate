package sweepers

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const resourceResource = "twingate_resource"

func init() {
	resource.AddTestSweepers(resourceResource, &resource.Sweeper{
		Name: resourceResource,
		F: newTestSweeper(resourceResource,
			func(client *client.Client, ctx context.Context) ([]Resource, error) {
				resources, err := client.ReadResources(ctx)
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
				return client.DeleteResource(ctx, id)
			},
		),
	})
}
