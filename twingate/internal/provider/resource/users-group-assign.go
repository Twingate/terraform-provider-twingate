package resource

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	fieldUserIDs = "user_ids"
	fieldGroupID = "group_id"
)

func UsersGroupAssign() *schema.Resource {
	return &schema.Resource{
		Description:   "Any existing relationships are replaced with exactly what is set in this resource.",
		CreateContext: createUsersGroupAssign,
		ReadContext:   readUsersGroupAssign,
		DeleteContext: deleteUsersGroupAssign,
		UpdateContext: updateUsersGroupAssign,

		Schema: map[string]*schema.Schema{
			fieldGroupID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Group resource id obtained from the twingate_group or twingate_groups data sources.",
			},
			fieldUserIDs: {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "List of user ids obtained from the twingate_user or twingate_users data sources.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func createUsersGroupAssign(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return updateUsersGroupAssign(ctx, resourceData, meta)
}

func updateUsersGroupAssign(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	group, err := client.AssignGroupUsers(ctx,
		resourceData.Get(fieldGroupID).(string),
		readResourceDataIDs(resourceData, fieldUserIDs),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return usersGroupAssignReadHelper(resourceData, group, nil)
}

func readResourceDataIDs(resourceData *schema.ResourceData, fieldName string) []string {
	return utils.Map[interface{}, string](resourceData.Get(fieldName).(*schema.Set).List(), func(id interface{}) string {
		return id.(string)
	})
}

func deleteUsersGroupAssign(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	_, err := client.RemoveGroupUsers(ctx, resourceData.Id(), readResourceDataIDs(resourceData, fieldUserIDs))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func readUsersGroupAssign(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)
	group, err := client.ReadGroup(ctx, resourceData.Id())

	return usersGroupAssignReadHelper(resourceData, group, err)
}

func usersGroupAssignReadHelper(resourceData *schema.ResourceData, group *model.Group, err error) diag.Diagnostics {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			resourceData.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if err := resourceData.Set(fieldGroupID, group.ID); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(fieldUserIDs, group.UserIDs); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(group.ID)

	return nil
}
