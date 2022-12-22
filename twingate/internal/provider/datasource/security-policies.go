package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const fieldSecurityPolicies = "security_policies"

func SecurityPolicies() *schema.Resource {
	return &schema.Resource{
		Description: "A Security Policy defined in Twingate for your Network or for individual Resources on your Network.",
		ReadContext: readSecurityPolicies,
		Schema: map[string]*schema.Schema{
			fieldSecurityPolicies: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						fieldID: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Find a Security Policy by id.",
						},
						fieldName: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Find a Security Policy by name.",
						},
					},
				},
			},
		},
	}
}

func readSecurityPolicies(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	securityPolicies, err := client.ReadSecurityPolicies(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	data := convertSecurityPoliciesToTerraform(securityPolicies)

	if err := resourceData.Set(fieldSecurityPolicies, data); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId("security-policies-all")

	return nil
}
