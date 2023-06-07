package sweepers

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	twingate "github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resources"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	name := twingate.TwingateUser
	resource.AddTestSweepers(name, &resource.Sweeper{
		Name: name,
		F: newTestSweeper(name,
			func(client *client.Client, ctx context.Context) ([]Resource, error) {
				resources, err := client.ReadUsers(ctx)
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
				return client.DeleteUser(ctx, id)
			},
		),
	})
}
