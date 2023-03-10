package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceConnectorRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	connectorID := resourceData.Get(attr.ID).(string)

	connector, err := c.ReadConnector(ctx, connectorID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.Name, connector.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.RemoteNetworkID, connector.NetworkID); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.StatusUpdatesEnabled, *connector.StatusUpdatesEnabled); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(connectorID)

	return nil
}

func Connector() *schema.Resource {
	return &schema.Resource{
		Description: "Connectors provide connectivity to Remote Networks. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		ReadContext: datasourceConnectorRead,
		Schema: map[string]*schema.Schema{
			attr.ID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Connector. The ID for the Connector must be obtained from the Admin API.",
			},
			// computed
			attr.Name: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Connector.",
			},
			attr.RemoteNetworkID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the Remote Network the Connector is attached to.",
			},
			attr.StatusUpdatesEnabled: {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Determines whether status notifications are enabled for the Connector.",
			},
		},
	}
}
