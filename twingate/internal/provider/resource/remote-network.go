package resource

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func RemoteNetwork() *schema.Resource {
	return &schema.Resource{
		Description:   "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		CreateContext: remoteNetworkCreate,
		ReadContext:   remoteNetworkRead,
		UpdateContext: remoteNetworkUpdate,
		DeleteContext: remoteNetworkDelete,

		Schema: map[string]*schema.Schema{
			attr.ID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the Remote Network",
			},
			attr.Name: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Remote Network",
			},
			attr.Location: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(model.Locations, false),
				Description:  fmt.Sprintf("The location of the Remote Network. Must be one of the following: %s.", strings.Join(model.Locations, ", ")),
				Default:      model.LocationOther,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func remoteNetworkCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	remoteNetwork, err := c.CreateRemoteNetwork(ctx, &model.RemoteNetwork{
		Name:     resourceData.Get(attr.Name).(string),
		Location: resourceData.Get(attr.Location).(string),
	})

	return resourceRemoteNetworkReadHelper(resourceData, remoteNetwork, err)
}

func remoteNetworkUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Updating remote network id %s", resourceData.Id())

	var name string
	if resourceData.HasChange(attr.Name) {
		name = resourceData.Get(attr.Name).(string)
	}

	c := meta.(*client.Client)
	remoteNetwork, err := c.UpdateRemoteNetwork(ctx, &model.RemoteNetwork{
		ID:       resourceData.Id(),
		Name:     name,
		Location: resourceData.Get(attr.Location).(string),
	})

	return resourceRemoteNetworkReadHelper(resourceData, remoteNetwork, err)
}

func remoteNetworkDelete(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	err := c.DeleteRemoteNetwork(ctx, resourceData.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleted remote network id %s", resourceData.Id())

	return nil
}

func remoteNetworkRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	remoteNetwork, err := c.ReadRemoteNetworkByID(ctx, resourceData.Id())

	return resourceRemoteNetworkReadHelper(resourceData, remoteNetwork, err)
}

func resourceRemoteNetworkReadHelper(resourceData *schema.ResourceData, remoteNetwork *model.RemoteNetwork, err error) diag.Diagnostics {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			resourceData.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.Name, remoteNetwork.Name); err != nil {
		return ErrAttributeSet(err, attr.Name)
	}

	if err := resourceData.Set(attr.Location, remoteNetwork.Location); err != nil {
		return ErrAttributeSet(err, attr.Location)
	}

	resourceData.SetId(remoteNetwork.ID)

	return nil
}
