package resource

import (
	"context"
	"errors"
	"log"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
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
			attr.Name: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the group",
			},
			// optional
			attr.IsAuthoritative: {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Determines whether User assignments to this Group will override any existing assignments. Default is `true`. If set to `false`, assignments made outside of Terraform will be ignored.",
			},
			attr.UserIDs: {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of User IDs that have permission to access the Group.",
			},
			// computed
			attr.ID: {
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

	group, err := c.CreateGroup(ctx, convertGroup(resourceData))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Group %s created with id %v", group.Name, group.ID)

	return resourceGroupReadHelper(resourceData, group, nil)
}

func groupUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	group := convertGroup(resourceData)

	if err := deleteGroupUserIDs(ctx, resourceData, group, client); err != nil {
		return diag.FromErr(err)
	}

	group, err := client.UpdateGroup(ctx, group)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Updated group id %v", group.ID)

	return resourceGroupReadHelper(resourceData, group, err)
}

func deleteGroupUserIDs(ctx context.Context, resourceData *schema.ResourceData, group *model.Group, client *client.Client) error {
	userIDs := getGroupUserIDsToDelete(ctx, resourceData, group.Users, group, client)

	return client.DeleteGroupUsers(ctx, group.ID, userIDs) //nolint
}

func getGroupUserIDsToDelete(ctx context.Context, resourceData *schema.ResourceData, currentIDs []string, group *model.Group, client *client.Client) []string {
	oldIDs := getOldGroupUserIDs(ctx, resourceData, group, client)
	if len(oldIDs) == 0 {
		return nil
	}

	return setDifference(oldIDs, currentIDs)
}

func getOldGroupUserIDs(ctx context.Context, resourceData *schema.ResourceData, group *model.Group, client *client.Client) []string {
	if group.IsAuthoritative {
		result, err := client.ReadGroup(ctx, group.ID)
		if err != nil {
			return nil
		}

		return result.Users
	}

	old, _ := resourceData.GetChange(attr.UserIDs)

	return convertIDs(old)
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
	if group != nil {
		group.IsAuthoritative = convertAuthoritativeFlag(resourceData)
	}

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

	if !group.IsAuthoritative {
		group.Users = setIntersection(convertUsers(resourceData), group.Users)
	}

	resourceData.SetId(group.ID)

	if err := resourceData.Set(attr.Name, group.Name); err != nil {
		return ErrAttributeSet(err, attr.Name)
	}

	if _, exists := resourceData.GetOk(attr.UserIDs); exists {
		if err := resourceData.Set(attr.UserIDs, group.Users); err != nil {
			return ErrAttributeSet(err, attr.UserIDs)
		}
	}

	if err := resourceData.Set(attr.IsAuthoritative, group.IsAuthoritative); err != nil {
		return ErrAttributeSet(err, attr.IsAuthoritative)
	}

	return nil
}

func convertGroup(data *schema.ResourceData) *model.Group {
	return &model.Group{
		ID:              data.Id(),
		Name:            data.Get(attr.Name).(string),
		Users:           convertUsers(data),
		IsAuthoritative: convertAuthoritativeFlag(data),
	}
}
