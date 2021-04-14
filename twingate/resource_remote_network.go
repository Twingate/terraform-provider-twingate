package twingate

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRemoteNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRemoteNetworkCreate,
		ReadContext:   resourceRemoteNetworkRead,
		UpdateContext: resourceRemoteNetworkUpdate,
		DeleteContext: resourceRemoteNetworkDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the remote network",
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
	remoteNetwork, err := client.createRemoteNetwork(&remoteNetworkName)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(remoteNetwork.Id)
	log.Printf("[INFO] Remote network %s created with id %s", remoteNetworkName, d.Id())
	resourceRemoteNetworkRead(ctx, d, m)

	return diags
}

func resourceRemoteNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	remoteNetworkName := d.Get("name").(string)

	if d.HasChange("name") {
		remoteNetworkId := d.Id()
		log.Printf("[INFO] Updating remote network id %s", remoteNetworkId)
		if err := client.updateRemoteNetwork(&remoteNetworkId, &remoteNetworkName); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceRemoteNetworkRead(ctx, d, m)
}

func resourceRemoteNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkId := d.Id()

	err := client.deleteRemoteNetwork(&remoteNetworkId)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleted remote network id %s", d.Id())

	return diags
}

func resourceRemoteNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkId := d.Id()

	log.Printf("[INFO] Reading remote network id %s", d.Id())

	remoteNetwork, err := client.readRemoteNetwork(&remoteNetworkId)

	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("name", remoteNetwork.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
