package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceResourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	resourceID := resourceData.Get(attr.ID).(string)

	resource, err := c.ReadResource(ctx, resourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.Name, resource.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.Address, resource.Address); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.RemoteNetworkID, resource.RemoteNetworkID); err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.Protocols, resource.Protocols.ToTerraform()); err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(resourceID)

	return nil
}

func Resource() *schema.Resource { //nolint:funlen
	portsResource := schema.Resource{
		Schema: map[string]*schema.Schema{
			attr.Policy: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Whether to allow or deny all ports, or restrict protocol access within certain port ranges: Can be `%s` (only listed ports are allowed), `%s`, or `%s`", model.PolicyRestricted, model.PolicyAllowAll, model.PolicyDenyAll),
			},
			attr.Ports: {
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
			attr.ID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Resource. The ID for the Resource must be obtained from the Admin API.",
			},
			// computed
			attr.Name: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Resource",
			},
			attr.Address: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Resource's address, which may be an IP address, CIDR range, or DNS address",
			},
			attr.RemoteNetworkID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Remote Network ID that the Resource is associated with. Resources may only be associated with a single Remote Network.",
			},
			attr.Protocols: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "By default (when this argument is not defined) no restriction is applied, and all protocols and ports are allowed.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						attr.AllowIcmp: {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to allow ICMP (ping) traffic",
						},
						attr.TCP: {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &portsResource,
						},
						attr.UDP: {
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
