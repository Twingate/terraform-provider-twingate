package twingate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceRemoteNetworkRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	networkNameID := resourceData.Get("id").(string)
	remoteNetwork, err := client.readRemoteNetwork(ctx, networkNameID)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("name", string(remoteNetwork.Name)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func datasourceRemoteNetwork() *schema.Resource {
	return &schema.Resource{
		Description: "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		ReadContext: datasourceRemoteNetworkRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Remote Network",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Remote Network",
			},
		},
	}
}
