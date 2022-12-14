package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceUsersRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)

	users, err := c.ReadUsers(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	data := convertUsersToTerraform(users)

	if err := resourceData.Set("users", data); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId("users-all")

	return nil
}

func Users() *schema.Resource {
	return &schema.Resource{
		Description: userDescription,
		ReadContext: datasourceUsersRead,
		Schema: map[string]*schema.Schema{
			"users": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the User",
						},
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
							Description: "Indicates the User's role. Either ADMIN, DEVOPS, SUPPORT, or MEMBER.",
						},
					},
				},
			},
		},
	}
}
