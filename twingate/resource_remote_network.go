package twingate

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/twingate/go-graphql-client"
)

func resourceRemoteNetwork() *schema.Resource {
	return &schema.Resource{
		Description:   "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		CreateContext: resourceRemoteNetworkCreate,
		ReadContext:   resourceRemoteNetworkRead,
		UpdateContext: resourceRemoteNetworkUpdate,
		DeleteContext: resourceRemoteNetworkDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the Remote Network",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Remote Network",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceRemoteNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkName := d.Get("name").(string)
	remoteNetwork, err := client.createRemoteNetwork(graphql.String(remoteNetworkName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(remoteNetwork.ID.(string))
	log.Printf("[INFO] Remote network %s created with id %s", remoteNetworkName, d.Id())
	resourceRemoteNetworkRead(ctx, d, m)

	return diags
}

func resourceRemoteNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	remoteNetworkName := d.Get("name").(string)

	if d.HasChange("name") {
		remoteNetworkID := d.Id()
		log.Printf("[INFO] Updating remote network id %s", remoteNetworkID)

		if err := client.updateRemoteNetwork(graphql.ID(remoteNetworkID), graphql.String(remoteNetworkName)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceRemoteNetworkRead(ctx, d, m)
}

func resourceRemoteNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkID := d.Id()

	err := client.deleteRemoteNetwork(remoteNetworkID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleted remote network id %s", d.Id())

	return diags
}

func resourceRemoteNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkID := d.Id()

	log.Printf("[INFO] Reading remote network id %s", d.Id())

	remoteNetwork, err := client.readRemoteNetwork(remoteNetworkID)

	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("name", remoteNetwork.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
