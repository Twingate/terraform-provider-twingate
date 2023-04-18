package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceUserRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	userID := resourceData.Get(attr.ID).(string)

	user, err := c.ReadUser(ctx, userID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.FirstName, user.FirstName); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.LastName, user.LastName); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.Email, user.Email); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.IsAdmin, user.IsAdmin()); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.Role, user.Role); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(userID)

	return nil
}

const userDescription = "Users in Twingate can be given access to Twingate Resources and may either be added manually or automatically synchronized with a 3rd party identity provider. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/users)."

func User() *schema.Resource {
	return &schema.Resource{
		Description: userDescription,
		ReadContext: datasourceUserRead,
		Schema: map[string]*schema.Schema{
			attr.ID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the User. The ID for the User can be obtained from the Admin API or the URL string in the Admin Console.",
			},
			// computed
			attr.FirstName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The first name of the User",
			},
			attr.LastName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last name of the User",
			},
			attr.Email: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address of the User",
			},
			attr.IsAdmin: {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether the User is an admin",
			},
			attr.Role: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Indicates the User's role. Either ADMIN, DEVOPS, SUPPORT, or MEMBER",
			},
		},
	}
}
