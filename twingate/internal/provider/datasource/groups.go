package datasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func datasourceGroupsRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	filter := buildFilter(resourceData)

	groups, err := c.ReadGroups(ctx, filter)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("groups", convertGroupsToTerraform(groups)); err != nil {
		return diag.FromErr(err)
	}

	id := "all-groups"
	if filter.HasName() {
		id = "groups-by-name-" + *filter.Name
	}

	resourceData.SetId(id)

	return nil
}

func Groups() *schema.Resource {
	return &schema.Resource{
		Description: "Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).",
		ReadContext: datasourceGroupsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Returns only Groups that exactly match this name.",
			},
			"is_active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Returns only Groups matching the specified state.",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  fmt.Sprintf("Returns only Groups of the specified type (valid: `%s`, `%s`, `%s`).", model.GroupTypeManual, model.GroupTypeSynced, model.GroupTypeSystem),
				ValidateFunc: validation.StringInSlice([]string{model.GroupTypeManual, model.GroupTypeSynced, model.GroupTypeSystem}, false),
			},
			"groups": {
				Type:        schema.TypeList,
				Optional:    true,
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
				},
			},
		},
	}
}

func buildFilter(resourceData *schema.ResourceData) *model.GroupsFilter {
	groupName, hasName := resourceData.GetOk("name")
	groupType, hasType := resourceData.GetOk("type")

	// GetOk does not provide correct value for exists flag (second output value)
	groupIsActive, hasIsActive := resourceData.GetOkExists("is_active") //nolint

	if !hasName && !hasType && !hasIsActive {
		return nil
	}

	filter := &model.GroupsFilter{}

	if hasName {
		val := groupName.(string)
		filter.Name = &val
	}

	if hasType {
		val := groupType.(string)
		filter.Type = &val
	}

	if hasIsActive {
		val := groupIsActive.(bool)
		filter.IsActive = &val
	}

	return filter
}
