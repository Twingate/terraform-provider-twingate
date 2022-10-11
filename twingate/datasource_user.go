package twingate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceUserRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	userID := resourceData.Get("id").(string)
	user, err := client.readUser(ctx, userID)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("first_name", user.FirstName); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("last_name", user.LastName); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("email", user.Email); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("is_admin", user.IsAdmin()); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("role", user.Role); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(userID)

	return diags
}

func datasourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Users in Twingate can be given access to Twingate Resources and may either be added manually or automatically synchronized with a 3rd party identity provider. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/users).",
		ReadContext: datasourceUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the User. The ID for the User must be obtained from the Admin API.",
			},
			// computed
			"first_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The first name of the User",
			},
			"last_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last name of the User",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address of the User",
			},
			"is_admin": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether the User is an admin",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Indicates the User's role. Either ADMIN, DEVOPS, SUPPORT, or MEMBER",
			},
		},
	}
}
