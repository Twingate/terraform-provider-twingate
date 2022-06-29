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
		Description: "Connectors provide connectivity to Remote Networks. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		ReadContext: datasourceConnectorRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Connector. The ID for the Connector must be obtained from the Admin API.",
			},
			// computed
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Connector",
			},
			"remote_network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the Remote Network the Connector is attached to",
			},
		},
	}
}
