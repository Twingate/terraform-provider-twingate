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
	fieldServices  = "services"
	fieldResources = "resources"
	fieldKeys      = "keys"
)

func Services() *schema.Resource {
	return &schema.Resource{
		Description: "Services offer a way to provide programmatic, centrally-controlled, and consistent access controls. For more information, see Twingate's [documentation](https://www.twingate.com/docs/services).",
		ReadContext: readServices,
		Schema: map[string]*schema.Schema{
			fieldName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter results by the name of the service account.",
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
							Description: "ID of the service account resource",
						},
						fieldName: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the service account",
						},
						fieldResources: {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of twingate_resource IDs that the service account is assigned to.",
						},
						fieldKeys: {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of twingate_service_key IDs that are assigned to the service account.",
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
