package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceGroupRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	groupID := resourceData.Get("id").(string)

	group, err := c.ReadGroup(ctx, groupID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("name", group.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("type", group.Type); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("is_active", group.IsActive); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("security_policy_id", group.SecurityPolicyID); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(groupID)

	return nil
}

func Group() *schema.Resource {
	return &schema.Resource{
		Description: "Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).",
		ReadContext: datasourceGroupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Group. The ID for the Group must be obtained from the Admin API.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Group",
			},
			"is_active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if the Group is active",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the Group",
			},
			"security_policy_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Security Policy assigned to the Group.",
			},
		},
	}
}
