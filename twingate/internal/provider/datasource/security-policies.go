package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SecurityPolicies() *schema.Resource {
	return &schema.Resource{
		Description: "Security Policies are defined in the Twingate Admin Console and determine user and device authentication requirements for Resources.",
		ReadContext: readSecurityPolicies,
		Schema: map[string]*schema.Schema{
			attr.SecurityPolicies: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						attr.ID: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Return a matching Security Policy by its ID. The ID for the Security Policy can be obtained from the Admin API or the URL string in the Admin Console.",
						},
						attr.Name: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Return a Security Policy that exactly matches this name.",
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

	if err := resourceData.Set(attr.SecurityPolicies, data); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId("security-policies-all")

	return nil
}
