package datasource

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceConnectorsRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)

	connectors, err := c.ReadConnectors(ctx)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("connectors", convertConnectorsToTerraform(connectors)); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId("all-connectors")

	return nil
}

func Connectors() *schema.Resource {
	return &schema.Resource{
		Description: "Connectors provide connectivity to Remote Networks. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		ReadContext: datasourceConnectorsRead,
		Schema: map[string]*schema.Schema{
			"connectors": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of Connectors",
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
							Description: "The Name of the Connector",
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
