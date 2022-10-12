package twingate

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var ErrNotAllowChangeRemoteNetworkID = errors.New("connectors cannot be moved between Remote Networks: you must either create a new Connector or destroy and recreate the existing one")

func resourceConnector() *schema.Resource {
	return &schema.Resource{
		Description:   "Connectors provide connectivity to Remote Networks. This resource type will create the Connector in the Twingate Admin Console, but in order to successfully deploy it, you must also generate Connector tokens that authenticate the Connector with Twingate. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		CreateContext: resourceConnectorCreate,
		ReadContext:   resourceConnectorRead,
		DeleteContext: resourceConnectorDelete,
		UpdateContext: resourceConnectorUpdate,
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
			const key = "remote_network_id"
			oldVal, _ := d.GetChange(key)
			old := oldVal.(string)
			if old != "" && d.HasChange(key) {
				return ErrNotAllowChangeRemoteNetworkID
			}

			return nil
		},

		Schema: map[string]*schema.Schema{
			// required
			"remote_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Remote Network the Connector is attached to",
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

func resourceConnectorCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	remoteNetworkID := resourceData.Get("remote_network_id").(string)
	connectorName := resourceData.Get("name").(string)
	connector, err := client.createConnector(ctx, remoteNetworkID, connectorName)

	return resourceConnectorReadHelper(resourceData, connector, err)
}
func resourceConnectorUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// only `name` allowed to change
	if !resourceData.HasChange("name") {
		return nil
	}

	client := meta.(*Client)

	log.Printf("[INFO] Updating name of connector id %s", resourceData.Id())
	connector, err := client.updateConnector(ctx, resourceData.Id(), resourceData.Get("name").(string))

	return resourceConnectorReadHelper(resourceData, connector, err)
}
func resourceConnectorDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	connectorID := resourceData.Id()

	err := client.deleteConnector(ctx, connectorID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Destroyed connector id %s", connectorID)

	return diags
}

func resourceConnectorRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	connector, err := client.readConnector(ctx, resourceData.Id())

	return resourceConnectorReadHelper(resourceData, connector, err)
}

func resourceConnectorReadHelper(resourceData *schema.ResourceData, connector *Connector, err error) diag.Diagnostics {
	if err != nil {
		if errors.Is(err, ErrGraphqlResultIsEmpty) {
			// clear state
			resourceData.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if err := resourceData.Set("name", string(connector.Name)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %w ", err))
	}

	var connectorRemoteNetworkID string
	if connector.RemoteNetwork != nil {
		connectorRemoteNetworkID = connector.RemoteNetwork.ID.(string)
	}

	if connectorRemoteNetworkID != "" {
		if err := resourceData.Set("remote_network_id", connectorRemoteNetworkID); err != nil {
			return diag.FromErr(fmt.Errorf("error setting remote_network_id: %w ", err))
		}
	}

	resourceData.SetId(connector.ID.(string))

	return nil
}
