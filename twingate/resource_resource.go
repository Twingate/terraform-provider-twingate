package twingate

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hasura/go-graphql-client"
)

func resourceResource() *schema.Resource { //nolint:funlen
	portsResource := schema.Resource{
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"RESTRICTED", "ALLOW_ALL"}, false),
				Description:  "Whether to allow all ports or restrict protocol access within certain port ranges: Can be `RESTRICTED` (only listed ports are allowed) or `ALLOW_ALL`",
			},
			"ports": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of port ranges between 1 and 65535 inclusive, in the format `100-200` for a range, or `8080` for a single port",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}

	return &schema.Resource{
		Description:   "Resources in Twingate represent servers on the private network that clients can connect to. Resources can be defined by IP, CIDR range, FQDN, or DNS zone. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).",
		CreateContext: resourceResourceCreate,
		UpdateContext: resourceResourceUpdate,
		ReadContext:   resourceResourceRead,
		DeleteContext: resourceResourceDelete,

		Schema: map[string]*schema.Schema{
			// required
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Resource",
			},
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Resource's IP/CIDR or FQDN/DNS zone",
			},
			"group_ids": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of Group IDs that have permission to access the Resource, cannot be generated by Terraform and must be retrieved from the Twingate Admin Console or API",
			},
			"remote_network_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Remote Network ID where the Resource lives",
			},
			"protocols": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no restriction, and all protocols and ports are allowed.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_icmp": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to allow ICMP (ping) traffic",
						},
						"tcp": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem:     &portsResource,
						},
						"udp": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem:     &portsResource,
						},
					},
				},
			},
			// computed
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Autogenerated ID of the Resource, encoded in base64",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func convertGroupsGraphql(a []interface{}) []*graphql.ID {
	res := []*graphql.ID{}
	for _, elem := range a {
		id := graphql.ID(elem.(string))
		res = append(res, &id)
	}

	return res
}

func extractProtocolsFromContext(p interface{}) *ProtocolsInput {
	protocolsMap := p.(map[string]interface{})
	protocolsInput := newProcolsInput()
	protocolsInput.AllowIcmp = graphql.Boolean(protocolsMap["allow_icmp"].(bool))

	u := protocolsMap["udp"].([]interface{})
	if len(u) > 0 {
		udp := u[0].(map[string]interface{})
		protocolsInput.UDP.Policy = graphql.String(udp["policy"].(string))
		p, err := convertPortsGraphql(udp["ports"].([]interface{}))
		if err != nil {
			log.Printf("[INFO] Cannot convert udp ports %v", udp["ports"].([]interface{}))
			return nil
		}
		if len(p) > 0 {
			protocolsInput.UDP.Ports = p
		}
	}

	t := protocolsMap["tcp"].([]interface{})
	if len(t) > 0 {
		tcp := t[0].(map[string]interface{})
		protocolsInput.TCP.Policy = graphql.String(tcp["policy"].(string))
		p, err := convertPortsGraphql(tcp["ports"].([]interface{}))
		if err != nil {
			log.Printf("[INFO] Cannot convert tcp ports %v", tcp["ports"].([]interface{}))
			return nil
		}
		if len(p) > 0 {
			protocolsInput.TCP.Ports = p
		}
	}

	return protocolsInput
}

func extractResource(d *schema.ResourceData) *Resource {
	resource := &Resource{
		Name:            graphql.String(d.Get("name").(string)),
		RemoteNetworkID: graphql.ID(d.Get("remote_network_id").(string)),
		Address:         graphql.String(d.Get("address").(string)),
		GroupsIds:       convertGroupsGraphql(d.Get("group_ids").([]interface{})),
	}

	p := d.Get("protocols").([]interface{})

	if len(p) > 0 {
		resource.Protocols = extractProtocolsFromContext(p[0])
	} else {
		resource.Protocols = newEmptyProtocols()
	}

	return resource
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	resource := extractResource(d)
	err := client.createResource(resource)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resource.ID.(string))
	log.Printf("[INFO] Created resource %s", resource.Name)

	return resourceResourceRead(ctx, d, m)
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	if d.HasChanges("protocols", "remote_network_id", "name", "address", "group_ids") {
		resource := extractResource(d)
		resource.ID = d.Id()

		err := client.updateResource(resource)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceResourceRead(ctx, d, m)
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	resourceID := d.Id()

	err := client.deleteResource(resourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] deleted resource id %s", d.Id())

	return diags
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Client)
	resourceID := d.Id()

	resource, err := client.readResource(resourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", resource.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %w ", err))
	}

	if err := d.Set("remote_network_id", resource.RemoteNetworkID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting remote network: %w ", err))
	}

	if err := d.Set("address", resource.Address); err != nil {
		return diag.FromErr(fmt.Errorf("error setting address: %w ", err))
	}

	if err := d.Set("group_ids", resource.stringGroups()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting group_ids: %w ", err))
	}

	if len(d.Get("protocols").([]interface{})) > 0 {
		protocols := flattenProtocols(resource.Protocols)
		if err := d.Set("protocols", protocols); err != nil {
			return diag.FromErr(fmt.Errorf("error setting protocols: %w ", err))
		}
	}

	return diags
}

func flattenProtocols(protocols *ProtocolsInput) []interface{} {
	if protocols != nil {
		p := make(map[string]interface{})

		p["allow_icmp"] = protocols.AllowIcmp
		p["tcp"] = flattenPortsGraphql(protocols.TCP.Policy, protocols.TCP.Ports)
		p["udp"] = flattenPortsGraphql(protocols.UDP.Policy, protocols.UDP.Ports)

		return []interface{}{p}
	}

	return make([]interface{}, 0)
}

func flattenPortsGraphql(policy graphql.String, ports []*PortRangeInput) []interface{} {
	p := []string{}
	for _, port := range ports {
		p = append(p, strconv.Itoa(int(port.Start)))
		p = append(p, strconv.Itoa(int(port.End)))
	}

	c := make(map[string]interface{})

	c["policy"] = policy
	c["ports"] = p

	return []interface{}{c}
}
