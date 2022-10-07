package twingate

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func resourceRemoteNetworkCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	remoteNetwork, err := client.createRemoteNetwork(ctx, resourceData.Get("name").(string))

	if err != nil {
		return diag.FromErr(err)
	}

	remoteNetworkID := remoteNetwork.ID.(string)
	remoteNetworkName := string(remoteNetwork.Name)

	log.Printf("[INFO] Remote network %s created with id %s", remoteNetworkName, remoteNetworkID)

	return resourceRemoteNetworkReadHelper(resourceData, remoteNetworkID, remoteNetworkName, nil)
}

func resourceRemoteNetworkUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	remoteNetworkID := resourceData.Id()
	remoteNetworkName := resourceData.Get("name").(string)

	if resourceData.HasChange("name") {
		log.Printf("[INFO] Updating remote network id %s", remoteNetworkID)

		remoteNetwork, err := client.updateRemoteNetwork(ctx, remoteNetworkID, remoteNetworkName)
		if err != nil {
			return diag.FromErr(err)
		}

		remoteNetworkID = remoteNetwork.ID.(string)
		remoteNetworkName = string(remoteNetwork.Name)
	}

	return resourceRemoteNetworkReadHelper(resourceData, remoteNetworkID, remoteNetworkName, nil)
}

func resourceRemoteNetworkDelete(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkID := resourceData.Id()

	err := client.deleteRemoteNetwork(ctx, remoteNetworkID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleted remote network id %s", remoteNetworkID)

	return diags
}

func resourceRemoteNetworkRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Reading remote network id %s", resourceData.Id())

	var (
		remoteNetworkID   string
		remoteNetworkName string
	)

	client := meta.(*Client)

	remoteNetwork, err := client.readRemoteNetworkByID(ctx, resourceData.Id())

	if remoteNetwork != nil {
		remoteNetworkName = string(remoteNetwork.Name)
		remoteNetworkID = remoteNetwork.ID.(string)
	}

	return resourceRemoteNetworkReadHelper(resourceData, remoteNetworkID, remoteNetworkName, err)
}

func resourceRemoteNetworkReadHelper(resourceData *schema.ResourceData, remoteNetworkID, remoteNetworkName string, err error) diag.Diagnostics {
	if err != nil {
		if errors.Is(err, ErrGraphqlResultIsEmpty) {
			// clear state
			resourceData.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	err = resourceData.Set("name", remoteNetworkName)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(remoteNetworkID)

	return nil
}
