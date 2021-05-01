package twingate

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceResource() *schema.Resource { //nolint:funlen
	portsResource := schema.Resource{
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"RESTRICTED", "ALLOW_ALL"}, false),
				Description:  "Whether to allow all ports or restrict protocol access within certain port ranges.",
			},
			"ports": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "List of port ranges 1 and 65535 inclusively, in the format '100-200' for a range , or '8080' a single port ",
				},
			},
		},
	}

	return &schema.Resource{
		CreateContext: resourceResourceCreate,
		UpdateContext: resourceResourceUpdate,
		ReadContext:   resourceResourceRead,
		DeleteContext: resourceResourceDelete,

		Schema: map[string]*schema.Schema{
			// required
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "name of the resource",
			},
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The resource's IP/FQDN",
			},
			"groups": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of group IDs added to the resource",
			},
			"remote_network_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Remote Network ID to assign to the resource",
			},
			"protocols": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no restriction i.e. all protocols and ports are allowed.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_icmp": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to allow ICMP throughput",
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
				Description: "Autogenerated ID of the resource in encoded in base64",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}
func convertSlice(a []interface{}) []string {
	var res = make([]string, 0)
	for _, elem := range a {
		res = append(res, elem.(string))
	}

	return res
}

func extractProtocolsFromContext(p interface{}) *Protocols {
	protocols := &Protocols{}
	protocolsMap := p.(map[string]interface{})
	protocols.AllowIcmp = protocolsMap["allow_icmp"].(bool)
	u := protocolsMap["udp"].([]interface{})
	t := protocolsMap["tcp"].([]interface{})
	if len(u) > 0 {
		udp := u[0].(map[string]interface{})
		protocols.UDPPolicy = udp["policy"].(string)
		protocols.UDPPorts = convertSlice(udp["ports"].([]interface{}))
	}
	if len(t) > 0 {
		tcp := t[0].(map[string]interface{})
		protocols.TCPPolicy = tcp["policy"].(string)
		protocols.TCPPorts = convertSlice(tcp["ports"].([]interface{}))
	}

	return protocols
}
func newEmptyProtocols() *Protocols {
	return &Protocols{
		true, "ALLOW_ALL", []string{}, "ALLOW_ALL", []string{},
	}
}

func extractResource(d *schema.ResourceData) *Resource {
	resource := &Resource{
		Name:            d.Get("name").(string),
		RemoteNetworkId: d.Get("remote_network_id").(string),
		Address:         d.Get("address").(string),
		Groups:          convertSlice(d.Get("groups").([]interface{})),
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
	d.SetId(resource.Id)

	log.Printf("[INFO] Created resource %s", resource.Name)

	return resourceResourceRead(ctx, d, m)
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	if d.HasChanges("protocols", "remote_network_id", "name", "address", "groups") {
		resource := extractResource(d)
		resource.Id = d.Id()
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

	resourceId := d.Id()

	err := client.deleteResource(resourceId)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] deleted resource id %s", d.Id())

	return diags
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	var diags diag.Diagnostics

	resourceId := d.Id()

	resource, err := client.readResource(resourceId)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", resource.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %w ", err))
	}
	if err := d.Set("remote_network_id", resource.RemoteNetworkId); err != nil {
		return diag.FromErr(fmt.Errorf("error setting remote network: %w ", err))
	}
	if err := d.Set("address", resource.Address); err != nil {
		return diag.FromErr(fmt.Errorf("error setting address: %w ", err))
	}
	if err := d.Set("groups", resource.Groups); err != nil {
		return diag.FromErr(fmt.Errorf("error setting groups: %w ", err))
	}
	if len(d.Get("protocols").([]interface{})) > 0 {
		protocols := flattenProtocols(resource.Protocols)
		if err := d.Set("protocols", protocols); err != nil {
			return diag.FromErr(fmt.Errorf("error setting protocols: %w ", err))
		}
	}

	return diags
}

func flattenProtocols(protocols *Protocols) []interface{} {
	if protocols != nil {
		p := make(map[string]interface{})

		p["allow_icmp"] = protocols.AllowIcmp
		p["tcp"] = flattenPorts(protocols.TCPPolicy, protocols.TCPPorts)
		p["udp"] = flattenPorts(protocols.UDPPolicy, protocols.UDPPorts)

		return []interface{}{p}
	}

	return make([]interface{}, 0)
}
func flattenPorts(policy string, ports []string) []interface{} {
	c := make(map[string]interface{})
	c["policy"] = policy
	c["ports"] = ports

	return []interface{}{c}
}
