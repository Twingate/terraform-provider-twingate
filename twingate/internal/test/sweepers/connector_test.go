package sweepers

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const resourceConnector = "twingate_connector"

func init() {
	resource.AddTestSweepers(resourceConnector, &resource.Sweeper{
		Name: resourceConnector,
		F: newTestSweeper(resourceConnector,
			func(client *transport.Client, ctx context.Context) ([]Resource, error) {
				resources, err := client.ReadConnectors(ctx)
				if err != nil {
					return nil, err
				}

				items := make([]Resource, 0, len(resources))
				for _, r := range resources {
					items = append(items, r)
				}
				return items, nil
			},
			func(client *transport.Client, ctx context.Context, id string) error {
				return client.DeleteConnector(ctx, id)
			},
		),
	})
}
