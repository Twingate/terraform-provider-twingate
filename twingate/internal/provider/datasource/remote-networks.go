package datasource

import (
	"context"
	"fmt"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceRemoteNetworksRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	remoteNetworks, err := client.ReadRemoteNetworks(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("remote_networks", convertRemoteNetworksToTerraform(remoteNetworks)); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId("all-remote-networks")

	return nil
}

func RemoteNetworks() *schema.Resource {
	return &schema.Resource{
		Description: "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		ReadContext: datasourceRemoteNetworksRead,
		Schema: map[string]*schema.Schema{
			"remote_networks": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of Remote Networks",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the Remote Network",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the Remote Network",
						},
						"location": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: fmt.Sprintf("The location of the Remote Network. Must be one of the following: %s.", strings.Join(model.Locations, ", ")),
						},
					},
				},
			},
		},
	}
}
