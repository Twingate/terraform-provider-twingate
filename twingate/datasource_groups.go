package twingate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceGroupsRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	groupName := resourceData.Get("name").(string)
	groups, err := client.readGroupsByName(ctx, groupName)

	if err != nil {
		return diag.FromErr(err)
	}

	data := make([]interface{}, 0, len(groups))
	for _, group := range groups {
		data = append(data, map[string]interface{}{
			"id":   group.ID.(string),
			"name": string(group.Name),
		})
	}

	if err := resourceData.Set("groups", data); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(groupName)

	return diags
}

func datasourceGroups() *schema.Resource {
	return &schema.Resource{
		Description: "Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).",
		ReadContext: datasourceGroupsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Group",
			},
			"groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of Groups",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the Group",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the Group",
						},
					},
				},
			},
		},
	}
}
