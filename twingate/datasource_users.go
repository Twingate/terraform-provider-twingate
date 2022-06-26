package twingate

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceUserRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	users, err := client.readUsers(ctx)

	if err != nil {
		return diag.FromErr(err)
	}
	data := convertUsersToTerraform(users)
	log.Println("data", data)

	if err := resourceData.Set("users", data); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId("123")

	return diags
}

func convertUsersToTerraform(users map[string]*User) map[string]interface{} {
	//out := make(map[string][]interface{})
	//
	//for key, user := range users {
	//	out[key] = []interface{}{
	//		map[string]interface{}{
	//			"id": []interface{}{user.ID},
	//			//"first_name": user.FirstName,
	//			//"last_name":  user.LastName,
	//			//"email":      user.Email,
	//			//"is_admin":   user.IsAdmin,
	//		},
	//	}
	//}
	//
	//return out

	out := make(map[string]interface{})

	for key, user := range users {
		out[key] = user.ID
	}

	return out
}

func datasourceUsers() *schema.Resource {
	return &schema.Resource{
		Description: "Twingate users. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/users).",
		ReadContext: datasourceUserRead,
		Schema: map[string]*schema.Schema{
			"users": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					Computed: true,
					//MaxItems:    1,

					//Elem: &schema.Resource{
					//	Schema: map[string]*schema.Schema{
					//		"id": {
					//			Type:        schema.TypeString,
					//			Computed:    true,
					//			Description: "The ID of the User",
					//		},
					//		//"first_name": {
					//		//	Type:        schema.TypeString,
					//		//	Computed:    true,
					//		//	Description: "The first name of the User",
					//		//},
					//		//"last_name": {
					//		//	Type:        schema.TypeString,
					//		//	Computed:    true,
					//		//	Description: "The last name of the User",
					//		//},
					//		//"email": {
					//		//	Type:        schema.TypeString,
					//		//	Computed:    true,
					//		//	Description: "The email of the User",
					//		//},
					//		//"is_admin": {
					//		//	Type:        schema.TypeBool,
					//		//	Computed:    true,
					//		//	Description: "Indicates if the User admin or not",
					//		//},
					//	},
					//},
				},
			},
		},
	}
}
