package resource

import (
	"context"
	"errors"
	"log"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func RemoteNetwork() *schema.Resource {
	return &schema.Resource{
		Description:   "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		CreateContext: remoteNetworkCreate,
		ReadContext:   remoteNetworkRead,
		UpdateContext: remoteNetworkUpdate,
		DeleteContext: remoteNetworkDelete,

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

func remoteNetworkCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)

	var diags diag.Diagnostics

	remoteNetworkName := resourceData.Get("name").(string)
	remoteNetwork, err := client.CreateRemoteNetwork(ctx, remoteNetworkName)

	if err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(remoteNetwork.ID.(string))
	log.Printf("[INFO] Remote network %s created with id %s", remoteNetworkName, resourceData.Id())
	remoteNetworkRead(ctx, resourceData, meta)

	return diags
}

func remoteNetworkUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)

	remoteNetworkName := resourceData.Get("name").(string)

	if resourceData.HasChange("name") {
		remoteNetworkID := resourceData.Id()
		log.Printf("[INFO] Updating remote network id %s", remoteNetworkID)

		if err := client.UpdateRemoteNetwork(ctx, remoteNetworkID, remoteNetworkName); err != nil {
			return diag.FromErr(err)
		}
	}

	return remoteNetworkRead(ctx, resourceData, meta)
}

func remoteNetworkDelete(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*transport.Client)

	var diags diag.Diagnostics

	remoteNetworkID := resourceData.Id()

	err := client.DeleteRemoteNetwork(ctx, remoteNetworkID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleted remote network id %s", remoteNetworkID)

	return diags
}

func remoteNetworkRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)

	var diags diag.Diagnostics

	remoteNetworkID := resourceData.Id()

	log.Printf("[INFO] Reading remote network id %s", remoteNetworkID)

	remoteNetwork, err := client.ReadRemoteNetworkByID(ctx, remoteNetworkID)

	if err != nil {
		if errors.Is(err, transport.ErrGraphqlResultIsEmpty) {
			// clear state
			resourceData.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	err = resourceData.Set("name", remoteNetwork.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
