package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceResourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	resourceID := resourceData.Get("id").(string)

	resource, err := c.ReadResource(ctx, resourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("name", resource.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("address", resource.Address); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("remote_network_id", resource.RemoteNetworkID); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("protocols", provider.ConvertProtocolsToTerraform(resource.Protocols)); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(resourceID)

	return nil
}

func Resource() *schema.Resource { //nolint:funlen
	portsResource := schema.Resource{
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Whether to allow or deny all ports, or restrict protocol access within certain port ranges: Can be `%s` (only listed ports are allowed), `%s`, or `%s`", model.PolicyRestricted, model.PolicyAllowAll, model.PolicyDenyAll),
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
		Description: "Resources in Twingate represent any network destination address that you wish to provide private access to for users authorized via the Twingate Client application. Resources can be defined by either IP or DNS address, and all private DNS addresses will be automatically resolved with no client configuration changes. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).",
		ReadContext: datasourceResourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Resource. The ID for the Resource must be obtained from the Admin API.",
			},
			// computed
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Resource",
			},
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Resource's address, which may be an IP address, CIDR range, or DNS address",
			},
			"remote_network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Remote Network ID that the Resource is associated with. Resources may only be associated with a single Remote Network.",
			},
			"protocols": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "By default (when this argument is not defined) no restriction is applied, and all protocols and ports are allowed.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_icmp": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to allow ICMP (ping) traffic",
						},
						"tcp": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &portsResource,
						},
						"udp": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &portsResource,
						},
					},
				},
			},
		},
	}
}
