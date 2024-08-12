package sweepers

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const resourceDNSProfile = "twingate_dns_filtering_profile"

func init() {
	resource.AddTestSweepers(resourceDNSProfile, &resource.Sweeper{
		Name: resourceDNSProfile,
		F: newTestSweeper(resourceDNSProfile,
			func(client *client.Client, ctx context.Context) ([]Resource, error) {
				resources, err := client.ReadShallowDNSFilteringProfiles(ctx)
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
				return client.DeleteDNSFilteringProfile(ctx, id)
			},
		),
	})
}
