package twingate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceConnectorsRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	connectors, err := client.readConnectorsWithRemoteNetwork(ctx)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("connectors", convertConnectorsToTerraform(connectors)); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId("all-connectors")

	return diags
}

func convertConnectorsToTerraform(connectors []*Connector) []interface{} {
	out := make([]interface{}, 0, len(connectors))

	for _, connector := range connectors {
		out = append(out, convertConnectorToTerraform(connector))
	}

	return out
}

func convertConnectorToTerraform(connector *Connector) map[string]interface{} {
	return map[string]interface{}{
		"id":                connector.ID.(string),
		"name":              string(connector.Name),
		"remote_network_id": connector.RemoteNetwork.ID.(string),
	}
}

func datasourceConnectors() *schema.Resource {
	return &schema.Resource{
		Description: "Connectors provide connectivity to Remote Networks. This resource type will create the Connector in the Twingate Admin Console, but in order to successfully deploy it, you must also generate Connector tokens that authenticate the Connector with Twingate. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		ReadContext: datasourceConnectorsRead,
		Schema: map[string]*schema.Schema{
			"connectors": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of connectors",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the Connector",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the Connector",
						},
						"remote_network_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the Remote Network attached to the Connector",
						},
					},
				},
			},
		},
	}
}
