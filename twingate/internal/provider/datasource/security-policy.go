package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SecurityPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "A Security Policy defined in Twingate for your Network or for individual Resources on your Network.",
		ReadContext: readSecurityPolicy,
		Schema: map[string]*schema.Schema{
			fieldID: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Find a Security Policy by id.",
				ExactlyOneOf: []string{fieldName},
			},
			fieldName: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Find a Security Policy by name.",
				ExactlyOneOf: []string{fieldID},
			},
		},
	}
}

func readSecurityPolicy(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	securityPolicy, err := client.ReadSecurityPolicy(ctx, resourceData.Get(fieldID).(string), resourceData.Get(fieldName).(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(fieldName, securityPolicy.Name); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(securityPolicy.ID)

	return nil
}
