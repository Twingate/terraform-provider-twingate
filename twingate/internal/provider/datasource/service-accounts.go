package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	fieldID              = "id"
	fieldName            = "name"
	fieldServiceAccounts = "service_accounts"
	fieldResourceIDs     = "resource_ids"
	fieldKeyIDs          = "key_ids"
)

func ServiceAccounts() *schema.Resource {
	return &schema.Resource{
		Description: "Service Accounts offer a way to provide programmatic, centrally-controlled, and consistent access controls.",
		ReadContext: readServiceAccounts,
		Schema: map[string]*schema.Schema{
			fieldName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter results by the name of the Service Account.",
			},
			fieldServiceAccounts: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of Service Accounts",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						fieldID: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the Service Account resource",
						},
						fieldName: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the Service Account",
						},
						fieldResourceIDs: {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of twingate_resource IDs that the Service Account is assigned to.",
						},
						fieldKeyIDs: {
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

	name := resourceData.Get(fieldName).(string)

	services, err := client.ReadServiceAccounts(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(fieldServiceAccounts, convertServicesToTerraform(services)); err != nil {
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
