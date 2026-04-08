package sweepers

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/resource"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	sdk.AddTestSweepers(resource.TwingateSSHCertificateAuthority, &sdk.Sweeper{
		Name: resource.TwingateSSHCertificateAuthority,
		F: newTestSweeper(resource.TwingateSSHCertificateAuthority,
			func(providerClient *client.Client, ctx context.Context) ([]Resource, error) {
				resources, err := providerClient.ReadSSHCertificateAuthorities(ctx)
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
				return providerClient.DeleteSSHCertificateAuthority(ctx, id)
			},
		),
	})
}
