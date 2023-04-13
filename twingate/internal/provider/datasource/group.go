package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceGroupRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	groupID := resourceData.Get(attr.ID).(string)

	group, err := c.ReadGroup(ctx, groupID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.Name, group.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.Type, group.Type); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.IsActive, group.IsActive); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.SecurityPolicyID, group.SecurityPolicyID); err != nil {
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
			attr.ID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Group. The ID for the Group can be obtained from the Admin API or the URL string in the Admin Console.",
			},
			attr.Name: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Group",
			},
			attr.IsActive: {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if the Group is active",
			},
			attr.Type: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the Group",
			},
			attr.SecurityPolicyID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Security Policy assigned to the Group.",
			},
		},
	}
}
