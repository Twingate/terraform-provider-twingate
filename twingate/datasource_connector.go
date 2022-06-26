package twingate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceConnectorRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	connectorID := resourceData.Get("id").(string)
	connector, err := client.readConnector(ctx, connectorID)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("name", string(connector.Name)); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("remote_network_id", connector.RemoteNetwork.ID.(string)); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(connectorID)

	return diags
}

func datasourceConnector() *schema.Resource {
	return &schema.Resource{
		Description: "Connectors provide connectivity to Remote Networks. This resource type will create the Connector in the Twingate Admin Console, but in order to successfully deploy it, you must also generate Connector tokens that authenticate the Connector with Twingate. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		ReadContext: datasourceConnectorRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Connector",
			},
			// computed
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the Connector",
			},
			"remote_network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the Remote Network attached to the Connector",
			},
		},
	}
}
