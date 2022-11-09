package resource

import (
	"context"
	"errors"
	"log"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Group() *schema.Resource {
	return &schema.Resource{
		Description:   "Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).",
		CreateContext: groupCreate,
		ReadContext:   groupRead,
		DeleteContext: groupDelete,
		UpdateContext: groupUpdate,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the group",
			},
			// computed
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Autogenerated ID of the Resource, encoded in base64",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func groupCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)

	group, err := c.CreateGroup(ctx, resourceData.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Group %s created with id %v", group.Name, group.ID)

	return resourceGroupReadHelper(resourceData, group, nil)
}

func groupUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)

	group, err := c.UpdateGroup(ctx, resourceData.Id(), resourceData.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Updated group id %v", group.ID)

	return resourceGroupReadHelper(resourceData, group, err)
}

func groupDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	groupID := resourceData.Id()

	err := c.DeleteGroup(ctx, groupID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleted group id %s", resourceData.Id())

	return nil
}

func groupRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	group, err := c.ReadGroup(ctx, resourceData.Id())

	return resourceGroupReadHelper(resourceData, group, err)
}

func resourceGroupReadHelper(resourceData *schema.ResourceData, group *model.Group, err error) diag.Diagnostics {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			resourceData.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if err := resourceData.Set("name", group.Name); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(group.ID)

	return nil
}
