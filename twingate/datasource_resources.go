package twingate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceResourcesRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	resourceName := resourceData.Get("name").(string)
	resources, err := client.readResourcesByName(ctx, resourceName)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("resources", convertResourcesToTerraform(resources)); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId("query resources by name: " + resourceName)

	return diags
}

func convertResourcesToTerraform(resources []*Resource) []interface{} {
	out := make([]interface{}, 0, len(resources))

	for _, res := range resources {
		rawData := convertResourceToTerraform(res)
		if rawData == nil {
			continue
		}

		out = append(out, rawData)
	}

	return out
}

func convertResourceToTerraform(resource *Resource) interface{} {
	if resource == nil {
		return nil
	}

	return map[string]interface{}{
		"id":                resource.ID.(string),
		"name":              string(resource.Name),
		"address":           string(resource.Address),
		"remote_network_id": resource.RemoteNetworkID.(string),
		"protocols":         convertProtocolsToTerraform(resource.Protocols),
	}
}

func datasourceResources() *schema.Resource { //nolint:funlen
	portsResource := schema.Resource{
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Whether to allow or deny all ports, or restrict protocol access within certain port ranges: Can be `%s` (only listed ports are allowed), `%s`, or `%s`", policyRestricted, policyAllowAll, policyDenyAll),
			},
			"ports": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of port ranges between 1 and 65535 inclusive, in the format `100-200` for a range, or `8080` for a single port",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}

	return &schema.Resource{
		Description: "Resources in Twingate represent servers on the private network that clients can connect to. Resources can be defined by IP, CIDR range, FQDN, or DNS zone. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).",
		ReadContext: datasourceResourcesRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Resource",
			},
			// computed
			"resources": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of Resources",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of the Resource",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the Resource",
						},
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Resource's IP/CIDR or FQDN/DNS zone",
						},
						"remote_network_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Remote Network ID where the Resource lives",
						},
						"protocols": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no restriction, and all protocols and ports are allowed.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allow_icmp": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether to allow ICMP (ping) traffic",
									},
									"tcp": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &portsResource,
									},
									"udp": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &portsResource,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
