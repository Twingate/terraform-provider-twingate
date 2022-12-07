package datasource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	fieldID        = "id"
	fieldName      = "name"
	fieldServices  = "service_accounts"
	fieldResources = "resource_ids"
	fieldKeys      = "key_ids"
)

func Services() *schema.Resource {
	return &schema.Resource{
		Description: "Service Accounts offer a way to provide programmatic, centrally-controlled, and consistent access controls.",
		ReadContext: readServices,
		Schema: map[string]*schema.Schema{
			fieldName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter results by the name of the Service Account.",
			},
			fieldServices: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of Services",
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
						fieldResources: {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of twingate_resource IDs that the Service Account is assigned to.",
						},
						fieldKeys: {
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

func readServices(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	name := resourceData.Get(fieldName).(string)

	services, err := client.ReadServices(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(fieldServices, convertServicesToTerraform(services)); err != nil {
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
