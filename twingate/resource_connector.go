package twingate

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/twingate/go-graphql-client"
)

func resourceConnector() *schema.Resource {
	return &schema.Resource{
		Description:   "Connectors provide connectivity to Remote Networks. This resource type will create the Connector in the Twingate Admin Console, but in order to successfully deploy it, you must also generate Connector tokens that authenticate the Connector with Twingate. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		CreateContext: resourceConnectorCreate,
		ReadContext:   resourceConnectorRead,
		DeleteContext: resourceConnectorDelete,
		UpdateContext: resourceConnectorUpdate,

		Schema: map[string]*schema.Schema{
			// required
			"remote_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Remote Network to attach the Connector to",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the Connector, if not provided one will be generated",
			},
			// computed
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Autogenerated ID of the Connector, encoded in base64",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	remoteNetworkID := d.Get("remote_network_id").(string)
	connector, err := client.createConnector(remoteNetworkID)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(connector.ID.(string))
	log.Printf("[INFO] Created conector %s", connector.Name)

	if d.Get("name").(string) != "" {
		return resourceConnectorUpdate(ctx, d, m)
	}

	return resourceConnectorRead(ctx, d, m)
}
func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	connectorName := d.Get("name").(string)

	if d.HasChange("name") {
		connectorID := d.Id()
		log.Printf("[INFO] Updating name of connector id %s", connectorID)

		if err := client.updateConnector(graphql.ID(connectorID), graphql.String(connectorName)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceConnectorRead(ctx, d, m)
}
func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	connectorID := d.Id()

	err := client.deleteConnector(connectorID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Destroyed connector id %s", d.Id())

	return diags
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	connectorID := d.Id()
	connector, err := client.readConnector(connectorID)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", connector.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %w ", err))
	}

	return diags
}
