package sweepers

import (
	"context"
	"errors"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/test"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const resourceConnector = "twingate_connector"

func init() {
	resource.AddTestSweepers(resourceConnector, &resource.Sweeper{
		Name: resourceConnector,
		F: newTestSweeper(resourceConnector,
			func(c *client.Client, ctx context.Context) ([]Resource, error) {
				resources, err := c.ReadConnectors(ctx, test.Prefix(), attr.FilterByPrefix)
				if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
					return nil, err
				}

				items := make([]Resource, 0, len(resources))
				for _, r := range resources {
					items = append(items, r)
				}
				return items, nil
			},
			func(client *client.Client, ctx context.Context, id string) error {
				return client.DeleteConnector(ctx, id)
			},
		),
	})
}
