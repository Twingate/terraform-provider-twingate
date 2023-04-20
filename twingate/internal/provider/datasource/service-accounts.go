package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ServiceAccounts() *schema.Resource {
	return &schema.Resource{
		Description: "Service Accounts offer a way to provide programmatic, centrally-controlled, and consistent access controls.",
		ReadContext: readServiceAccounts,
		Schema: map[string]*schema.Schema{
			attr.Name: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter results by the name of the Service Account.",
			},
			attr.ServiceAccounts: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of Service Accounts",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						attr.ID: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the Service Account resource",
						},
						attr.Name: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the Service Account",
						},
						attr.ResourceIDs: {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of twingate_resource IDs that the Service Account is assigned to.",
						},
						attr.KeyIDs: {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of twingate_service_account_key IDs that are assigned to the Service Account.",
						},
					},
				},
			},
		},
	}
}

func readServiceAccounts(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	name := resourceData.Get(attr.Name).(string)

	services, err := client.ReadServiceAccounts(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.ServiceAccounts, convertServicesToTerraform(services)); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(terraformServicesDatasourceID(name))

	return nil
}

func terraformServicesDatasourceID(name string) string {
	id := "all-services"
	if name != "" {
		id = "service-by-name-" + name
	}

	return id
}
