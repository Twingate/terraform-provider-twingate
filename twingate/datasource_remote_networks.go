package twingate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceRemoteNetworksRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	networkName := resourceData.Get("name").(string)
	networks, err := client.readRemoteNetworksByName(ctx, networkName)

	if err != nil {
		return diag.FromErr(err)
	}

	data := make([]interface{}, 0, len(networks))
	for _, network := range networks {
		data = append(data, map[string]interface{}{
			"id":   network.ID.(string),
			"name": string(network.Name),
		})
	}

	if err := resourceData.Set("remote_networks", data); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(networkName)

	return diags
}

func datasourceRemoteNetworks() *schema.Resource {
	return &schema.Resource{
		Description: "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		ReadContext: datasourceRemoteNetworksRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Remote Network (case-sensitive, exact match)",
			},
			"remote_networks": {
				Type:        schema.TypeList,
				Computed:    true,
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
					},
				},
			},
		},
	}
}
