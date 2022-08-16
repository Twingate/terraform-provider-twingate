package twingate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceRemoteNetworkRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	networkID := resourceData.Get("id").(string)
	networkName := resourceData.Get("name").(string)
	network, err := client.readRemoteNetwork(ctx, networkID, networkName)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("name", network.Name); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(network.ID.(string))

	return diags
}

func datasourceRemoteNetwork() *schema.Resource {
	return &schema.Resource{
		Description: "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		ReadContext: datasourceRemoteNetworkRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The ID of the Remote Network",
				ExactlyOneOf: []string{"name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The name of the Remote Network",
				ExactlyOneOf: []string{"id"},
			},
		},
	}
}
